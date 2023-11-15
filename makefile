SHELL := /bin/bash

# Variables
DOCKER_COMPOSE_DIR := infra
KIND_CLUSTER := starter-cluster
VERSION := 1.0

# ==============================================================================
# Install dependencies for MacOS

dev-setup-mac:
	brew update
	brew list kind || brew install kind
	brew list kubectl || brew install kubectl
	brew list kustomize || brew install kustomize
.PHONY: dev-setup-mac

# ==============================================================================
# Docker Compose Commands

compose-up:
	docker-compose -f $(DOCKER_COMPOSE_DIR)/docker-compose-minio.yaml -p fs up -d
	docker-compose -f $(DOCKER_COMPOSE_DIR)/docker-compose-postgres.yaml -p db up -d
.PHONY: compose-up

compose-down:
	docker-compose -f $(DOCKER_COMPOSE_DIR)/docker-compose-minio.yaml -p fs down -v --remove-orphans
	docker-compose -f $(DOCKER_COMPOSE_DIR)/docker-compose-postgres.yaml -p db down -v --remove-orphans
.PHONY: compose-down

# ==============================================================================
# Testing Commands

documented-tests-integration:
	gotestdox ./api/... --tags=integration
	gotestdox ./db/... --tags=integration
	gotestdox ./vault/... --tags=integration
	gotestdox ./service/... --tags=integration
.PHONY: documented-tests-integration

documented-tests:
	gotestdox ./api/...
	gotestdox ./db/...
	gotestdox ./vault/...
	gotestdox ./service/...
.PHONY: documented-tests

tests-summary:
	go test ./api -cover -json | tparse -all
	go test ./db -cover -json | tparse -all
	go test ./vault -cover -json | tparse -all
	go test ./service -cover -json | tparse -all
.PHONY: tests-summary

# ==============================================================================
# Run Commands

run-race:
	go run -race main.go
.PHONY: run-race

run:
	go run main.go
.PHONY: run

# ==============================================================================
# Docker Build

build-service:
	docker build \
		-f $(DOCKER_COMPOSE_DIR)/docker/dockerfile \
		-t service-evidences:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.
.PHONY: build-service

# ==============================================================================
# KIND Kubernetes Cluster Management

kind-create:
	kind create cluster \
		--name $(KIND_CLUSTER) \
		--image kindest/node-arm64:v1.23.4@sha256:0415a7bb1275c23c9315d97c5bfd7aecfdf8a8ea0562911f653711d947f7bec0 \
		--config $(DOCKER_COMPOSE_DIR)/k8s/kind-config.yaml
	kubectl config set-context --current --namespace=service-system
.PHONY: kind-create

tidy:
	go mod tidy
.PHONY: tidy

kind-up:
	docker start $(KIND_CLUSTER)-control-plane
.PHONY: kind-up

kind-down:
	kind delete cluster --name $(KIND_CLUSTER)
.PHONY: kind-down

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces
.PHONY: kind-status

kind-status-service:
	kubectl get pods -o wide --watch
.PHONY: kind-status-service

# ==============================================================================
# Code Generation and Migrations

generate-sqlc:
	cd db/query && sqlc generate
.PHONY: generate-sqlc

migrate-up:
	migrate -path db/migration -database 'postgres://postgres:postgres@localhost:5432/DER?sslmode=disable' up
.PHONY: migrate-up

migrate-down:
	migrate -path db/migration -database 'postgres://postgres:postgres@localhost:5432/DER?sslmode=disable' down
.PHONY: migrate-down

# ==============================================================================
# Default target
.DEFAULT_GOAL := run
