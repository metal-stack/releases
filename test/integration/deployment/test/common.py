import unittest


class TestinfraCommon(unittest.TestCase):
    def service_enabled_and_running(self, host, service_name):
        service = host.service(service_name)
        self.assertTrue(service.is_enabled, "%s is not enabled" % service_name)
        self.assertTrue(service.is_running, "%s is not running" % service_name)
