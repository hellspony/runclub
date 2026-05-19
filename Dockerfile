# Stage 1: Build Vue frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web/admin
COPY web/admin/package.json web/admin/package-lock.json* ./
RUN npm ci --ignore-scripts
COPY web/admin/ .
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/web/admin/dist ./web/admin/dist
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/runclub ./cmd/server

# Stage 3: Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata sqlite-libs
WORKDIR /app
COPY --from=backend-builder /bin/runclub /bin/runclub
COPY --from=backend-builder /app/web/admin/dist /app/web/admin/dist
COPY config.example.yaml /app/config.yaml
EXPOSE 8080
VOLUME /data
CMD ["/bin/runclub"]
