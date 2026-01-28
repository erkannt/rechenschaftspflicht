.PHONY: dev
dev: .env
	@set -o allexport; . ./.env; set +o allexport; hivemind

.env:
	cp .env.example .env

.PHONY: prod
prod: build-container
	docker compose up

.PHONY: build-container
build-container:
	docker build -t rechenschaftspflicht .

.PHONY: check
check:
	cd src && go vet ./...
	cd src && golangci-lint run ./...
