# This is the Go package to run a target against. Useful for running tests of one package, for example.
PACKAGE ?= ./...

# The docker image tag. Needs to be overwritten when building an official release to be pushed to the registry.
DOCKER_IMAGE_TAG ?= local

# The docker image repository.
DOCKER_IMAGE_REPOSITORY ?= backbone81/ctf-ui-operator

# The docker image containing repository and tag.
DOCKER_IMAGE ?= $(DOCKER_IMAGE_REPOSITORY):$(DOCKER_IMAGE_TAG)

# The log level to use when executing the operator locally.
LOG_LEVEL ?= 0

# We want to have our binaries in the bin subdirectory available. In addition we want them to have priority over
# binaries somewhere else on the system.
export PATH := $(CURDIR)/bin:$(PATH)

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: run
run: lint ## Run the operator on your host.
	go run ./cmd/ctf-ui-operator --enable-developer-mode --log-level $(LOG_LEVEL)

.PHONY: test
test: lint ## Run tests.
	ginkgo run -p --race --coverprofile cover.out --output-dir ./tmp $(PACKAGE)
	go tool cover -html=tmp/cover.out -o tmp/cover.html

.PHONY: test-e2e
test-e2e: docker-build kuttl/setup/setup.yaml ## Run end-to-end tests.
	cat kuttl/setup/setup.yaml
	kubectl kuttl test --config kuttl/kuttl-test.yaml

##@ Build

.PHONY: build
build: lint ## Build the operator binary.
	go build ./cmd/ctf-ui-operator

.PHONY: docker-build
docker-build: lint ## Build the operator docker image.
	docker build -t $(DOCKER_IMAGE) .

.PHONY: clean
clean: ## Remove temporary files.
	kind delete cluster
	chmod -R ug+w tmp
	rm -rf tmp
	rm -f ctf-ui-operator

##@ Deployment

.PHONY: init-local
init-local: ## Install the local kind cluster.
	kind create cluster --config=scripts/kind-config.yaml
	$(MAKE) install

.PHONY: install
install: lint ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	kubectl apply -f manifests/ctf-ui-operator-crd.yaml

.PHONY: uninstall
uninstall: ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f manifests/ctf-ui-operator-crd.yaml

V1ALPHA1_DEEPCOPY_FILE := api/v1alpha1/zz_generated.deepcopy.go
V1ALPHA1_TYPE_FILES := $(filter-out $(V1ALPHA1_DEEPCOPY_FILE), $(wildcard api/v1alpha1/*.go))
$(V1ALPHA1_DEEPCOPY_FILE): $(V1ALPHA1_TYPE_FILES)
	controller-gen object paths=./api/v1alpha1/...

V1ALPHA1_CRD_FILE := manifests/ctf-ui-operator-crd.yaml
$(V1ALPHA1_CRD_FILE): $(V1ALPHA1_TYPE_FILES)
	rm -f tmp/ui.ctf.backbone81_*.yaml
	controller-gen crd paths=./api/v1alpha1/... output:crd:artifacts:config=tmp
	cat tmp/ui.ctf.backbone81_*.yaml > manifests/ctf-ui-operator-crd.yaml

V1ALPHA1_CLUSTERROLE_FILE := manifests/ctf-ui-operator-clusterrole.yaml
V1ALPHA1_CONTROLLER_FILES := $(shell go list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}}{{"\n"}}{{end}}' ./internal/controller/...)
$(V1ALPHA1_CLUSTERROLE_FILE): $(V1ALPHA1_CONTROLLER_FILES)
	controller-gen rbac:roleName=ctf-ui-operator paths=./internal/controller/... output:rbac:artifacts:config=tmp
	mv tmp/role.yaml manifests/ctf-ui-operator-clusterrole.yaml

manifests/kustomization.yaml: $(V1ALPHA1_CRD_FILE) $(V1ALPHA1_CLUSTERROLE_FILE) $(filter-out manifests/kustomization.yaml, $(wildcard manifests/*.yaml))
	rm -f $@
	cd manifests && kustomize create --autodetect
	for f in manifests/*.yaml; do yq --prettyPrint --inplace "$$f"; done

kuttl/setup/setup.yaml: $(wildcard manifests/*.yaml)
	kustomize build manifests > $@

.PHONY: generate
generate: $(V1ALPHA1_DEEPCOPY_FILE) $(V1ALPHA1_CRD_FILE) $(V1ALPHA1_CLUSTERROLE_FILE) manifests/kustomization.yaml kuttl/setup/setup.yaml

.PHONY: prepare
prepare: generate
	go mod tidy
	go fmt $(PACKAGE)
	go vet $(PACKAGE)

.PHONY: lint
lint: prepare
	golangci-lint run --fix
