.PHONY: all build build-frontend build-backend run clean dev-frontend dev-backend

all: build

# Build everything
build: build-frontend copy-dist build-backend

# Build frontend
build-frontend:
	@echo "Building frontend..."
	cd web && npm install && npm run build

# Copy frontend dist to internal/web
copy-dist:
	@echo "Copying frontend dist to internal/web..."
	@rm -rf internal/web/dist
	@cp -r web/dist internal/web/

# Build backend
build-backend:
	@echo "Building Go binary..."
	go build -o nccl-test-web cmd/server/main.go

# Run the application
run: build
	@echo "Starting server..."
	./nccl-test-web

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f nccl-test-web
	@rm -rf web/dist
	@rm -rf web/node_modules
	@rm -rf internal/web/dist
	@rm -rf data/iplist

# Development mode - run backend
dev-backend:
	go run cmd/server/main.go

# Development mode - run frontend
dev-frontend:
	cd web && npm run dev
