.PHONY: dev
dev: .env
	@set -o allexport; . ./.env; set +o allexport; hivemind

.env:
	cp .env.example .env

.PHONY: prod
prod: build-container
	docker compose up --build

.PHONY: build-container
build-container:
	docker build -t rechenschaftspflicht .

.PHONY: check
check:
	cd src && go build ./...
	cd src && go vet ./...
	cd src && golangci-lint run ./...

.PHONY: fix
fix:
	cd src && go fmt ./...
	cd src && go fix ./...
