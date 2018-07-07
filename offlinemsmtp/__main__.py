import argparse
import os
import sys
from datetime import datetime

from offlinemsmtp import util
from offlinemsmtp.daemon import Daemon


def main():
    # Parse the arguments
    parser = argparse.ArgumentParser(description='Offline wrapper for msmtp.')
    parser.add_argument(
        '-o',
        '--outbox-directory',
        dest='dir',
        default=os.path.expanduser('~/.offlinemsmtp-outbox'),
        help=('set the directory to use as the outbox. Defaults to '
              '~/.offlinemsmtp-outbox.'),
    )
    parser.add_argument(
        '-d',
        '--daemon',
        action='store_true',
        help='run the offlinemsmtp daemon.',
    )
    parser.add_argument(
        '-s',
        '--silent',
        action='store_true',
        help='set to disable all logging and notifications',
    )
    parser.add_argument(
        '-i',
        '--interval',
        type=int,
        default=60,
        help=('set the interval (in seconds) at which to attempt to flush the '
              'send queue. Defaults to 60.'),
    )
    parser.add_argument(
        '-C',
        '--file',
        default=os.path.expanduser('~/.msmtprc'),
        help='the msmtp configuration file to use',
    )

    args, rest_args = parser.parse_known_args()
    util.SILENT = args.silent

    if args.daemon:
        Daemon.run(args)
    else:
        root_dir = os.path.expanduser(args.dir)
        filename = datetime.now().strftime('%Y-%m-%d_%H-%M-%S')
        with open(os.path.join(root_dir, filename), 'w+') as f:
            # Write the arguments on the first line so that the daemon can pass
            # them through.
            f.write(' '.join(rest_args) + '\n')

            # Write all of stdout to the file.
            for line in sys.stdin:
                f.write(line)


if __name__ == '__main__':
    main()
