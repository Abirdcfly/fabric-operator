#
# Copyright contributors to the Hyperledger Fabric Operator project
#
# SPDX-License-Identifier: Apache-2.0
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at:
#
# 	  http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

IMAGE ?= hyperledgerk8s/fabric-operator
TAG ?= $(shell git rev-parse --short HEAD)
ARCH ?= $(shell go env GOARCH)
OS = $(shell go env GOOS)
SEMREV_LABEL ?= v1.0.0-$(shell git rev-parse --short HEAD)
BUILD_DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VER ?= 1.18.4

# For compatibility with legacy install-fabric.sh conventions, strip the
# leading semrev 'v' character when preparing dist and release artifacts.
VERSION=$(shell echo $(SEMREV_LABEL) | sed -e  's/^v\(.*\)/\1/')

DOCKER_IMAGE_REPO ?= ""

DOCKER_BUILD ?= docker build

BUILD_ARGS+=--build-arg BUILD_ID=$(VERSION)
BUILD_ARGS+=--build-arg BUILD_DATE=$(BUILD_DATE)
BUILD_ARGS+=--build-arg GO_VER=$(GO_VER)

.PHONY: build

build:
	mkdir -p bin && go build -o bin/operator

image: setup
	$(DOCKER_BUILD) -f Dockerfile $(BUILD_ARGS) -t $(IMAGE):$(TAG) .
	docker tag $(IMAGE):$(TAG) $(IMAGE):latest

image-multi-arch: setup
	@printf "[worker.oci]\n\
	  max-parallelism = 1" > /tmp/buildkitd.toml
	docker buildx create --use --config /tmp/buildkitd.toml
	docker buildx build -f Dockerfile $(BUILD_ARGS) -t $(IMAGE):$(TAG) --platform=linux/arm64,linux/amd64 . --push
	docker buildx build -f Dockerfile $(BUILD_ARGS) -t $(IMAGE):latest --platform=linux/arm64,linux/amd64 . --push

govendor:
	@go mod vendor

setup: govendor manifests bundle generate

login:
	docker login --username $(DOCKER_USERNAME) --password $(DOCKER_PASSWORD) $(DOCKER_IMAGE_REPO)

#######################################
#### part of autogenerate makefile ####
#######################################

# Current Operator version
VERSION ?= "1.0.0"
# Default bundle image tag
BUNDLE_IMG ?= controller-bundle:$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:crdVersions=v1"

# KIND cluster for local development, integration, and E2E testing
KIND_CLUSTER_NAME ?= fabric
KIND_KUBE_VERSION ?= v1.24.4								# Matches integ IKS cluster rev.  v1.23.4 is current
KIND_NODE_IMAGE   ?= kindest/node:$(KIND_KUBE_VERSION)

# Integration test parameters
INT_TEST_TIMEOUT  ?= 60m
INT_TEST_NAME     ?= *

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet manifests
	@scripts/run-unit-tests.sh

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

local:
	CLUSTERTYPE=K8S OPERATOR_NAMESPACE=default OPERATOR_LOCAL_MODE=true go run ./main.go

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Create a KIND K8s and Nginx ingress controller on :80 / :443
kind: kustomize
	kind create cluster --name $(KIND_CLUSTER_NAME) --config=integration/kind-config.yaml --image $(KIND_NODE_IMAGE)
	$(KUSTOMIZE) build config/ingress/kind | kubectl apply -f -

# Destroy the local KIND cluster
unkind:
	kind delete cluster --name $(KIND_CLUSTER_NAME)

# Run integration tests.  Target a specific test package by specifying INT_TEST in the make env.
# If INT_TEST is unspecified, run ALL tests (slow!!)
integration-tests:
	ginkgo -v -failFast -timeout $(INT_TEST_TIMEOUT) ./integration/$(INT_TEST_NAME)

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	@scripts/checks.sh

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

.PHONY: mocks
mocks: counterfeiter
	go generate ./...

counterfeiter:
ifeq (, $(shell which counterfeiter))
	@{ \
	set -e ;\
	COUNTERFEITER_TMP_DIR=$$(mktemp -d) ;\
	cd $$COUNTERFEITER_TMP_DIR ;\
	go mod init tmp ;\
	go install github.com/maxbrunsfeld/counterfeiter/v6 ;\
	rm -rf $$COUNTERFEITER_TMP_DIR ;\
	}
endif

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go install sigs.k8s.io/kustomize/kustomize/v4@v4.5.7 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

# Generate bundle manifests and metadata, then validate generated files.
bundle: manifests
	operator-sdk generate kustomize manifests -q
	kustomize build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

# Build the bundle image.
bundle-build:
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: opm
OPM = ./bin/opm
opm:
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.15.1/$(OS)-$(ARCH)-opm ;\
	chmod +x $(OPM) ;\
	}
else 
OPM = $(shell which opm)
endif
endif
BUNDLE_IMGS ?= $(BUNDLE_IMG) 
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION) ifneq ($(origin CATALOG_BASE_IMG), undefined) FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG) endif 
.PHONY: catalog-build
catalog-build: opm
	$(OPM) index add --container-tool docker --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

.PHONY: catalog-push
catalog-push: ## Push the catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)
