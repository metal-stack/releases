import time


class FlakyBackoff:
    def __init__(self, backoff_seconds=1):
        self._backoff_seconds = backoff_seconds

    def backoff(self):
        time.sleep(self._backoff_seconds)
        return True
