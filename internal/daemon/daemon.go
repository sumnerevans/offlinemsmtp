package daemon

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	inotify "github.com/rjeczalik/notify"
	"github.com/rs/zerolog"

	"github.com/sumnerevans/offlinemsmtp/internal/notify"
)

var (
	hostRe    = regexp.MustCompile(`^host = (.+)`)
	portRe    = regexp.MustCompile(`^port = (.+)`)
	subjectRe = regexp.MustCompile(`^Subject: (.+)`)
)

type Config struct {
	RootDir      string
	ConfigFile   string
	SendMailFile string
	Silent       bool
	Interval     time.Duration
}

type Daemon struct {
	Config
	queue    []string
	notifier *notify.Notifier
}

func New(cfg Config) *Daemon {
	return &Daemon{Config: cfg}
}

func (d *Daemon) Run(ctx context.Context) error {
	log := zerolog.Ctx(ctx)
	d.notifier = notify.New(d.Silent, *log)
	defer d.notifier.Close()

	d.notifier.Send("offlinemsmtp daemon started", 5*time.Second, notify.UrgencyLow)

	if err := os.MkdirAll(d.RootDir, 0o755); err != nil {
		return fmt.Errorf("create outbox directory: %w", err)
	}

	entries, err := os.ReadDir(d.RootDir)
	if err != nil {
		return fmt.Errorf("read outbox directory: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".tmp-") {
			d.queue = append(d.queue, filepath.Join(d.RootDir, e.Name()))
		}
	}

	// Watch NetworkManager for connectivity changes so we can flush
	// immediately when the system comes online.
	var nmSignals chan *dbus.Signal
	sysBus, err := dbus.ConnectSystemBus()
	if err != nil {
		log.Warn().Err(err).Msg("cannot connect to system D-Bus; network state changes will not trigger flush")
	} else {
		defer sysBus.Close()
		if err := sysBus.AddMatchSignal(
			dbus.WithMatchInterface("org.freedesktop.NetworkManager"),
			dbus.WithMatchMember("StateChanged"),
		); err != nil {
			log.Warn().Err(err).Msg("cannot watch NetworkManager signals; network state changes will not trigger flush")
		} else {
			nmSignals = make(chan *dbus.Signal, 16)
			sysBus.Signal(nmSignals)
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
			d.queue = append(d.queue, ei.Path())
		case <-ticker.C:
		}
		if len(d.queue) > 0 {
			d.flushQueue(ctx)
		}
	}
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
		d.notifier.Send("Sending email disabled", 5*time.Second, notify.UrgencyLow)
		return
	}

	pending := d.queue
	d.queue = nil

	var failed []string
	for _, path := range pending {
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			log.Error().Err(err).Str("file", path).Msg("cannot read queued message")
			failed = append(failed, path)
			continue
		}

		msmtpArgs, message, err := parseQueueFile(data)
		if err != nil {
			log.Error().Err(err).Str("file", path).Msg("malformed queue file")
			failed = append(failed, path)
			continue
		}

		if !d.canSend(ctx, msmtpArgs, message) {
			failed = append(failed, path)
			continue
		}

		sendingHandle := d.notifier.Send(
			fmt.Sprintf("Sending %s...", filepath.Base(path)),
			10*time.Minute,
			notify.UrgencyLow,
		)
		sendErr := d.send(ctx, msmtpArgs, message)
		sendingHandle.Close()

		if sendErr != nil {
			log.Error().Err(sendErr).Str("file", path).Msg("msmtp failed")
			d.notifier.Send(
				fmt.Sprintf("Message did not send. Putting back in queue.\nError: %v", sendErr),
				30*time.Second,
				notify.UrgencyCritical,
			)
			failed = append(failed, path)
			continue
		}

		log.Info().Str("file", path).Msg("message sent, removing from queue")
		d.notifier.Send("Message sent successfully. Removing from queue.", 5*time.Second, notify.UrgencyLow)
		if err := os.Remove(path); err != nil {
			log.Error().Err(err).Str("file", path).Msg("cannot remove sent message")
		}
	}

	d.queue = append(failed, d.queue...)
}

func (d *Daemon) buildCmd(msmtpArgs string, extra ...string) []string {
	cmd := []string{"/usr/bin/env", "msmtp", "--debug", "-C", d.ConfigFile}
	cmd = append(cmd, extra...)
	if msmtpArgs != "" {
		cmd = append(cmd, strings.Fields(msmtpArgs)...)
	}
	return cmd
}

func (d *Daemon) canSend(ctx context.Context, msmtpArgs string, message []byte) bool {
	log := zerolog.Ctx(ctx)
	cmdArgs := d.buildCmd(msmtpArgs, "-P")

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = bytes.NewReader(message)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	_ = cmd.Run()

	var host string
	var port int
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if m := hostRe.FindStringSubmatch(line); m != nil {
			host = strings.TrimSpace(m[1])
		} else if m := portRe.FindStringSubmatch(line); m != nil {
			p, err := strconv.Atoi(strings.TrimSpace(m[1]))
			if err == nil {
				port = p
			}
		}
		if host != "" && port != 0 {
			break
		}
	}

	if host == "" || port == 0 {
		log.Warn().Msg("could not parse host/port from msmtp pretend output")
		return false
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 2*time.Second)
	if err != nil {
		subject := extractSubject(message)
		d.notifier.Send(
			fmt.Sprintf("Cannot connect to %s:%d to send message with subject: %q", host, port, subject),
			5*time.Second,
			notify.UrgencyLow,
		)
		return false
	}
	conn.Close()
	return true
}

func (d *Daemon) send(ctx context.Context, msmtpArgs string, message []byte) error {
	cmdArgs := d.buildCmd(msmtpArgs)
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = bytes.NewReader(message)
	return cmd.Run()
}

func parseQueueFile(data []byte) (msmtpArgs string, message []byte, err error) {
	nl := bytes.IndexByte(data, '\n')
	if nl < 0 {
		return "", nil, fmt.Errorf("missing newline")
	}
	return strings.TrimSpace(string(data[:nl])), data[nl+1:], nil
}

func extractSubject(message []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(message))
	for scanner.Scan() {
		if m := subjectRe.FindStringSubmatch(scanner.Text()); m != nil {
			return m[1]
		}
	}
	return "<no subject>"
}
