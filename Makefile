
# Image URL to use all building/pushing image targets
IMG ?= docker.io/jungler/controller:latest
# Kubernetes Certs dir
CERTDIR ?= /etc/kubernetes/pki
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
#CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build --load-restrictor=noneconfig/crd | kubectl apply -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build --load-restrictor=LoadRestrictionsNone config/default | kubectl create -f -

# Undeploy Kubeadm operator
undeploy:
	kustomize build --load-restrictor=LoadRestrictionsNone config/default | kubectl delete -f -

debug: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build --load-restrictor=LoadRestrictionsNone config/debug | kubectl create -f -

baremetal:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go
	#./manager --mode=manager --manager-pod=baremetal --manager-namespace=operator-system --agent-image=jungler/controller:latest --agent-metrics-rbac=false
	./manager --mode=manager --manager-pod=baremetal --manager-namespace=operator-system --agent-image=jungler/controller:latest

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}
	docker save ${IMG} -o controller.tar
	ctr -n k8s.io images delete ${IMG}
	ctr -n k8s.io images import controller.tar
	ctr -n k8s.io images ls | grep ${IMG}

docker-build-debug: test
	docker build -f Dockerfile.debug . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# Prepare New root-ca for ca-rotation
prepare_ca_rotation:
ifneq ($(wildcard ${CERTDIR}/ca.crt.old),)
	@echo "${CERTDIR}/ca.crt.old is already exist, pleace check."
else
	CERTDIR=${CERTDIR} hack/prepare-ca-rotation.sh
endif

# Clean failed/canceled ca-rotation
clean_ca_rotation:
ifeq ($(wildcard ${CERTDIR}/ca.crt.old),)
	@echo "${CERTDIR}/ca.crt.old is not exist, pleace check."
else
	CERTDIR=${CERTDIR} hack/clean-ca-rotation.sh
endif

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	#go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.1
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
