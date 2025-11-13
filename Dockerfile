# Build stage for frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /build/web
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# Build stage for backend
FROM golang:1.21-alpine AS backend-builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /build/web/dist ./internal/web/dist
RUN go build -o nccl-test-web cmd/server/main.go

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=backend-builder /build/nccl-test-web .
RUN mkdir -p /app/data
EXPOSE 8080
CMD ["./nccl-test-web"]
