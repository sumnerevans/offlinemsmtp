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
	OutboxDir    string `arg:"-o,--outbox-directory" help:"outbox directory" default:"$HOME/.offlinemsmtp-outbox"`
	Daemon       bool   `arg:"-d,--daemon" help:"run the offlinemsmtp daemon"`
	Silent       bool   `arg:"-s,--silent" help:"disable all logging and notifications"`
	Interval     int    `arg:"-i,--interval" help:"flush interval in seconds" default:"60"`
	ConfigFile   string `arg:"-C,--file" help:"msmtp configuration file" default:"$HOME/.msmtprc"`
	SendMailFile string `arg:"--send-mail-file" help:"only send mail if this file exists"`
	LogFile      string `arg:"-l,--logfile" help:"file to write logs to"`
	LogLevel     string `arg:"-m,--loglevel" help:"minimum log level (trace/debug/info/warn/error)" default:"warn"`
}

func (args) Description() string {
	return "offlinemsmtp -- offline wrapper for msmtp"
}

func main() {
	// Split os.Args at "--" to separate our flags from msmtp args.
	var msmtpArgs []string
	for i, a := range os.Args {
		if a == "--" {
			msmtpArgs = os.Args[i+1:]
			os.Args = os.Args[:i]
			break
		}
	}

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
		logWriter = f
	}
	if a.Silent {
		logWriter = io.Discard
	}

	logger := zerolog.New(logWriter).Level(level).With().Timestamp().Logger()
	ctx := logger.WithContext(context.Background())

	if a.Daemon {
		logger.Info().Msg("starting offlinemsmtp daemon")
		d := daemon.New(daemon.Config{
			RootDir:      a.OutboxDir,
			ConfigFile:   a.ConfigFile,
			SendMailFile: a.SendMailFile,
			Silent:       a.Silent,
			Interval:     time.Duration(a.Interval) * time.Second,
		})
		if err := d.Run(ctx); err != nil {
			logger.Fatal().Err(err).Msg("daemon error")
		}
		return
	}

	// Queue mode: save stdin + msmtp args to the outbox.
	if err := os.MkdirAll(a.OutboxDir, 0o755); err != nil {
		logger.Fatal().Err(err).Msg("cannot create outbox directory")
	}
	filename := filepath.Join(a.OutboxDir, time.Now().Format("2006-01-02_15-04-05"))
	f, err := os.Create(filename)
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot create outbox file")
	}
	defer f.Close()

	fmt.Fprintln(f, strings.Join(msmtpArgs, " "))
	if _, err := io.Copy(f, bufio.NewReader(os.Stdin)); err != nil {
		logger.Fatal().Err(err).Msg("cannot write email to outbox")
	}
	logger.Info().Str("file", filename).Msg("queued email")
}
