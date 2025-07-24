include .env
export $(shell sed 's/=.*//' .env)


PAYLOAD := payload.json
FIRST_SECRET := $(shell echo $(HMAC_SECRET) | cut -d',' -f1)

up:
	docker-compose up --build

down:
	docker-compose down -v --remove-orphans

restart:
	docker-compose down -v --remove-orphans && docker-compose up --build

build:
	docker-compose build --no-cache


curl:
	@echo "🚀 Sending HMAC request..."
	@BODY="$$(jq -c . $(PAYLOAD))"; \
	SIG="$$(echo -n "$$BODY" | openssl dgst -sha256 -hmac "$(FIRST_SECRET)" | sed 's/^.*= //')"; \
	echo "🔏 Signature: $$SIG"; \
	curl -s -X POST http://localhost:8080/webhook \
	  -H "Content-Type: application/json" \
	  -H "X-Signature-256: $$SIG" \
	  -d "$$BODY"

curl-github:
	@echo "🐙 Sending GitHub-style HMAC request..."
	@BODY="$$(jq -c . $(PAYLOAD))"; \
	SIG="$$(echo -n "$$BODY" | openssl dgst -sha256 -hmac "$(FIRST_SECRET)" | sed 's/^.*= //')"; \
	echo "🔏 GitHub-style Signature: sha256=$$SIG"; \
	curl -s -X POST http://localhost:8080/webhook \
	  -H "Content-Type: application/json" \
	  -H "X-Hub-Signature-256: sha256=$$SIG" \
	  -d "$$BODY"
