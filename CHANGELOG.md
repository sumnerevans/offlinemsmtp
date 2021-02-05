# Version 0.3.9

* INFRASTRUCTURE

  * Got rid of `setup.py` and replaced with `pyproject.toml`.
  * Convert from Twine to Poetry for publishing to PyPi.
  * Use a nix shell for a more consistent development environment.
  * Added a custom style check for TODOs and ensuring that all instances of the
    version are correct. It also checks to make sure that the version number
    throughout the codebase match up.
  * Improvements to the build pipeline
  * Use PyPi credentials that are scoped to only the offlinemsmtp project

# Version 0.3.8

* Use `/usr/bin/env` to find `msmtp` executable for compatibility with NixOS.

# Version 0.3.7

* Fixed dependency issue where sphinx was required as an `install_dependency`
  rather than a dev dependency.

* INFRASTRUCTURE

  * Migrated to sr.ht because of usability regressions in GitLab.
  * Migrated from Pipenv to Poetry because Poetry is actually fast.
  * Added `CONTRIBUTING.md` document to help onboard contributors.
  * Added a `.editorconfig` file to help create consistent development
    environments for contributors.

# Version 0.3.6

* Added the ability to to delimit arguments that should always be sent to
  `msmtp` using `--`.
* Added better README documentation.
* The AUR package now automatically installs the `.service` file to
  `/usr/lib/systemd/user`.
* Convert to use the logging library instead of pure print.

# Version 0.3.5

* Use a real socket to try and connect to the SMTP server instead of ping.
* `offlinemsmtp` only tries to connect to the server that it is sending mail
  to for determining if it should attempt to send that element of the queue.

# Version 0.3.4

* Added `--send-mail-file` config option to allow the user to specify a file
  which must exist for mail sending to be enabled.

# Version 0.3.3

* Change failure timeout on notifications to 30s
