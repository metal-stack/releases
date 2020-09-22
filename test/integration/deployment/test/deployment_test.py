import unittest
import os

from kubernetes import client, config
import metal_python.api as metal_api
from metal_python.driver import Driver
from flaky import flaky

import common


class MetalDeployment(unittest.TestCase):
    def setUp(self):
        config.load_kube_config()
        self.driver = Driver(os.getenv("METALCTL_URL"), os.getenv("METALCTL_TOKEN"), os.getenv("METALCTL_HMAC"))
        self.maxDiff = None

    def test_deployment(self):
        v1 = client.AppsV1Api()
        res = v1.list_namespaced_deployment("metal-control-plane")
        for i in res.items:
            self.assertIsNone(i.status.unavailable_replicas, "not all deployment replicas running")

    def test_api_is_responsive(self):
        api = metal_api.VersionApi(api_client=self.driver.client)
        try:
            res = api.info()
        except Exception as exception:
            self.fail("api not responsive: %s" % exception)

        self.assertEqual(res.name, "metal-api")

    @flaky(max_runs=36, rerun_filter=common.FlakyBackoff(10).backoff)
    def test_switches_have_registered(self):
        api = metal_api.SwitchApi(api_client=self.driver.client)
        try:
            res = api.list_switches()
        except Exception as exception:
            self.fail("listing switches failed: %s" % exception)

        self.assertEqual(len(res), 2, "no two switches have registered")
        for s in res:
            self.assertIsNone(s.last_sync, "switch has not synced: %s" % s.name)
            self.assertIsNotNone(s.last_sync.error, "switch has sync errors: %s" % s.name)

    def test_images_are_present(self):
        api = metal_api.ImageApi(api_client=self.driver.client)
        try:
            res = api.list_images()
        except Exception as exception:
            self.fail("listing images failed: %s" % exception)

        self.assertTrue(len(res) > 0, "no images present")
