SHELL := /bin/bash

# ==============================================================================
# Install dependencies for MacOS

dev-setup-mac:
	brew update
	brew list kind || brew install kind
	brew list kubectl || brew install kubectl
	brew list kustomize || brew install kustomize

# ==============================================================================
# -d detached mode
# -p name of the project

docker-compose-testing:
	docker-compose -f infra/docker-compose-minio.yaml -p fs up -d
	docker-compose -f infra/docker-compose-postgres.yaml -p db up -d

docker-compose-teardown:
	docker-compose -f infra/docker-compose-minio.yaml -p fs down -v --remove-orphans
	docker-compose -f infra/docker-compose-postgres.yaml -p db down -v --remove-orphans

# ==============================================================================
documented-tests-integration:
	gotestdox ./api/... --tags=integration
	gotestdox ./db/... --tags=integration
	gotestdox ./vault/... --tags=integration
	gotestdox ./service/... --tags=integration

# ==============================================================================
documented-tests:
	gotestdox ./api/...
	gotestdox ./db/...
	gotestdox ./vault/...
	gotestdox ./service/...

# ==============================================================================
tests-summary:
	go test ./api -cover -json | tparse -all
	go test ./db -cover -json | tparse -all
	go test ./vault -cover -json | tparse -all
	go test ./service -cover -json | tparse -all

# ==============================================================================
# Run the app locally with race conditions
run-race:
	go run -race main.go

# ======================================================================
# ==============================================================================
# Run the app locally
run:
	go run main.go

# ======================================================================
VERSION := 1.0

all: service

service:
	docker build \
		-f infra/docker/dockerfile \
		-t service-evidences:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.
# ========================================================================
# Running from within k8s/kind

KIND_CLUSTER := starter-cluster

kind-create:
	kind create cluster \
		--name $(KIND_CLUSTER) \
		--image kindest/node-arm64:v1.23.4@sha256:0415a7bb1275c23c9315d97c5bfd7aecfdf8a8ea0562911f653711d947f7bec0 \
		--config infra/k8s/kind-config.yaml
	kubectl config set-context --current --namespace=service-system

tidy:
	go mod tidy

kind-up:
	docker start $(KIND_CLUSTER)-control-plane

kind-down:
	kind delete cluster --name $(KIND_CLUSTER)

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

kind-status-service:
	kubectl get pods -o wide --watch
# ========================================================================
# Running sqlc to generate code
generate-sqlc:
	cd db/query && sqlc generate

# ========================================================================
# Running migrations for local postgres
migrate-up:
	migrate -path db/migration -database 'postgres://postgres:postgres@localhost:5432/DER?sslmode=disable' up
migrate-down:
	migrate -path db/migration -database 'postgres://postgres:postgres@localhost:5432/DER?sslmode=disable' down


