#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import os
import subprocess
import yaml
from retry import retry
from multiprocessing import Pool

with open(os.path.join(os.path.dirname(__file__), "..", "release.yaml"), "r") as f:
    release = yaml.safe_load(f)


def check_image_exists(c):
    name = c["name"] + ":" + c["tag"]
    try:
        check(name)
        print("✅ %s" % name)
    except Exception as e:
        print("💥 %s" % name)
        raise e


@retry(subprocess.CalledProcessError, tries=60, delay=10)
def check(name):
    subprocess.check_call(
        ["docker", "manifest", "inspect", name],
        env=dict(DOCKER_CLI_EXPERIMENTAL="enabled"),
        stderr=subprocess.DEVNULL,
        stdout=subprocess.DEVNULL,
    )


if __name__ == '__main__':
    image_components = list()
    for _, owner_domain in release.get("docker-images", dict()).items():
        for _, application_domain in owner_domain.items():
            for _, component in application_domain.items():
                image_components.append(component)

    with Pool(5) as pool:
        pool.map(check_image_exists, image_components)
