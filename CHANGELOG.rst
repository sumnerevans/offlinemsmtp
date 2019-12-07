v0.3.5
======

* Use a real socket to try and connect to the SMTP server instead of ping.
* ``offlinemsmtp`` only tries to connect to the server that it is sending mail
  to for determining if it should attempt to send that element of the queue.

v0.3.4
======

- Added ``--send-mail-file`` config option to allow the user to specify a file
  which must exist for mail sending to be enabled.

v0.3.3
======

- Change failure timeout on notifications to 30s
