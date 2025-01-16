ifeq ($(CI),true)
MINI_LAB_PATH := $(PWD)/mini-lab
DOCKER_TTY_ARG=
else
MINI_LAB_PATH := $(PWD)/../mini-lab
DOCKER_TTY_ARG=t
endif

define mini-lab-env
		$(eval ENV_FILE := $(MINI_LAB_PATH)/.env)
		$(eval include $(MINI_LAB_PATH)/.env)
		$(eval export sed 's/=.*//' $(MINI_LAB_PATH)/.env)
endef

.PHONY: prep
prep:
	@./test/integration/prep.sh $(MINI_LAB_PATH)
	@kind get clusters | grep metal-control-plane > /dev/null || $(MAKE) mini-lab

.PHONY: mini-lab
mini-lab:
	cd $(MINI_LAB_PATH) && make

.PHONY: wait-for-images
wait-for-images:
	docker run --rm -i$(DOCKER_TTY_ARG) \
		-v $(PWD):/test:ro \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(HOME)/.docker:/root/.docker \
		-w /test \
		-e PYTHONUNBUFFERED=1 \
		python:alpine sh -c 'apk add docker-cli --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community && pip install --upgrade pip retry requests pyyaml && ./test/wait_for_images.py'
# docker-cli has to be version >= 23.x as otherwise some images will find no manifests.
# the --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community flag can be removed
# as soon as the release went from edge into alpine's mainline.

.PHONY: integration-ansible-modules
integration-ansible-modules: prep
	@$(call mini-lab-env)
	docker run --rm -i$(DOCKER_TTY_ARG) \
		-v $(MINI_LAB_PATH):/mini-lab \
		-v $(PWD)/test/integration/ansible-modules:/integration:ro \
		-v $(PWD)/test/integration/ansible-modules/output:/output \
		-w /mini-lab \
		-e METAL_ANSIBLE_INVENTORY_CONFIG=/root/.ansible/roles/metal-ansible-modules/inventory/metal_config.example.yaml \
		-e PYTHONUNBUFFERED=1 \
		--env-file $(MINI_LAB_PATH)/.env \
		--network host \
		ghcr.io/metal-stack/metal-deployment-base:$(DEPLOYMENT_BASE_IMAGE_TAG) /integration/integration.sh

.PHONY: integration-deployment
integration-deployment: prep
	@$(call mini-lab-env)
	docker run --rm -i$(DOCKER_TTY_ARG) \
		-v $(MINI_LAB_PATH):/mini-lab \
		-v $(PWD)/test/integration/deployment:/integration:ro \
		-v $(PWD)/test/integration/deployment/output:/output \
		-w /integration \
		-e PYTHONUNBUFFERED=1 \
		--network host \
		ghcr.io/metal-stack/metal-deployment-base:$(DEPLOYMENT_BASE_IMAGE_TAG) /integration/integration.sh

.PHONY: render-junit
render-junit:
	docker run --rm -v $(PWD)/test/integration/ansible-modules/output:/results maxmiorim/junit-viewer > results.html
