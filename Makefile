.PHONY: build run test clean docker-build

BINARY_NAME=llm-proxy

build:
	go build -o bin/$(BINARY_NAME) ./cmd/server/main.go

run: build
	./bin/$(BINARY_NAME)

dev-backend:
	go run github.com/air-verse/air@latest -c .air.toml

dev-frontend:
	cd frontend && npm run dev

dev:
	make -j2 dev-backend dev-frontend

test:
	go test ./...

clean:
	rm -rf bin/
	rm -f data/llm-proxy.db

docker-build:
	docker build -t $(BINARY_NAME) .
