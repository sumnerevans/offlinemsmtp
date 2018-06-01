import argparse
import os
import sys
from datetime import datetime

from offlinemsmtp.daemon import Daemon


def main():
    # Parse the arguments
    parser = argparse.ArgumentParser(description='offlinemsmtp')
    parser.add_argument(
        '-o',
        '--outbox-directory',
        dest='dir',
        default=os.path.expanduser('~/.offlinemsmtp-outbox'),
        help='The directory to use as the outbox.',
    )
    parser.add_argument(
        '-d',
        '--daemon',
        action='store_true',
        help='Run the offlinemsmtp daemon.',
    )
    parser.add_argument(
        '-i',
        '--interval',
        type=int,
        default=60,
        help='The interval (in seconds) at which to attempt to flush the send queue.',
    )

    args, rest_args = parser.parse_known_args()

    if args.daemon:
        Daemon.run(args)
    else:
        print('Enqueueing message...')
        root_dir = os.path.expanduser(args.dir)
        filename = datetime.now().strftime('%Y-%m-%d_%H-%M-%S')
        with open(os.path.join(root_dir, filename), 'w+') as f:
            # Write the arguments so that the daemon can pass them through.
            f.write(' '.join(rest_args) + '\n')
            # Write all of stdout to the file.
            for line in sys.stdin:
                f.write(line)


if __name__ == '__main__':
    main()
