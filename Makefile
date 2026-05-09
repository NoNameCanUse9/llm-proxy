.PHONY: build run test clean docker-build doc

BINARY_NAME=llm-proxy

build-frontend:
	cd frontend && npm install && npm run build

build: build-frontend
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
	rm -f data/data.db

docker-build:
	docker build -t $(BINARY_NAME) .

doc:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go --parseDependency --parseInternal
