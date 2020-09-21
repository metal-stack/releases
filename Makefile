ifeq ($(CI),true)
MINI_LAB_PATH := $(PWD)/mini-lab
else
MINI_LAB_PATH := $(PWD)/../mini-lab
endif

.PHONY: prep
prep:
	./test/integration/prep.sh $(MINI_LAB_PATH)

.PHONY: integration-ansible-modules
integration-ansible-modules: prep
	# TODO: base image needs to be taken dynamically
	make -C $(MINI_LAB_PATH)
	docker run --rm -it \
		-v $(MINI_LAB_PATH):/mini-lab \
		-v $(PWD)/test/integration/ansible-modules:/integration:ro \
		-v $(PWD)/test/integration/ansible-modules/output:/output \
		-w /mini-lab \
		-e METAL_ANSIBLE_INVENTORY_CONFIG=/root/.ansible/roles/metal-ansible-modules/inventory/metal_config.example.yaml \
		--env-file $(MINI_LAB_PATH)/.env \
		--network host \
		metalstack/metal-deployment-base:v0.0.6 /integration/integration.sh

.PHONY: integration-mini-lab
integration-mini-lab: prep
	cd $(MINI_LAB_PATH) && ./test.sh
