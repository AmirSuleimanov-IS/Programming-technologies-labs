# Сборка: статический бинарник без CGO
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# Запуск: минимальный образ без компилятора и исходников
FROM alpine:3.22
RUN adduser -D -H app
WORKDIR /app
COPY --from=build /out/api /app/api
USER app
ENV APP_ENV=production
EXPOSE 8090
ENTRYPOINT ["/app/api"]
