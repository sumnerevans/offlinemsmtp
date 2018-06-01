offlinemsmtp
============

Allows you to use ``msmtp`` offline.

Installation
------------

Using PyPi::

    pip install --user offlinemsmtp

.. To enable ``offlinemsmtp`` using systemd (doesn't work right now)::

..    systemctl --user daemon-reload
..    systemctl --user enable --now offlinemsmtp

Other projects
--------------

- https://github.com/dcbaker/py-mailqueued - looks cool, I didn't see it when I
  was researching, but it's probably better than my implementation, even thought
  I had a lot of fun doing mine
