ifeq ($(CI),true)
MINI_LAB_PATH := $(PWD)/mini-lab
else
MINI_LAB_PATH := $(PWD)/../mini-lab
endif

.PHONY: prep
prep:
	@./test/integration/prep.sh $(MINI_LAB_PATH)
	@kind get clusters | grep metal-control-plane > /dev/null || $(MAKE) mini-lab
	@$(eval include $(MINI_LAB_PATH)/.env)

.PHONY: mini-lab
mini-lab:
	make -C $(MINI_LAB_PATH)

.PHONY: integration-ansible-modules
integration-ansible-modules: prep
	docker run --rm -it \
		-v $(MINI_LAB_PATH):/mini-lab \
		-v $(PWD)/test/integration/ansible-modules:/integration:ro \
		-v $(PWD)/test/integration/ansible-modules/output:/output \
		-w /mini-lab \
		-e METAL_ANSIBLE_INVENTORY_CONFIG=/root/.ansible/roles/metal-ansible-modules/inventory/metal_config.example.yaml \
		--env-file $(MINI_LAB_PATH)/.env \
		--network host \
		metalstack/metal-deployment-base:$(DEPLOYMENT_BASE_IMAGE_TAG) /integration/integration.sh

.PHONY: integration-deployment
integration-deployment: prep
	docker run --rm -it \
		-v $(MINI_LAB_PATH):/mini-lab \
		-v $(PWD)/test/integration/deployment:/integration:ro \
		-v $(PWD)/test/integration/deployment/output:/output \
		-w /integration \
		--network host \
		metalstack/metal-deployment-base:$(DEPLOYMENT_BASE_IMAGE_TAG) /integration/integration.sh

.PHONY: render-junit
render-junit:
	docker run --rm -v $(PWD)/test/integration/ansible-modules/output:/results maxmiorim/junit-viewer > results.html
