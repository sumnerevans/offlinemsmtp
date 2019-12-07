import os
import re
from subprocess import run

import gi
gi.require_version('Notify', '0.7')
from gi.repository import Notify

SILENT = False
NOTIFICATIONS_INITIALIZED = False
_APP_NAME = 'offlinemsmtp'


def test_internet(msmtp_config_file, account):
    """
    Tests whether or not the computer can connect to the given account's SMTP
    server.
    """

    # Extract the hosts from the configuration file.
    host_re = re.compile(r'^host (.*)$')
    hosts = []
    with open(os.path.expanduser(msmtp_config_file)) as config_file:
        for line in config_file:
            match = host_re.match(line)
            if match:
                hosts.append(match.group(1))

    # See if we can ping them.
    for h in hosts:
        if run(['ping', '-c', '1', h]).returncode != 0:
            return False
    return True


def notify(message, timeout=None, urgency=Notify.Urgency.LOW):
    """Creates and shows a ``gi.repository.Notify.Notification`` object."""
    global NOTIFICATIONS_INITIALIZED
    print(message)

    if SILENT:
        return

    # Initialize the notifications if necessary.
    if not NOTIFICATIONS_INITIALIZED:
        Notify.init(_APP_NAME)
        NOTIFICATIONS_INITIALIZED = True

    # Create, show, and return the notification
    notification = Notify.Notification.new(_APP_NAME, message)
    if timeout:
        notification.set_timeout(timeout)
    notification.set_urgency(urgency)
    notification.show()
    return notification
