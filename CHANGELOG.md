# Version 1.0.0

offlinemsmtp has been rewritten in Go. In general, the functionality of v0.4.0
is preserved, however if you were using some uncommonly-used features, you may
need to update your configuration slightly. This version also relicenses the
project under under MIT.

## Breaking Changes

- Installation and building now uses the Go tooling. You can no longer install
  from pip. See [the installation instructions](README.md#Installation) for the
  latest instructions for installation.
- Arguments passed to `msmtp` must now be separated by `--` (e.g.
  `offlinemsmtp -C ~/.msmtprc -- -t --read-envelope-from`). Previously,
  unrecognised flags were forwarded automatically.
- Log levels use zerolog names (`trace`, `debug`, `info`, `warn`, `error`)
  instead of Python's `DEBUG`, `WARNING`, etc.
- `-l` now logs to both STDERR and the specified file simultaneously. This is
  likely not breaking in most cases, but relevant if you were relying on
  receiving no output from the program.

## New Features

- **Network Online Detection:** when NetworkManager reports the network has come
  online, offlinemsmtp tries to send enqueued emails (no need to wait for the
  next flush interval).
- **Notifications have been greatly improved.**
  - They now show the email subject so you can tell which message is being sent,
    succeeded, or failed.
  - The "Sending..." notification stays visible until the send completes or
    fails, then updates in place with the result.
- **Additional Configuration Options**
  - `--debounce-interval` (default 10s) prevents repeated sends when multiple
    filesystem events arrive in quick succession.
  - `--msmtp-path` lets you specify the full path to the msmtp binary, useful on
    NixOS and other non-standard PATH setups.

# Version 0.4.0

- **Dependency change**: the `watchdog` dependency has been replaced by
  `inotify`.

- Infrastructure/DX Changes

  - Added pre-commit and isort
  - Migrated to GitHub
  - Added dependabot to auto-update GitHub Actions versions
  - Converted the CI to not use Nix for linting and building (it's now way
    faster)

# Version 0.3.10

- Require latest `PyGObject` and `watchdog` dependencies.

# Version 0.3.9

- Wait for one second after the path gets created on disk to allow the file to
  be fully written.
- Fixed `PyGObject` dependency

- INFRASTRUCTURE

  - Migrated to GitHub and GitHub Actions
  - Added a custom style check for TODOs and ensuring that all instances of the
    version are correct.
  - Add a `shell.nix` for a more consistent development environment, and use it
    with direnv.
  - Got rid of `setup.py` and replaced with `pyproject.toml`.

# Version 0.3.8

- Use `/usr/bin/env` to find `msmtp` executable for compatibility with NixOS.

# Version 0.3.7

- Fixed dependency issue where sphinx was required as an `install_dependency`
  rather than a dev dependency.

- INFRASTRUCTURE

  - Migrated to sr.ht because of usability regressions in GitLab.
  - Migrated from Pipenv to Poetry because Poetry is actually fast.
  - Added `CONTRIBUTING.md` document to help onboard contributors.
  - Added a `.editorconfig` file to help create consistent development
    environments for contributors.

# Version 0.3.6

- Added the ability to to delimit arguments that should always be sent to
  `msmtp` using `--`.
- Added better README documentation.
- The AUR package now automatically installs the `.service` file to
  `/usr/lib/systemd/user`.
- Convert to use the logging library instead of pure print.

# Version 0.3.5

- Use a real socket to try and connect to the SMTP server instead of ping.
- `offlinemsmtp` only tries to connect to the server that it is sending mail to
  for determining if it should attempt to send that element of the queue.

# Version 0.3.4

- Added `--send-mail-file` config option to allow the user to specify a file
  which must exist for mail sending to be enabled.

# Version 0.3.3

- Change failure timeout on notifications to 30s
