# Makefile para Order Service

.PHONY: help build run test clean docker-build docker-run migrate lint swagger

# Variables
SERVICE_NAME=order-service
BINARY_NAME=order-service
DOCKER_IMAGE=farmanexo/$(SERVICE_NAME):latest
ENV?=local
DB_URL?=postgresql://admin:admin@localhost:5432/order_db?sslmode=disable

help: ## Muestra esta ayuda
	@echo "Comandos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install: ## Instala las dependencias
	@echo "Instalando dependencias..."
	go mod download
	go mod tidy
	go install github.com/swaggo/swag/cmd/swag@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

build: ## Compila el binario
	@echo "Compilando $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) cmd/server/main.go

run: swagger ## Ejecuta el servicio
	@echo "Ejecutando $(SERVICE_NAME) en modo $(ENV)..."
	ENV=$(ENV) go run cmd/server/main.go

dev: swagger ## Ejecuta en modo local con auto-reload
	@echo "Ejecutando en modo local..."
	ENV=local go run cmd/server/main.go

swagger: ## Genera documentación Swagger
	@echo "Generando documentación Swagger..."
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
	@echo "Swagger generado en /docs"

test: ## Ejecuta los tests
	@echo "Ejecutando tests..."
	go test -v -race -cover ./...

test-coverage: ## Ejecuta tests con coverage
	@echo "Ejecutando tests con coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generado: coverage.html"

lint: ## Ejecuta el linter
	@echo "Ejecutando linter..."
	golangci-lint run ./...

clean: ## Limpia archivos generados
	@echo "Limpiando..."
	rm -rf bin/
	rm -rf docs/
	rm -f coverage.out coverage.html

# Docker commands
docker-build: ## Construye la imagen Docker
	@echo "Construyendo imagen Docker..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Ejecuta el contenedor Docker
	@echo "Ejecutando contenedor..."
	docker run -p 4011:4011 --env-file .env.$(ENV) $(DOCKER_IMAGE)

docker-push: ## Sube la imagen a Docker Hub
	@echo "Subiendo imagen..."
	docker push $(DOCKER_IMAGE)

# Database Migration commands
migrate-create: ## Crea una nueva migración (uso: make migrate-create NAME=nombre)
	@echo "Creando migración: $(NAME)"
	migrate create -ext sql -dir migrations -seq $(NAME)

migrate-up: ## Ejecuta todas las migraciones pendientes
	@echo "Ejecutando migraciones UP..."
	migrate -path migrations -database "$(DB_URL)" up

migrate-down: ## Revierte la última migración
	@echo "Revirtiendo última migración..."
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-down-all: ## Revierte TODAS las migraciones (PELIGRO)
	@echo "ADVERTENCIA: Revertiendo TODAS las migraciones..."
	migrate -path migrations -database "$(DB_URL)" down -all

migrate-force: ## Fuerza la versión de migración (uso: make migrate-force VERSION=1)
	@echo "Forzando versión de migración a: $(VERSION)"
	migrate -path migrations -database "$(DB_URL)" force $(VERSION)

migrate-version: ## Muestra la versión actual de la BD
	@echo "Versión actual de migraciones:"
	migrate -path migrations -database "$(DB_URL)" version

# Development helpers
format: ## Formatea el código
	@echo "Formateando código..."
	go fmt ./...
	goimports -w .

gen-mocks: ## Genera mocks para testing
	@echo "Generando mocks..."
	mockery --all --output=tests/mocks --case=underscore

.DEFAULT_GOAL := help
