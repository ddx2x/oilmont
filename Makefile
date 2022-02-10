# the makefile describe
REPO=oilmont
VERSION="${VERSION:-"${COMMIT_ID:0:8}"}"
CRD_OPTIONS ?= "crd:trivialVersions=true"

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	curl -OL https://github.com/ddx2x/controller-tools/archive/v0.4.1.tar.gz && tar -zxvf v0.4.1.tar.gz && cd controller-tools-0.4.1 ;\
	cd ./cmd/controller-gen && go install && cd ../helpgen && go install && cd ../type-scaffold && go install ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# Just install controller-gen tools set
install-tools: controller-gen
	@echo "install controller-gen done"

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=deploy/crds > /dev/null 2>&1 &
	  
dep:
	go mod vendor

build: dep
	go build ./cmd/...


gen-dockerfile: 
	go run ./build/main.go

docker:
	@sh start.sh build

run:
	@sh start.sh run