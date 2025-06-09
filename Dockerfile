FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN npm install -g pnpm

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o tablepilot main.go

RUN cd ui && pnpm install && pnpm build

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/tablepilot /app/tablepilot
COPY --from=builder /app/ui/dist /app/ui/dist
COPY --from=builder /app/config /app/config

EXPOSE 8083

ENTRYPOINT ["tablepilot", "serve"]