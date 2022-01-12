import unittest
import os

from kubernetes import client, config
import metal_python.api as metal_api
from metal_python.driver import Driver
from flaky import flaky
import testinfra

import common


class MetalControlPlaneDeployment(unittest.TestCase):
    def __init__(self, *args, **kwargs):
        super(MetalControlPlaneDeployment, self).__init__(*args, **kwargs)
        config.load_kube_config()
        self.driver = Driver(os.getenv("METALCTL_URL"), os.getenv("METALCTL_TOKEN"), os.getenv("METALCTL_HMAC"))
        self.maxDiff = None

    def test_deployment(self):
        v1 = client.AppsV1Api()
        res = v1.list_namespaced_deployment("metal-control-plane")
        for i in res.items:
            self.assertIsNone(i.status.unavailable_replicas, "not all deployment replicas running")

    def test_stateful_sets(self):
        v1 = client.AppsV1Api()
        res = v1.list_namespaced_stateful_set("metal-control-plane")
        for i in res.items:
            self.assertEqual(i.status.current_replicas, i.status.replicas, "not all stateful set replicas running")

    def test_api_is_responsive(self):
        api = metal_api.VersionApi(api_client=self.driver.client)
        try:
            res = api.info()
        except Exception as exception:
            self.fail("api not responsive: %s" % exception)

        self.assertEqual(res.name, "metal-api")

    def test_api_is_healthy(self):
        api = metal_api.HealthApi(api_client=self.driver.client)
        try:
            res = api.health()
        except Exception as exception:
            self.fail("api not responsive: %s" % exception)

        self.assertEqual(res.status, "healthy")
        self.assertEqual(res.message, "")

    @flaky(max_runs=36, rerun_filter=common.FlakyBackoff(10).backoff)
    def test_switches_have_registered(self):
        api = metal_api.SwitchApi(api_client=self.driver.client)
        try:
            res = api.list_switches()
        except Exception as exception:
            self.fail("listing switches failed: %s" % exception)

        self.assertEqual(len(res), 2, "no two switches have registered")
        for s in res:
            self.assertIsNotNone(s.last_sync, "switch has not synced: %s" % s.name)
            self.assertIsNone(s.last_sync.error, "switch has sync errors: %s" % s.name)

    def test_images_are_present(self):
        api = metal_api.ImageApi(api_client=self.driver.client)
        try:
            res = api.list_images()
        except Exception as exception:
            self.fail("listing images failed: %s" % exception)

        self.assertTrue(len(res) > 0, "no images present")


@unittest.skipUnless(os.path.isdir("/mini-lab"), "tests can only run from deployment container")
class MetalSwitchPlaneDeployment(common.TestinfraCommon):
    def __init__(self, *args, **kwargs):
        super(MetalSwitchPlaneDeployment, self).__init__(*args, **kwargs)
        self.hosts = testinfra.get_hosts(["ssh://mini-lab-leaf01", "ssh://mini-lab-leaf02"])

    def test_metal_core_service(self):
        for host in self.hosts:
            self.service_enabled_and_running(host, "metal-core")

    def test_pixiecore_service(self):
        for host in self.hosts:
            self.service_enabled_and_running(host, "pixiecore")

    def test_frr_service(self):
        for host in self.hosts:
            self.service_enabled_and_running(host, "frr")

    def test_dhcpd_service(self):
        self.service_enabled_and_running(self.hosts[0], "dhcpd")
