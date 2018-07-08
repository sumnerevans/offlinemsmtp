import os
import time
from queue import Queue
from subprocess import PIPE, call

from gi.repository import Notify
from watchdog.events import FileSystemEventHandler
from watchdog.observers import Observer

from offlinemsmtp import util


class Daemon(FileSystemEventHandler):
    """Listens for changes to the outbox directory."""

    def __init__(self, args):
        """Initialize the daemon."""
        self.connected = False
        self.silent = args.silent
        self.config_file = os.path.expanduser(args.file)

        # Initialize the queue
        self.queue = Queue()
        for file in os.listdir(args.dir):
            self.queue.put(os.path.join(args.dir, file))

    def on_created(self, event):
        """Handle file creation."""
        print(f'New message detected: {event.src_path}')

        self.queue.put(event.src_path)
        if util.test_internet(self.config_file):
            self.flush_queue()
        else:
            # Only notify if it's not going to be sent immediately.
            util.notify(f'New message detected: {event.src_path}')

    def flush_queue(self):
        """Sends all emails in the queue."""
        failed = []
        while not self.queue.empty():
            message = self.queue.get()
            if not os.path.exists(message):
                # It was removed, nothing we can do about that.
                continue

            # Create a sending notification that lives "forever". It will be
            # closed when the sender process completes.
            sending_notification = util.notify(
                f'Sending {message}...',
                timeout=600000,
            )

            # Open the message.
            with open(message, 'rb') as message_content:
                msmtp_args = message_content.readline().decode()
                message_content = message_content.read()

            # Send the message.
            command = [
                '/usr/bin/msmtp', '-C', self.config_file, *msmtp_args.split()
            ]
            sender = call(
                command,
                input=message_content,
                stdout=PIPE,
                stderr=PIPE,
            )
            sending_notification.close()

            # Determine whether or not the send was successful or not.
            if sender.returncode == 0:
                util.notify('Message sent successfully. Removing from queue.')
                os.remove(message)
            else:
                util.notify(
                    f'Message did not send. Putting message back into the '
                    f'queue to try later.\n'
                    f'Return Code: {sender.returncode}\n'
                    f'Error: {sender.stderr.decode()}',
                    urgency=Notify.Urgency.CRITICAL)
                failed.append(message)

        # Re-enqueue the failed messages.
        for f in failed:
            self.queue.put(f)

    @staticmethod
    def run(args):
        """Run the offlinemsmtp daemon."""
        util.notify('offlinemsmtp daemon started')
        # Listen on the outbox directory for new files.
        daemon = Daemon(args)
        observer = Observer()
        observer.schedule(daemon, args.dir, recursive=True)
        observer.start()

        try:
            # Every interval, check whether there's anything to send and see if
            # there's an internet connection. If there is, try to flush the
            # send queue.
            while True:
                if not daemon.queue.empty() and util.test_internet(args.file):
                    daemon.flush_queue()

                time.sleep(args.interval)
        except KeyboardInterrupt:
            observer.stop()
        observer.join()
