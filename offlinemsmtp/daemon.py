import os
import time
from subprocess import run
from queue import Queue

from watchdog.events import FileSystemEventHandler
from watchdog.observers import Observer

from offlinemsmtp import util


class Daemon(FileSystemEventHandler):
    """Listens for changes to the outbox directory."""

    def __init__(self, args):
        self.connected = False
        self.queue = Queue()

        for file in os.listdir(args.dir):
            self.queue.put(os.path.join(args.dir, file))

    def on_created(self, event):
        """Detects when a file is created."""
        print(f'New message: {event.src_path}')
        self.queue.put(event.src_path)
        if util.test_internet():
            self.flush_queue()

    def flush_queue(self):
        """Sends all emails in the queue."""
        print('Flushing the send queue...')
        failed = []
        while not self.queue.empty():
            message = self.queue.get()
            if not os.path.exists(message):
                # It was removed, nothing we can do about that.
                continue

            print(f'Sending {message}...')

            with open(message, 'rb') as message_content:
                msmtp_args = message_content.readline().decode()
                message_content = message_content.read()

            command = ['/usr/bin/msmtp', *msmtp_args.split()]

            sender = run(command, input=message_content)
            if sender.returncode == 0:
                print('Message sent successfully. Removing.')
                os.remove(message)
            else:
                print('Message did not send. Putting message back into the queue to try later.')
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
