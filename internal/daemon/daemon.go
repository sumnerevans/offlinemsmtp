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

	"github.com/fsnotify/fsnotify"
	"github.com/godbus/dbus/v5"
	"github.com/rs/zerolog"

	"github.com/sumnerevans/offlinemsmtp/internal/notify"
)

var subjectRe = regexp.MustCompile(`(?m)^Subject: ([^\r\n]+)`)

const sendTimeout = 90 * time.Second

type Config struct {
	RootDir          string
	ConfigFile       string
	MsmtpPath        string
	SendMailFile     string
	Silent           bool
	Interval         time.Duration
	DebounceInterval time.Duration
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
	d.notifier = notify.New(d.Silent, log.With().Str("component", "notifier").Logger())
	defer d.notifier.Close()

	if err := os.MkdirAll(d.RootDir, 0o755); err != nil {
		return fmt.Errorf("create outbox directory: %w", err)
	}

	d.notifier.Send("Listener started", fmt.Sprintf("Watching %s for new messages", d.RootDir), 5*time.Second, notify.UrgencyLow)

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

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create file watcher: %w", err)
	}
	defer watcher.Close()
	if err := watcher.Add(d.RootDir); err != nil {
		return fmt.Errorf("watch outbox directory: %w", err)
	}

	ticker := time.NewTicker(d.Interval)
	defer ticker.Stop()

	var lastFlush time.Time
	for {
		if d.DebounceInterval == 0 || time.Since(lastFlush) >= d.DebounceInterval {
			d.flushQueue(ctx)
			lastFlush = time.Now()
		} else {
			log.Debug().Dur("next_in", d.DebounceInterval-time.Since(lastFlush)).Msg("debouncing flush")
		}

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
			if !ok {
				continue
			}
			// NM_STATE_CONNECTED_SITE=60, NM_STATE_CONNECTED_GLOBAL=70
			if state < 60 {
				log.Debug().Uint32("nm_state", state).Msg("network not connected, skipping flush")
				continue
			}
			log.Info().Uint32("nm_state", state).Msg("network connected, flushing queue")
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if !event.Has(fsnotify.Create) && !event.Has(fsnotify.Write) && !event.Has(fsnotify.Rename) {
				log.Debug().Str("file", event.Name).Stringer("op", event.Op).Msg("ignoring event")
				continue
			}
			if strings.HasPrefix(filepath.Base(event.Name), ".tmp-") {
				log.Debug().Str("file", event.Name).Msg("ignoring tmp file event")
				continue
			}
			log.Info().Str("file", event.Name).Msg("new file detected")
		case <-ticker.C:
			log.Debug().Msg("ticker fired, flushing queue")
		}
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
	// NM_CONNECTIVITY_FULL = 4
	return ok && connectivity >= 4
}

func (d *Daemon) flushQueue(ctx context.Context) {
	log := zerolog.Ctx(ctx).With().Str("operation", "flush").Logger()
	log.Debug().Msg("flushing queue")
	start := time.Now()

	entries, err := os.ReadDir(d.RootDir)
	if err != nil {
		log.Err(err).Msg("cannot read outbox directory")
		return
	}
	entries = slices.DeleteFunc(entries, func(e os.DirEntry) bool {
		return e.IsDir() || strings.HasPrefix(e.Name(), ".tmp-")
	})

	if len(entries) == 0 {
		log.Info().Msg("no messages to send")
		return
	}
	if !d.isOnline() {
		log.Warn().Msg("not online, skipping flush")
		return
	}
	if d.SendMailFile != "" {
		if _, err := os.Stat(d.SendMailFile); err != nil {
			log.Debug().Str("send_mail_file", d.SendMailFile).Msg("sending email disabled because SendMailFile not present")
			d.notifier.Send("Sending email disabled", fmt.Sprintf("Skipping sending emails because %s does not exist", d.SendMailFile), 5*time.Second, notify.UrgencyLow)
			return
		}
	}

	// ReadDir returns entries sorted by name (= chronological for our
	// timestamp filenames). Reverse so newest messages are attempted first;
	// older ones are more likely to have already failed.
	slices.Reverse(entries)
	for _, e := range entries {
		path := filepath.Join(d.RootDir, e.Name())
		log := log.With().Str("file", path).Logger()

		if data, err := os.ReadFile(path); err != nil {
			log.Err(err).Msg("cannot read queued message")
			continue
		} else if msmtpArgs, message, err := parseQueueFile(data); err != nil {
			log.Err(err).Msg("malformed queue file")
			continue
		} else if !d.sendMessage(ctx, path, msmtpArgs, message) {
			break
		}
	}

	log.Info().TimeDiff("elapsed_ms", time.Now(), start).Int("entry_count", len(entries)).Msg("queue flush complete")
}

func (d *Daemon) sendMessage(ctx context.Context, path string, msmtpArgs string, message []byte) bool {
	subject := extractSubject(message)
	log := zerolog.Ctx(ctx).With().Str("subject", subject).Logger()

	sendingHandle := d.notifier.Send("Sending message...", subject, sendTimeout, notify.UrgencyCritical)

	if !d.isOnline() {
		d.notifier.Replace(sendingHandle, "Not connected, will send later", subject, 5*time.Second, notify.UrgencyLow)
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	cmdArgs := d.buildCmd(msmtpArgs)
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = bytes.NewReader(message)
	log.Debug().Any("command", cmdArgs).Msg("running command")
	if sendErr := cmd.Run(); sendErr != nil {
		log.Err(sendErr).Str("file", path).Msg("msmtp failed")
		d.notifier.Replace(sendingHandle,
			fmt.Sprintf("Failed to send '%s'", subject),
			sendErr.Error(),
			30*time.Second, notify.UrgencyNormal)
		return true
	}

	log.Info().Str("file", path).Msg("message sent, removing from queue")
	d.notifier.Replace(sendingHandle, "Message sent successfully", subject, 5*time.Second, notify.UrgencyLow)
	if err := os.Remove(path); err != nil {
		log.Err(err).Str("file", path).Msg("cannot remove sent message")
	}

	return true
}

func (d *Daemon) buildCmd(msmtpArgs string) []string {
	cmd := []string{d.MsmtpPath, "--debug", "-C", d.ConfigFile}
	if msmtpArgs != "" {
		cmd = append(cmd, strings.Fields(msmtpArgs)...)
	}
	return cmd
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
