offlinemsmtp
============

Allows you to use ``msmtp`` offline.

Installation
------------

Using PyPi::

    pip install --user offlinemsmtp

On Arch Linux, you can install the ``offlinemsmtp`` package from the AUR. For
example, if you use ``aurman``::

    aurman -S offlinemsmtp

Run the daemon using systemd
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Create a file called ``~/.config/systemd/user/offlinemsmtp.service`` with the
following content::

    [Unit]
    Description=Offline msmtp

    [Service]
    Type=forking
    ExecStart=/usr/bin/offlinemsmtp --daemon

    [Install]
    WantedBy=default.target

Then, enable and start ``offlinemsmtp`` using systemd::

    systemctl --user daemon-reload
    systemctl --user enable --now offlinemsmtp

Usage
-----

Offlinemsmtp has two components: a daemon for listening to the outbox folder and
sending the mail when the network is available and a enqueuer for adding mail to
the send queue.

To run the daemon in the current command line (this is useful for testing), run
this command::

    offlinemsmtp --daemon

To enqueue emails, use the ``offlinemsmtp`` executable without ``--daemon``. All
parameters (besides the ones described below in `Command Line Arguments`_) are
forwarded on to ``msmtp``. Anything passed in via standard in will be forwarded
over standard in to ``msmtp`` when the mail is sent.

Configuration with Mutt
^^^^^^^^^^^^^^^^^^^^^^^

To use offlinemsmtp with mutt, just replace ``msmtp`` in your mutt configuration
file with ``offlinemsmtp``. Here is an example::

    set sendmail = "offlinemsmtp -a personal"

Command Line Arguments
^^^^^^^^^^^^^^^^^^^^^^

offlinemsmtp accepts a number of command line arguments:

- ``-h``, ``--help`` - shows a help message and exits.
- ``-o DIR``, ``--outbox-directory DIR`` - set the directory to use as the
  outbox. Defaults to ``~/.offlinemsmtp-outbox``.
- ``-d``, ``--daemon`` - run the offlinemsmtp daemon.
- ``-s``, ``--silent`` - set to disable all logging and notifications.
- ``-i INTERVAL``, ``--interval INTERVAL`` - set the interval (in seconds) at
  which to attempt to flush the send queue. Defaults to 60.
- ``-C FILE``, ``--file FILE`` - the msmtp configuration file to use.

Other projects
--------------

- https://github.com/dcbaker/py-mailqueued - looks cool, I didn't see it when I
  was researching, but it's probably better than my implementation, even thought
  I had a lot of fun doing mine
- https://github.com/venkytv/msmtp-offline - it's written in Ruby
