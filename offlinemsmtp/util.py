import requests


def test_internet():
    """
    Tests whether or not the computer is currently connected to the internet.
    """
    try:
        requests.get('https://google.com', timeout=1)
        return True
    except requests.ConnectionError:
        return False
