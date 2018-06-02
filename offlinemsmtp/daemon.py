import os
import time
from queue import Queue
from subprocess import PIPE, run

from gi.repository import Notify
from watchdog.events import FileSystemEventHandler
from watchdog.observers import Observer

from offlinemsmtp import util


class Daemon(FileSystemEventHandler):
    """Listens for changes to the outbox directory."""

    def __init__(self, args):
        self.connected = False
        self.silent = args.silent

        # Initialize the queue
        self.queue = Queue()
        for file in os.listdir(args.dir):
            self.queue.put(os.path.join(args.dir, file))

    def on_created(self, event):
        """Detects when a file is created."""
        print(f'New message detected: {event.src_path}')

        self.queue.put(event.src_path)
        if util.test_internet():
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

            sending_notification = util.notify(
                f'Sending {message}...',
                timeout=600000,
            )

            with open(message, 'rb') as message_content:
                msmtp_args = message_content.readline().decode()
                message_content = message_content.read()

            command = ['/usr/bin/msmtp', *msmtp_args.split()]

            sender = run(
                command, input=message_content, stdout=PIPE, stderr=PIPE)
            sending_notification.close()
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
        util.notify('offlinemsmtp daemon started')
        # Listen on the outbox directory for new files.
        daemon = Daemon(args)
        observer = Observer()
        observer.schedule(daemon, args.dir, recursive=True)
        observer.start()

        try:
            while True:
                # There's something enqueued that hasn't been sent yet.
                if not daemon.queue.empty() and util.test_internet():
                    daemon.flush_queue()

                time.sleep(args.interval)
        except KeyboardInterrupt:
            observer.stop()
        observer.join()
