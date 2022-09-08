"""Daemon that watches the offlinemsmtp outbox directory for queued emails.

Sends out the emails as soon as they arrive, if the system is online.
"""

import logging
import re
import socket
import threading
import time
from pathlib import Path
from queue import Queue
from subprocess import PIPE, run

import gi
import inotify.adapters

gi.require_version("Notify", "0.7")
from gi.repository import Notify

from offlinemsmtp import util


class Daemon:
    """Listens for changes to the outbox directory."""

    def __init__(self, args):
        """Initialize the daemon."""
        self.connected = False
        self.silent = args.silent
        self.config_file = Path(args.file).resolve()
        self.send_mail_file = Path(args.send_mail_file).resolve() if args.send_mail_file else None
        self.root_dir = Path(args.dir).resolve()

        # Initialize the queue
        self.queue = Queue()
        self.root_dir.mkdir(parents=True, exist_ok=True)
        for file in self.root_dir.iterdir():
            self.queue.put(self.root_dir.joinpath(file))

    def send_enabled(self):
        """Is the file allowing us to send out emails present?

        Always returns True if such a file is not configured.
        """
        return self.send_mail_file is None or self.send_mail_file.exists()

    def on_created(self, filename: Path):
        """Handle file creation."""
        logging.info("New message detected: %s", filename)

        self.queue.put(filename)
        self.flush_queue()

    def flush_queue(self):
        """Send all emails in the queue."""
        if not self.send_enabled():
            util.notify("Sending email disabled", timeout=5000)
            return

        failed = []
        while not self.queue.empty():
            message_path = self.queue.get()
            if not message_path.exists():
                # It was removed, nothing we can do about that.
                continue

            # Open the message.
            with open(message_path, "rb") as message_content:
                msmtp_args = message_content.readline().decode()
                message_content = message_content.read()

            if not self.can_send_message(msmtp_args, message_content):
                failed.append(message_path)
                continue

            # Create a sending notification that lives "forever". It will be
            # closed when the msmtp process completes.
            sending_notification = util.notify(f"Sending {message_path}...", timeout=600000)

            # Send the message.
            logging.debug(self.get_msmtp_command(msmtp_args))
            send_cmd = run(self.get_msmtp_command(msmtp_args), input=message_content, check=False)
            if sending_notification:
                sending_notification.close()

            # Determine whether or not the send was successful or not.
            if send_cmd.returncode == 0:
                util.notify("Message sent successfully. Removing from queue.")
                message_path.unlink()
            else:
                util.notify(
                    f"Message did not send. Putting message back into the "
                    f"queue to try later.\n"
                    f"Return Code: {send_cmd.returncode}\n",
                    timeout=30000,  # 30 seconds
                    urgency=Notify.Urgency.CRITICAL,
                )
                failed.append(message_path)

        # Re-enqueue the failed messages.
        for file in failed:
            self.queue.put(file)

    host_re = re.compile("host = (.*)")
    port_re = re.compile("port = (.*)")
    subject_re = re.compile("Subject: (.*)")

    def get_msmtp_command(self, msmtp_args, pretend=False):
        """Full msmtp command to run to send emails."""
        args = ["/usr/bin/env", "msmtp", "--debug"]
        if pretend:
            args.append("-P")
        args += ["-C", str(self.config_file), *msmtp_args.split()]
        return args

    def can_send_message(self, msmtp_args, message_content):
        """Tests whether or not the computer can connect to the necessary server
        to send the given message.
        """
        test_run = run(
            self.get_msmtp_command(msmtp_args, pretend=True),
            input=message_content,
            stdout=PIPE,
            stderr=PIPE,
            check=False,
        )

        host, port = None, None
        for line in test_run.stdout.decode("utf-8").split("\n"):
            if host_match := self.host_re.match(line):
                host = host_match.group(1)
            elif port_match := self.port_re.match(line):
                port = int(port_match.group(1))

            if host and port:
                break

        # Try to connect to the socket.
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(2)  # 2 second timeout
        try:
            socket_open = sock.connect_ex((host, port))
        except socket.gaierror:
            return False
        finally:
            sock.close()

        # Notify if it's not available.
        if socket_open != 0:
            # Search for the subject in the message_content
            subject = "<no subject>"
            for line in message_content.decode("utf-8").split("\n"):
                subject_match = self.subject_re.match(line)
                if subject_match:
                    subject = subject_match.group(1)

            util.notify(
                f"Cannot connect to {host}:{port} to send message with " f'subject: "{subject}".',
                timeout=5000,
            )
        return socket_open == 0

    @staticmethod
    def run(args):
        """Run the offlinemsmtp daemon."""
        util.notify("offlinemsmtp daemon started")
        # Listen on the outbox directory for new files.
        daemon = Daemon(args)
        observer = inotify.adapters.Inotify()
        observer.add_watch(args.dir.resolve().as_posix())

        def watch_outbox(daemon, observer):
            """Watch the outbox directory and act upon files being added there."""
            for event in observer.event_gen(yield_nones=False):
                _, type_names, path, filename = event
                if "IN_CLOSE_WRITE" in type_names:
                    daemon.on_created(Path(path) / filename)

        observer_thread = threading.Thread(
            target=watch_outbox, args=(daemon, observer), daemon=True
        )
        observer_thread.start()

        try:
            # Every interval, check whether there's anything to send and see if
            # there's an internet connection. If there is, try to flush the
            # send queue.
            while True:
                if not daemon.queue.empty():
                    daemon.flush_queue()

                time.sleep(args.interval)
        except KeyboardInterrupt:
            pass
