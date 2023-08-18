#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import os
import subprocess
import yaml
from retry import retry
from multiprocessing import Pool
import requests

with open(os.path.join(os.path.dirname(__file__), "..", "release.yaml"), "r") as f:
    release = yaml.safe_load(f)

NUMBER_OF_POOLS = 6


def check_image_exists(image):
    @retry(tries=60, delay=10)
    def check(name):
        subprocess.check_output(
            ["docker manifest inspect " + name],
            env=dict(DOCKER_CLI_EXPERIMENTAL="enabled"),
            stderr=subprocess.STDOUT,
            shell=True, # this is required to pick up the config.json for image pull secrets
            universal_newlines=True,
        )

    image_name = image[0] + ":" + image[1]

    if image_name.startswith('k8s.gcr.io'):
        print("➖ %s (skipped because registry does not allow manifest inspection)" % image_name)
        return

    try:
        check(image_name)
        print("✅ %s" % image_name)
    except subprocess.CalledProcessError as e:
        print("💥 %s" % image_name)
        raise RuntimeError(e.output) from e


def check_url_exists(url):
    @retry(RuntimeError, tries=60, delay=10)
    def check(u):
        request = requests.head(u)
        if request.status_code == 200:
            return
        if request.status_code == 302:
            return
        raise RuntimeError("url %s does not exist: %s" % (u, request.status_code))

    try:
        check(url)
        print("✅ %s" % url)
    except Exception as e:
        print("💥 %s" % url)
        raise e


def get_recursively(search_dict, fields):
    fields_found = []

    matches = []
    for key, value in search_dict.items():
        if isinstance(fields, tuple):
            for field in fields:
                if key == field:
                    matches.append(value)
        else:
            field = fields
            if key == field:
                fields_found.append(value)

    if matches:
        fields_found.append(tuple(matches))

    for key, value in search_dict.items():
        if isinstance(value, dict):
            results = get_recursively(value, fields)
            for result in results:
                fields_found.append(result)

        elif isinstance(value, list):
            for item in value:
                if isinstance(item, dict):
                    more_results = get_recursively(item, fields)
                    for another_result in more_results:
                        fields_found.append(another_result)

    return fields_found


if __name__ == '__main__':
    images = get_recursively(release.get("docker-images", dict()), ("name", "tag"))
    binaries = get_recursively(release.get("binaries", dict()), "url")

    with Pool(NUMBER_OF_POOLS) as pool:
        pool.map(check_url_exists, binaries)
        pool.map(check_image_exists, images)
