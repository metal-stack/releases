import time
import unittest

class FlakyBackoff:
    def __init__(self, backoff_seconds=1):
        self._backoff_seconds = backoff_seconds

    def backoff(self, *args):
        time.sleep(self._backoff_seconds)
        return True


class TestinfraCommon(unittest.TestCase):
    def service_enabled_and_running(self, host, service_name):
        service = host.service(service_name)
        self.assertTrue(service.is_enabled, "%s is not enabled" % service_name)
        self.assertTrue(service.is_running, "%s is not running" % service_name)
