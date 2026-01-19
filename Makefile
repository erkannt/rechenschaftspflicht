.PHONY: dev
dev: .env
	@set -a; . ./.env; set +a; hivemind

.env:
	cp .env.example .env

.PHONY: prod
prod: build-container
	docker compose up

.PHONY: build-container
build-container:
	docker build -t rechenschaftspflicht .
