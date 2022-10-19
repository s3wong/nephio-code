.PHONY: fmt

GOPATH := $(shell go env GOPATH)
PROJECT_ID ?= $(shell gcloud config get-value project)
#IMG ?= gcr.io/${PROJECT_ID}/cnna-docker-repo/nf-deploy:latest
IMG ?= us-central1-docker.pkg.dev/${PROJECT_ID}/cnna-docker-repo/nf-deploy:latest

export CGO_ENABLED=0
all: build

build:
	go build -o /tmp/nfdeploy ./...

#docker-release: docker-build docker-push

#docker-build: build
#	docker build . -t ${IMG}

#docker-push:
#	docker push ${IMG}
