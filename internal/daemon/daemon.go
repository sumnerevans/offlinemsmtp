package daemon

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	inotify "github.com/rjeczalik/notify"
	"github.com/rs/zerolog"

	"github.com/sumnerevans/offlinemsmtp/internal/notify"
)

var subjectRe = regexp.MustCompile(`(?m)^Subject: ([^\r\n]+)`)

const sendTimeout = 90 * time.Second

type Config struct {
	RootDir      string
	ConfigFile   string
	MsmtpPath    string
	SendMailFile string
	Silent       bool
	Interval     time.Duration
}

type Daemon struct {
	Config
	notifier *notify.Notifier
	sysBus   *dbus.Conn
}

func New(cfg Config) *Daemon {
	return &Daemon{Config: cfg}
}

func (d *Daemon) Run(ctx context.Context) error {
	log := zerolog.Ctx(ctx)
	d.notifier = notify.New(d.Silent, *log)
	defer d.notifier.Close()

	d.notifier.Send("offlinemsmtp", "daemon started", 5*time.Second, notify.UrgencyLow)

	if err := os.MkdirAll(d.RootDir, 0o755); err != nil {
		return fmt.Errorf("create outbox directory: %w", err)
	}

	// Watch NetworkManager for connectivity changes so we can flush
	// immediately when the system comes online.
	var nmSignals chan *dbus.Signal
	var err error
	d.sysBus, err = dbus.ConnectSystemBus()
	if err != nil {
		log.Warn().Err(err).Msg("cannot connect to system D-Bus; network state changes will not trigger flush")
	} else {
		defer d.sysBus.Close()
		if err := d.sysBus.AddMatchSignal(
			dbus.WithMatchInterface("org.freedesktop.NetworkManager"),
			dbus.WithMatchMember("StateChanged"),
		); err != nil {
			log.Warn().Err(err).Msg("cannot watch NetworkManager signals; network state changes will not trigger flush")
		} else {
			nmSignals = make(chan *dbus.Signal, 16)
			d.sysBus.Signal(nmSignals)
		}
	}

	events := make(chan inotify.EventInfo, 16)
	if err := inotify.Watch(d.RootDir, events, inotify.InMovedTo); err != nil {
		return fmt.Errorf("watch outbox directory: %w", err)
	}
	defer inotify.Stop(events)

	ticker := time.NewTicker(d.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case sig, ok := <-nmSignals:
			if !ok {
				nmSignals = nil
				continue
			}
			if len(sig.Body) == 0 {
				continue
			}
			state, ok := sig.Body[0].(uint32)
			// NM_STATE_CONNECTED_SITE=60, NM_STATE_CONNECTED_GLOBAL=70
			if !ok || state < 60 {
				continue
			}
			log.Info().Uint32("nm_state", state).Msg("network connected, flushing queue")
		case ei := <-events:
			log.Info().Str("file", ei.Path()).Msg("new message detected")
			d.notifier.Send("offlinemsmtp", fmt.Sprintf("new message queued: %s", filepath.Base(ei.Path())), 5*time.Second, notify.UrgencyLow)
		case <-ticker.C:
		}
		d.flushQueue(ctx)
	}
}

func (d *Daemon) isOnline() bool {
	if d.sysBus == nil {
		return true
	}
	obj := d.sysBus.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")
	v, err := obj.GetProperty("org.freedesktop.NetworkManager.Connectivity")
	if err != nil {
		return true
	}
	connectivity, ok := v.Value().(uint32)
	// NM_CONNECTIVITY_LIMITED=3, NM_CONNECTIVITY_FULL=4
	return ok && connectivity >= 3
}

func (d *Daemon) sendEnabled() bool {
	if d.SendMailFile == "" {
		return true
	}
	_, err := os.Stat(d.SendMailFile)
	return err == nil
}

func (d *Daemon) flushQueue(ctx context.Context) {
	log := zerolog.Ctx(ctx)

	if !d.sendEnabled() {
		d.notifier.Send("offlinemsmtp", "sending email disabled", 5*time.Second, notify.UrgencyLow)
		return
	}

	entries, err := os.ReadDir(d.RootDir)
	if err != nil {
		log.Err(err).Msg("cannot read outbox directory")
		return
	}

	// ReadDir returns entries sorted by name (= chronological for our
	// timestamp filenames). Reverse so newest messages are attempted
	// first; older ones are more likely to have already failed.
	slices.Reverse(entries)
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".tmp-") {
			continue
		}
		path := filepath.Join(d.RootDir, e.Name())

		data, err := os.ReadFile(path)
		if err != nil {
			log.Err(err).Str("file", path).Msg("cannot read queued message")
			continue
		}

		msmtpArgs, message, err := parseQueueFile(data)
		if err != nil {
			log.Err(err).Str("file", path).Msg("malformed queue file")
			continue
		}

		d.sendMessage(ctx, path, msmtpArgs, message)
	}
}

func (d *Daemon) sendMessage(ctx context.Context, path string, msmtpArgs string, message []byte) {
	log := zerolog.Ctx(ctx)
	subject := extractSubject(message)

	sendingHandle := d.notifier.Send(subject, "sending...", sendTimeout, notify.UrgencyLow)

	if !d.isOnline() {
		d.notifier.Replace(sendingHandle, subject, "not connected, will retry later", 5*time.Second, notify.UrgencyLow)
		return
	}

	sendCtx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	sendErr := d.send(sendCtx, msmtpArgs, message)
	if sendErr != nil {
		log.Err(sendErr).Str("file", path).Msg("msmtp failed")
		d.notifier.Replace(sendingHandle, subject,
			fmt.Sprintf("failed to send, will retry later: %v", sendErr),
			30*time.Second, notify.UrgencyCritical)
		return
	}

	log.Info().Str("file", path).Msg("message sent, removing from queue")
	d.notifier.Replace(sendingHandle, subject, "sent successfully", 5*time.Second, notify.UrgencyLow)
	if err := os.Remove(path); err != nil {
		log.Err(err).Str("file", path).Msg("cannot remove sent message")
	}
}

func (d *Daemon) buildCmd(msmtpArgs string, extra ...string) []string {
	cmd := []string{d.MsmtpPath, "--debug", "-C", d.ConfigFile}
	cmd = append(cmd, extra...)
	if msmtpArgs != "" {
		cmd = append(cmd, strings.Fields(msmtpArgs)...)
	}
	return cmd
}

func (d *Daemon) send(ctx context.Context, msmtpArgs string, message []byte) error {
	cmdArgs := d.buildCmd(msmtpArgs)
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = bytes.NewReader(message)
	return cmd.Run()
}

func parseQueueFile(data []byte) (msmtpArgs string, message []byte, err error) {
	before, after, ok := bytes.Cut(data, []byte("\n"))
	if !ok {
		return "", nil, fmt.Errorf("missing newline")
	}
	return strings.TrimSpace(string(before)), after, nil
}

func extractSubject(message []byte) string {
	if m := subjectRe.FindSubmatch(message); m != nil {
		return string(m[1])
	}
	return "<no subject>"
}
