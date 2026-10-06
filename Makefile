.PHONY: build dev frontend backend test vet clean run docker-build docker-up docker-down docker-logs docker-ps deploy

build: frontend
	mkdir -p build
	CGO_ENABLED=0 go build -o ./build/runbook ./cmd/runbook

dev:
	@echo "Run the frontend separately with: cd web && npm run dev"
	go run ./cmd/runbook

frontend:
	cd web && npm ci && npm run build
	rm -rf cmd/runbook/web/dist/*
	cp -a web/dist/. cmd/runbook/web/dist/

backend:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf build web/dist/*
	: > web/dist/.gitkeep
	rm -rf cmd/runbook/web/dist/*
	: > cmd/runbook/web/dist/.gitkeep

run: build
	./build/runbook

docker-build:
	docker compose build

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f runbook

docker-ps:
	docker compose ps

# Deploy/refresh on the server: rebuild image and (re)start detached.
deploy: docker-up
