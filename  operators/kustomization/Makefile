KUSTOMIZE ?= kustomize
KUBECTL ?= kubectl

DEV_K_DIR := config/overlays/dev
PROD_K_DIR := config/overlays/prod

.PHONY: deploy-dev
deploy-dev:
	$(KUSTOMIZE) build $(DEV_K_DIR) | $(KUBECTL) apply -f -

.PHONY: deploy-prod
deploy-prod:
	$(KUSTOMIZE) build $(PROD_K_DIR) | $(KUBECTL) apply -f -

.PHONY: deploy-all
deploy-all: deploy-dev deploy-prod

.PHONY: delete-dev
delete-dev:
	$(KUSTOMIZE) build $(DEV_K_DIR) | $(KUBECTL) delete -f -

.PHONY: delete-prod
delete-prod:
	$(KUSTOMIZE) build $(PROD_K_DIR) | $(KUBECTL) delete -f -
