REPO_ROOT = $(shell git rev-parse --show-toplevel)
GENERATE_SCRIPT=$(REPO_ROOT)/npm/deploy/certs/generate_certs.sh

.PHONY: generate

generate: $(GENERATE_SCRIPT) ## Generate mock clients
	$(GENERATE_SCRIPT)
