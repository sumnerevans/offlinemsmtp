![offlinemsmtp](./logo/logo.png)

Allows you to use `msmtp` offline by queuing email until you have an internet
connection.

[![Build](https://github.com/sumnerevans/offlinemsmtp/actions/workflows/go.yml/badge.svg)](https://github.com/sumnerevans/offlinemsmtp/actions/workflows/go.yml)
[![Sponsor Badge](https://img.shields.io/github/sponsors/sumnerevans?logo=github)](https://github.com/sponsors/sumnerevans)

## Features

* Runs as a daemon and (at a configurable time interval) attempts to send the
  mail in the queue directory.
* Drop-in replacement for `msmtp` in your mutt config.
* Only attempts to send the queued email message if it can connect to the
  configured SMTP server.
* When a new email message comes into the queue and you are already online,
  `offlinemsmtp` will send it immediately.
* Integrates with system notifications so that you are notified when mail is
  being sent.
* Disable/enable sending of mail by the presence/absence of a file. This is
  useful if you want to have some sort of "offline mode".

## Installation

### Using `go install`

```
go install github.com/sumnerevans/offlinemsmtp/cmd/offlinemsmtp@latest
```

### Using Nix

Run ad-hoc:

```
nix run github:sumnerevans/offlinemsmtp
```

Using flakes:

```nix
{
  inputs.offlinemsmtp = {
    url = "github:sumnerevans/offlinemsmtp";
    inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = { offlinemsmtp, ... }: {
    // use the package: offlinemsmtp.packages."x86_64-linux".offlinemsmtp
  };
}
```

### From source

```
git clone https://github.com/sumnerevans/offlinemsmtp
cd offlinemsmtp
go build -o offlinemsmtp ./cmd/offlinemsmtp
```

## Run the daemon using systemd

Create a file called `~/.config/systemd/user/offlinemsmtp.service` with the
following content:

    [Unit]
    Description=offlinemsmtp

    [Service]
    ExecStart=/usr/bin/offlinemsmtp --daemon

    [Install]
    WantedBy=default.target

Then enable and start it:

    systemctl --user daemon-reload
    systemctl --user enable --now offlinemsmtp

## Usage

`offlinemsmtp` has two components: a daemon that watches the outbox directory
and sends mail when the network is available, and an enqueuer that adds mail to
the send queue.

To run the daemon:

    offlinemsmtp --daemon

To enqueue emails, use `offlinemsmtp` without `--daemon`. Pass msmtp arguments
after `--`. Anything on stdin is forwarded to `msmtp` when the mail is sent:

    offlinemsmtp -C ~/.msmtprc -- -t --read-envelope-from

### Configuration with Mutt

Replace `msmtp` in your mutt configuration with `offlinemsmtp`:

    set sendmail = "offlinemsmtp -a personal --"

### Command Line Arguments

- `-h`, `--help` - show help and exit
- `-o DIR`, `--outbox-directory DIR` - outbox directory (default: `~/.offlinemsmtp-outbox`)
- `-d`, `--daemon` - run the daemon
- `-s`, `--silent` - disable all logging and notifications
- `-i INTERVAL`, `--interval INTERVAL` - flush interval in seconds (default: 60)
- `-C FILE`, `--file FILE` - msmtp configuration file (default: `~/.msmtprc`)
- `--send-mail-file FILE` - only send mail if this file exists
- `-l FILE`, `--logfile FILE` - write logs to file
- `-m LEVEL`, `--loglevel LEVEL` - minimum log level: `trace`, `debug`, `info`, `warn`, `error` (default: `warn`)
- Everything after `--` is passed to `msmtp`

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md).

## Other projects

- https://github.com/marlam/msmtp-mirror/tree/master/scripts/msmtpqueue - included
  with `msmtp`, but fewer features.
- https://github.com/dcbaker/py-mailqueued - Python implementation.
- https://github.com/venkytv/msmtp-offline - Ruby implementation.
