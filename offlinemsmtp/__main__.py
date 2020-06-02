import argparse
import logging
import sys
from datetime import datetime
from pathlib import Path

from offlinemsmtp import util
from offlinemsmtp.daemon import Daemon


def main():
    # Parse the arguments
    parser = argparse.ArgumentParser(description="Offline wrapper for msmtp.")
    parser.add_argument(
        "-o",
        "--outbox-directory",
        dest="dir",
        default=Path.home().joinpath(".offlinemsmtp-outbox"),
        help=(
            "set the directory to use as the outbox. Defaults to "
            "~/.offlinemsmtp-outbox."
        ),
    )
    parser.add_argument(
        "-d", "--daemon", action="store_true", help="run the offlinemsmtp daemon.",
    )
    parser.add_argument(
        "-s",
        "--silent",
        action="store_true",
        help="set to disable all logging and notifications",
    )
    parser.add_argument(
        "-i",
        "--interval",
        type=int,
        default=60,
        help=(
            "set the interval (in seconds) at which to attempt to flush the send queue."
            " Defaults to 60."
        ),
    )
    parser.add_argument(
        "-C",
        "--file",
        default=Path.home().joinpath(".msmtprc"),
        help="the msmtp configuration file to use",
    )
    parser.add_argument(
        "--send-mail-file", default=None, help="only send mail if this file exists",
    )
    parser.add_argument(
        "-l", "--logfile", help="the filename to send logs to",
    )
    parser.add_argument(
        "-m", "--loglevel", help="the minium level of logging to do", default="WARNING",
    )

    if "--" in sys.argv:
        dash_arg_pos = sys.argv.index("--")
        before, after = (
            sys.argv[1:dash_arg_pos],  # Ignore argv[0] which is the program itself.
            sys.argv[dash_arg_pos + 1 :],
        )
        args, rest_args = parser.parse_known_args(before)
        rest_args.extend(after)
    else:
        args, rest_args = parser.parse_known_args()
    util.SILENT = args.silent

    min_log_level = getattr(logging, args.loglevel.upper(), None)
    if not isinstance(min_log_level, int):
        logging.error(f"Invalid log level: {args.loglevel.upper()}.")
        min_log_level = logging.WARNING

    logging.basicConfig(
        filename=args.logfile,
        level=min_log_level,
        format="%(asctime)s:%(levelname)s:%(name)s:%(module)s:%(message)s",
    )

    if args.daemon:
        logging.info("Starting the offlinemsmtp daemon.")
        Daemon.run(args)
    else:
        root_dir = Path(args.dir).resolve()
        root_dir.mkdir(parents=True, exist_ok=True)
        filename = datetime.now().strftime("%Y-%m-%d_%H-%M-%S")
        with open(root_dir.joinpath(filename), "w+") as f:
            # Write the arguments on the first line so that the daemon can pass
            # them through.
            f.write(" ".join(rest_args) + "\n")

            # Write all of stdout to the file.
            for line in sys.stdin:
                f.write(line)


if __name__ == "__main__":
    main()
