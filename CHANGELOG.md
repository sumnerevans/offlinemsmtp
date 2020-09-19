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
