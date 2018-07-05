offlinemsmtp
============

Allows you to use ``msmtp`` offline.

Installation
------------

Using PyPi::

    pip install --user offlinemsmtp

Run the daemon using systemd
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Create a file called ``~/.config/systemd/user/offlinemsmtp.service`` with the
following content::

    [Unit]
    Description=Offline msmtp

    [Service]
    ExecStart=/home/sumner/.local/bin/offlinemsmtp --daemon

    [Install]
    WantedBy=default.target

Then, enable and start ``offlinemsmtp`` using systemd::

    systemctl --user daemon-reload
    systemctl --user enable --now offlinemsmtp

Other projects
--------------

- https://github.com/dcbaker/py-mailqueued - looks cool, I didn't see it when I
  was researching, but it's probably better than my implementation, even thought
  I had a lot of fun doing mine
- https://github.com/venkytv/msmtp-offline - it's written in Ruby
