import logging
import re
import socket
import time
from pathlib import Path
from queue import Queue
from subprocess import PIPE, run

import gi

gi.require_version("Notify", "0.7")
from gi.repository import Notify
from offlinemsmtp import util
from watchdog.events import FileSystemEventHandler
from watchdog.observers import Observer


class Daemon(FileSystemEventHandler):
    """Listens for changes to the outbox directory."""

    def __init__(self, args):
        """Initialize the daemon."""
        self.connected = False
        self.silent = args.silent
        self.config_file = Path(args.file).resolve() if args.file else None
        self.send_mail_file = (
            Path(args.send_mail_file).resolve() if args.send_mail_file else None
        )
        self.root_dir = Path(args.dir).resolve() if args.dir else None

        # Initialize the queue
        self.queue = Queue()
        self.root_dir.mkdir(parents=True, exist_ok=True)
        for file in self.root_dir.iterdir():
            self.queue.put(self.root_dir.joinpath(file))

    def send_enabled(self):
        return self.send_mail_file is None or self.send_mail_file.exists()

    def on_created(self, event):
        """Handle file creation."""
        logging.info(f"New message detected: {event.src_path}")

        self.queue.put(Path(event.src_path))
        self.flush_queue()

    def flush_queue(self):
        """Sends all emails in the queue."""
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
            # closed when the sender process completes.
            sending_notification = util.notify(
                f"Sending {message_path}...", timeout=600000
            )

            # Send the message.
            sender = run(
                self.get_msmtp_command(msmtp_args),
                input=message_content,
                stdout=PIPE,
                stderr=PIPE,
            )
            sending_notification.close()

            # Determine whether or not the send was successful or not.
            if sender.returncode == 0:
                util.notify("Message sent successfully. Removing from queue.")
                message_path.unlink()
            else:
                util.notify(
                    f"Message did not send. Putting message back into the "
                    f"queue to try later.\n"
                    f"Return Code: {sender.returncode}\n"
                    f"Error: {sender.stderr.decode()}",
                    timeout=30000,  # 30 seconds
                    urgency=Notify.Urgency.CRITICAL,
                )
                failed.append(message_path)

        # Re-enqueue the failed messages.
        for f in failed:
            self.queue.put(f)

    host_re = re.compile("host = (.*)")
    port_re = re.compile("port = (.*)")
    subject_re = re.compile("Subject: (.*)")

    def get_msmtp_command(self, msmtp_args, pretend=False):
        args = ["/usr/bin/msmtp"]
        if pretend:
            args.append("-P")
        args += ["-C", str(self.config_file), *msmtp_args.split()]
        return args

    def can_send_message(self, msmtp_args, message_content):
        """
        Tests whether or not the computer can connect to the necessary server
        to send the given message.
        """
        test_run = run(
            self.get_msmtp_command(msmtp_args, pretend=True),
            input=message_content,
            stdout=PIPE,
            stderr=PIPE,
        )

        for line in test_run.stdout.decode("utf-8").split("\n"):
            # TODO convert this stuff to Walrus operators once that's more
            # supported by tools.
            host_match = self.host_re.match(line)
            if host_match:
                host = host_match.group(1)
                continue

            port_match = self.port_re.match(line)
            if port_match:
                port = int(port_match.group(1))

        # Try to connect to the socket.
        sock = socket.socket()
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
                f"Cannot connect to {host}:{port} to send message with "
                f'subject: "{subject}".',
                timeout=5000,
            )
        return socket_open == 0

    @staticmethod
    def run(args):
        """Run the offlinemsmtp daemon."""
        util.notify("offlinemsmtp daemon started")
        # Listen on the outbox directory for new files.
        daemon = Daemon(args)
        observer = Observer()
        observer.schedule(daemon, str(args.dir.resolve()), recursive=True)
        observer.start()

        try:
            # Every interval, check whether there's anything to send and see if
            # there's an internet connection. If there is, try to flush the
            # send queue.
            while True:
                if not daemon.queue.empty():
                    daemon.flush_queue()

                time.sleep(args.interval)
        except KeyboardInterrupt:
            observer.stop()
        observer.join()
