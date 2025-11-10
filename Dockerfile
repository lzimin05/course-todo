# Этап 1: Сборка приложения
FROM golang:1.24.2 AS builder

WORKDIR /app

# Копируем только файлы, необходимые для загрузки зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальные файлы
COPY . .

# Создаем пустой .env файл на случай, если его нет
RUN touch .env

# Собираем все приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/app/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migrations/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o seed ./script/seed.go

# Этап 2: Финальный образ
FROM alpine:3.18

WORKDIR /app

# Копируем только необходимые артефакты
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY --from=builder /app/seed .
COPY --from=builder /app/db/migrations ./db/migrations
COPY --from=builder /app/.env .

EXPOSE 8080

# Запускаем миграции и приложение
CMD sh -c "./migrate && ./main"