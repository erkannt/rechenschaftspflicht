.PHONY: dev
dev:
	hivemind

.PHONY: prod
prod: build-container
	docker compose up

.PHONY: build-container
build-container:
	docker build -t rechenschaftspflicht .
