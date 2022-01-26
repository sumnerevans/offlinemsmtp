![offlinemsmtp](./logo/logo.png)

Allows you to use `msmtp` offline by queuing email until you have an internet
connection.

[![Lint and Build](https://github.com/sumnerevans/offlinemsmtp/actions/workflows/build.yaml/badge.svg)](https://github.com/sumnerevans/offlinemsmtp/actions/workflows/build.yaml)
[![PyPi Version](https://img.shields.io/pypi/v/offlinemsmtp?color=4DC71F&logo=python&logoColor=fff)](https://pypi.org/project/offlinemsmtp/)
[![AUR Version](https://img.shields.io/aur/version/offlinemsmtp?logo=linux&logoColor=fff)](https://aur.archlinux.org/packages/offlinemsmtp/)
[![LiberaPay Donation Status](https://img.shields.io/liberapay/receives/sumner.svg?logo=liberapay)](https://liberapay.com/sumner/donate)

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

Using [PyPi](https://pypi.org/project/offlinemsmtp/):

    pip install --user offlinemsmtp

On Arch Linux, you can install the `offlinemsmtp` package from the
[AUR](https://aur.archlinux.org/packages/offlinemsmtp/). For example, if you use
`yay`:

    yay -S offlinemsmtp

## Run the daemon using systemd

Create a file called ``~/.config/systemd/user/offlinemsmtp.service`` with the
following content (if you installed via the AUR package, a service file was
already created for you in ``/usr/lib/systemd/user`` so you only need to do this
step if you want to customize the parameters passed to the daemon):

    [Unit]
    Description=offlinemsmtp

    [Service]
    ExecStart=/usr/bin/offlinemsmtp --daemon

    [Install]
    WantedBy=default.target

Then, enable and start `offlinemsmtp` using systemd:

    systemctl --user daemon-reload
    systemctl --user enable --now offlinemsmtp

## Usage

`offlinemsmtp` has two components: a daemon for listening to the outbox folder
and sending the mail when the network is available and a enqueuer for adding
mail to the send queue.

To run the daemon in the current command line (this is useful for testing), run
this command::

    offlinemsmtp --daemon

To enqueue emails, use the `offlinemsmtp` executable without `--daemon`. All
parameters (with a few caveats described below in [Command Line
Arguments](#command-line-arguments)) are forwarded on to `msmtp`. Anything
passed in via standard in will be forwarded over standard in to `msmtp` when the
mail is sent.

### Configuration with Mutt

To use offlinemsmtp with mutt, just replace `msmtp` in your mutt configuration
file with `offlinemsmtp`. Here is an example:

    set sendmail = "offlinemsmtp -a personal"

### Command Line Arguments

offlinemsmtp accepts a number of command line arguments:

- `-h`, `--help` - shows a help message and exits.
- `-o DIR`, `--outbox-directory DIR` - set the directory to use as the outbox.
  Defaults to `~/.offlinemsmtp-outbox`.
- `-d`, `--daemon` - run the offlinemsmtp daemon.
- `-s`, `--silent` - set to disable all logging and notifications.
- `-i INTERVAL`, `--interval INTERVAL` - set the interval (in seconds) at which
  to attempt to flush the send queue. Defaults to 60.
- `-C FILE`, `--file FILE` - the msmtp configuration file to use.
- `--send-mail-file FILE` - only send mail if this file exists (defaults to
  `None` meaning that no file is required for mail sending to be enabled)
- All remaining arguments are passed to `msmtp`. The `-C` argument is
  automatically passed to `msmtp`.
- Anything after a special `--` argument will be passed to `msmtp`. This allows
  you to pass arguments that may conflict with `offlinemsmtp` arguments to
  `msmtp`.

## Contributing

See the [CONTRIBUTING.md](./CONTRIBUTING.md) document for details on how to
contribute to the project.

## Other projects

- https://github.com/marlam/msmtp-mirror/tree/master/scripts/msmtpqueue - this
  is included with `msmtp`, but doesn't have all of the features that I want.
- https://github.com/dcbaker/py-mailqueued - looks cool, I didn't see it when I
  was researching, but it's probably better than my implementation, even thought
  I had a lot of fun doing mine.
- https://github.com/venkytv/msmtp-offline - it's written in Ruby.
