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

hmac:
	@echo "Generating HMAC signature..."
	@BODY="$$(jq -c . $(PAYLOAD))"; \
	SIG="$$(echo -n "$$BODY" | openssl dgst -sha256 -hmac "$(FIRST_SECRET)" | sed 's/^.*= //')"; \
	echo "Payload: $$BODY"; \
	echo "HMAC Signature: $$SIG"

curl-hmac:
	@echo "Sending HMAC-authenticated request..."
	@BODY="$$(jq -c . $(PAYLOAD))"; \
	SIG="$$(echo -n "$$BODY" | openssl dgst -sha256 -hmac "$(FIRST_SECRET)" | sed 's/^.*= //')"; \
	echo "Using Signature: $$SIG"; \
	curl -v -X POST http://localhost:8080/webhook \
	  -H "Content-Type: application/json" \
	  -H "X-Signature-256: $$SIG" \
	  -d "$$BODY"


curl-github:
	@echo "Sending GitHub-style HMAC request..."
	@BODY="$$(jq -c . $(PAYLOAD))"; \
	SIG="$$(echo -n "$$BODY" | openssl dgst -sha256 -hmac "$(FIRST_SECRET)" | sed 's/^.*= //')"; \
	echo "GitHub-style Signature: sha256=$$SIG"; \
	curl -s -X POST http://localhost:8080/webhook \
	  -H "Content-Type: application/json" \
	  -H "X-Hub-Signature-256: sha256=$$SIG" \
	  -d "$$BODY"

jwt:
	@echo "Generating JWT with shared secret..."
	go run ./scripts/gen_jwt.go

curl-jwt:
	@echo "Sending JWT request..."
	@TOKEN=$$(go run ./scripts/gen_jwt.go); \
	curl -v -X POST http://localhost:8080/secure-webhook \
	  -H "Authorization: Bearer $$TOKEN" \
	  -H "Content-Type: application/json" \
	  -d "$$(jq -c . $(PAYLOAD))"
