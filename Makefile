VERSION?=$$(cat VERSION)
BINARY?=octicketing

REGION?=eu-west-1
ACCOUNT_NUM=$(shell aws sts get-caller-identity --output text --query 'Account')

IMAGE?=octicketing
IMAGE_TAG?=$(VERSION)
IMAGE_REPOSITORY?=$(ACCOUNT_NUM).dkr.ecr.$(REGION).amazonaws.com

# all: gencerts-deploy deploy  build-image push clean

.PHONY: codebuild-local
codebuild-local: ## runc codebuild spec
	codebuild_build.sh -i aws/codebuild/standard:4.0 -a /Users/duektsah/go/src/github.com/tsahiduek/mysql-app/artifacts -s /Users/duektsah/go/src/github.com/tsahiduek/mysql-app -c

.PHONY: run-local
run-local: ## Build and run - localy
	rm -rf ./dist 
	mkdir -p ./dist/local
	cp .env ./dist/local/
	cp -r form dist/local/form/
	 go build -o dist/local/$(BINARY)
	DB_ENGINE=SQLITE  ./dist/local/$(BINARY) 

.PHONY: run-local-docker
run-local-docker: ## build linux binary and docker image
	docker run -d  --name octicketing -p 8080:8080  $(ACCOUNT_NUM).dkr.ecr.$(REGION).amazonaws.com/$(IMAGE):$(IMAGE_TAG)

.PHONY: pre-build
pre-build: ## pre build - get all dependencies
	go get ./...

.PHONY: build-image
build-image: ## Build binary and docker image
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./dist/linux/$(BINARY)
	docker build -t $(IMAGE_REPOSITORY)/$(IMAGE):$(IMAGE_TAG) .


.PHONY: login-ecr
login-ecr: ## Login ecr using the region and account number configured on the machine
	aws ecr get-login-password --region $(REGION) | docker login --username AWS --password-stdin $(ACCOUNT_NUM).dkr.ecr.$(REGION).amazonaws.com


.PHONY: push-ecr
push-ecr: ## post build-image step to tag and push version + latest images
	docker tag $(ACCOUNT_NUM).dkr.ecr.$(REGION).amazonaws.com/$(IMAGE):$(IMAGE_TAG) $(ACCOUNT_NUM).dkr.ecr.$(REGION).amazonaws.com/$(IMAGE):latest 
	docker push $(ACCOUNT_NUM).dkr.ecr.$(REGION).amazonaws.com/$(IMAGE):$(IMAGE_TAG)
	docker push $(ACCOUNT_NUM).dkr.ecr.$(REGION).amazonaws.com/$(IMAGE):latest


.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)