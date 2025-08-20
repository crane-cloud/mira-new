# File names
DOCKER_DEV_COMPOSE_FILE := docker-compose.yml
BINARY_NAME := mira
DEV_SERVICE := api
BUILDPACKS_DIR := buildpacks



help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build-image: ## Setup development environment
	@ ${INFO} "Setting up development environment"
	@ docker compose -f $(DOCKER_DEV_COMPOSE_FILE) build api imagebuilder
	@ ${INFO} "Image succesfully built"
	@ echo " "

swagger: build-image ## Generate Swagger API documentation in Docker
	@ ${INFO} "Generating Swagger documentation in Docker"
	@ docker compose -f $(DOCKER_DEV_COMPOSE_FILE) exec $(DEV_SERVICE) sh -c "swag init -g cmd/api/app.go -o docs/ || (go install github.com/swaggo/swag/cmd/swag@latest && swag init -g cmd/api/app.go -o docs/)"
	@ ${INFO} "Swagger documentation generated successfully"
	@ echo "📖 Access at: http://localhost:3000/api/docs/"

start:build-image ## Start development server
	@ ${INFO} "starting local development server"
	@ docker compose -f $(DOCKER_DEV_COMPOSE_FILE) up
	@ echo ""
	@ ${INFO} "Development environment started successfully!"
	@ echo "API Documentation: http://localhost:3000/api/docs/"
	@ echo "Health Check: http://localhost:3000/api/health"
	@ echo "NATS Monitoring: http://localhost:8222/varz"
	@ echo "Development Shell: make shell"
	@ echo ""


logs-builder: build ## View logs from image builder service only
	@ ${INFO} "Viewing image builder service logs"
	@ docker compose -f $(DOCKER_DEV_COMPOSE_FILE) logs -f imagebuilder

clean: ## Remove all project images and volumes
	@ ${INFO} "Cleaning your local environment"
	@ ${INFO} "Note: All ephemeral volumes will be destroyed"
	@ docker compose -f $(DOCKER_DEV_COMPOSE_FILE) down --rmi all
	@ ${INFO} "Clean complete"

stop: ## Stop all project images and volumes
	@ ${INFO} "Stoping your local development server"
	@ docker compose -f $(DOCKER_DEV_COMPOSE_FILE) down -v
	@ ${INFO} "Stop complete"

test: build-image ## Run all tests in Docker
	@ ${INFO} "Running tests in Docker"
	@ docker compose -f $(DOCKER_DEV_COMPOSE_FILE) exec $(DEV_SERVICE) sh -c "go test -v ./... && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html"
	@ ${INFO} "Tests completed"
	@ echo "Coverage report: coverage.html"



# Code Quality & Security (run in Docker)
lint: build-image ## Run linter (golangci-lint) in Docker
	@ ${INFO} "Running linter in Docker"
	@ docker compose -f $(DOCKER_DEV_COMPOSE_FILE) exec $(DEV_SERVICE) sh -c "golangci-lint run || (go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && golangci-lint run)"
	@ ${INFO} "Linting completed"

fmt: build-image ## Format Go code in Docker
	@ ${INFO} "Formatting Go code in Docker"
	@ docker compose -f $(DOCKER_DEV_COMPOSE_FILE) exec $(DEV_SERVICE) go fmt ./...
	@ ${INFO} "Code formatted"

vet: build-image ## Run go vet in Docker
	@ ${INFO} "Running go vet in Docker"
	@ docker compose -f $(DOCKER_DEV_COMPOSE_FILE) exec $(DEV_SERVICE) go vet ./...
	@ ${INFO} "Vet completed"

sec: build-image ## Run security scanner (gosec) in Docker
	@ ${INFO} "Running security scanner in Docker"
	@ docker compose -f $(DOCKER_DEV_COMPOSE_FILE) exec $(DEV_SERVICE) sh -c "gosec ./... || (go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest && gosec ./...)"
	@ ${INFO} "Security scan completed"

check: fmt vet lint test ## Run all code quality checks in Docker


shell: build-image ## Open shell in development container
	@ ${INFO} "Opening shell in development container"
	@ docker compose -f $(DOCKER_DEV_COMPOSE_FILE) exec $(DEV_SERVICE) /bin/bash

build-bp: ## Build the Mira Buildpacks
	@ ${INFO} "Building the Mira Buildpacks"
	@ ${INFO} "Building the Node.js Buildpack..."
	@ cd $(BUILDPACKS_DIR) && go build -o bin/build ./nodejs/cmd/build/main.go
	@ cd $(BUILDPACKS_DIR) && go build -o bin/detect ./nodejs/cmd/detect/main.go
	@ ${INFO} "Buildpacks built successfully"
	@ echo "Run with: ./$(BINARY_NAME) <command>"

build-base-images: ## Build the base images for the builder
	@ ${INFO} "Building base images for the builder"
	@ cd $(BUILDPACKS_DIR)/builder && ./build.sh
	@ ${INFO} "Base images built successfully"


create-builder: build-base-images ## Create builder
	@ ${INFO} "Creating Mira Builder..."
	@ cd $(BUILDPACKS_DIR)/builder && pack builder create cranecloudplatform/mira-builder:latest --config ./builder.toml
	@ ${INFO} "Builder created successfully"

# set default target
.DEFAULT_GOAL := help

# colors
YELLOW := $(shell tput -Txterm setaf 3)
NC := "\e[0m"

#shell Functions
INFO := @bash -c 'printf $(YELLOW); echo "===> $$1"; printf $(NC)' SOME_VALUE