package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/rs/zerolog"

	"github.com/sumnerevans/offlinemsmtp/internal/daemon"
)

type args struct {
	OutboxDir        string   `arg:"-o,--outbox-directory" help:"outbox directory" default:"$HOME/.offlinemsmtp-outbox"`
	Daemon           bool     `arg:"-d,--daemon" help:"run the offlinemsmtp daemon"`
	Silent           bool     `arg:"-s,--silent" help:"disable all logging and notifications"`
	Interval         int      `arg:"-i,--interval" help:"flush interval in seconds" default:"60"`
	DebounceInterval int      `arg:"--debounce-interval" help:"minimum seconds between flushes" default:"10"`
	ConfigFile       string   `arg:"-C,--file" help:"msmtp configuration file" default:"$HOME/.msmtprc"`
	MsmtpPath        string   `arg:"--msmtp-path" help:"path to msmtp binary" default:"msmtp"`
	SendMailFile     string   `arg:"--send-mail-file" help:"only send mail if this file exists"`
	LogFile          string   `arg:"-l,--logfile" help:"file to write logs to"`
	LogLevel         string   `arg:"-m,--loglevel" help:"minimum log level (trace/debug/info/warn/error)" default:"info"`
	MsmtpArgs        []string `arg:"positional" help:"arguments forwarded to msmtp"`
}

func (args) Description() string {
	return "offlinemsmtp -- offline wrapper for msmtp"
}

func main() {
	var a args
	arg.MustParse(&a)

	a.OutboxDir = os.ExpandEnv(a.OutboxDir)
	a.ConfigFile = os.ExpandEnv(a.ConfigFile)

	level, err := zerolog.ParseLevel(strings.ToLower(a.LogLevel))
	if err != nil {
		level = zerolog.WarnLevel
	}

	var logWriter io.Writer = zerolog.ConsoleWriter{Out: os.Stderr}
	if a.LogFile != "" {
		f, err := os.OpenFile(a.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot open log file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		logWriter = io.MultiWriter(logWriter, f)
	}
	if a.Silent {
		logWriter = io.Discard
	}

	logger := zerolog.New(logWriter).Level(level).With().Timestamp().Logger()
	ctx := logger.WithContext(context.Background())

	if a.Daemon {
		logger.Info().Msg("starting offlinemsmtp daemon")
		d := daemon.New(daemon.Config{
			RootDir:          a.OutboxDir,
			ConfigFile:       a.ConfigFile,
			MsmtpPath:        a.MsmtpPath,
			SendMailFile:     a.SendMailFile,
			Silent:           a.Silent,
			Interval:         time.Duration(a.Interval) * time.Second,
			DebounceInterval: time.Duration(a.DebounceInterval) * time.Second,
		})
		if err := d.Run(ctx); err != nil {
			logger.Fatal().Err(err).Msg("daemon error")
		}
		return
	}

	// Queue mode: write to a temp file in the outbox dir, then rename
	// atomically so the daemon never sees a partially-written file.
	if err := os.MkdirAll(a.OutboxDir, 0o755); err != nil {
		logger.Fatal().Err(err).Msg("cannot create outbox directory")
	}
	tmp, err := os.CreateTemp(a.OutboxDir, ".tmp-*")
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot create temp file in outbox")
	}
	tmpName := tmp.Name()

	fmt.Fprintln(tmp, strings.Join(a.MsmtpArgs, " "))
	if _, err := io.Copy(tmp, bufio.NewReader(os.Stdin)); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		logger.Fatal().Err(err).Msg("cannot write email to outbox")
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		logger.Fatal().Err(err).Msg("cannot close temp file")
	}

	filename := filepath.Join(a.OutboxDir, time.Now().Format("2006-01-02_15-04-05"))
	if err := os.Rename(tmpName, filename); err != nil {
		os.Remove(tmpName)
		logger.Fatal().Err(err).Msg("cannot move email into outbox")
	}
	logger.Info().Str("file", filename).Msg("queued email")
}
