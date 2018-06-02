import requests
from gi.repository import Notify

SILENT = False
NOTIFICATIONS_INITIALIZED = False
_APP_NAME = 'offlinemsmtp'


def test_internet():
    """
    Tests whether or not the computer is currently connected to the internet.
    """
    try:
        requests.get('https://google.com', timeout=1)
        return True
    except requests.ConnectionError:
        return False


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
