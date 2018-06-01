import os
import time
from queue import Queue
from subprocess import run

from gi.repository import Notify
from offlinemsmtp import util
from watchdog.events import FileSystemEventHandler
from watchdog.observers import Observer


class Daemon(FileSystemEventHandler):
    """Listens for changes to the outbox directory."""

    def __init__(self, args):
        self.connected = False

        # Initialize the queue
        self.queue = Queue()
        for file in os.listdir(args.dir):
            self.queue.put(os.path.join(args.dir, file))

        # Initialize the notifications.
        Notify.init('offlinemsmtp')

    def notify(self, message):
        Notify.Notification.new('offlinemsmtp', message).show()
        print(message)

    def on_created(self, event):
        """Detects when a file is created."""
        self.notify(f'New message: {event.src_path}')
        self.queue.put(event.src_path)
        if util.test_internet():
            self.flush_queue()

    def flush_queue(self):
        """Sends all emails in the queue."""
        self.notify('Flushing the send queue...')
        failed = []
        while not self.queue.empty():
            message = self.queue.get()
            if not os.path.exists(message):
                # It was removed, nothing we can do about that.
                continue

            self.notify(f'Sending {message}...')

            with open(message, 'rb') as message_content:
                msmtp_args = message_content.readline().decode()
                message_content = message_content.read()

            command = ['/usr/bin/msmtp', *msmtp_args.split()]

            sender = run(command, input=message_content)
            if sender.returncode == 0:
                self.notify('Message sent successfully. Removing.')
                os.remove(message)
            else:
                self.notify(f'Message did not send. Putting message back into '
                            f'the queue to try later.\n'
                            f'Error:\n{sender.returncode}')
                failed.append(message)

        for f in failed:
            self.queue.put(f)

    @staticmethod
    def run(args):
        print('Running daemon')

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
