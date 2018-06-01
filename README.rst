offlinemsmtp
============

Allows you to use ``msmtp`` offline.

To enable ``offlinemsmtp`` using systemd::

    systemctl --user daemon-reload
    systemctl --user enable --now offlinemsmtp

Other projects
--------------

- https://github.com/dcbaker/py-mailqueued - looks cool, I didn't see it when I
  was researching, but it's probably better than my implementation, even thought
  I had a lot of fun doing mine
