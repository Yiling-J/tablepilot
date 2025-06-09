# --- Frontend build stage ---
FROM node:23-alpine AS frontend-builder

WORKDIR /app/ui

COPY ui/package.json ui/pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install

COPY ui ./
RUN pnpm build


# --- Backend build stage ---
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

COPY --from=frontend-builder /app/ui/dist /app/ui/dist

RUN go build -o tablepilot main.go


# --- Final runtime image ---
FROM alpine:latest

WORKDIR /app

COPY --from=backend-builder /app/tablepilot /app/tablepilot

EXPOSE 8083

ENTRYPOINT ["./tablepilot", "serve"]
