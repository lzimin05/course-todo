CONFIG_SOURCE=config.yml
ENV_FILE=.env

create-env:
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "Generating $(ENV_FILE) from $(CONFIG_SOURCE)..."; \
		python3 -c "import yaml; config = yaml.safe_load(open('$(CONFIG_SOURCE)')); open('$(ENV_FILE)', 'w').write('\n'.join([f'{k}={v}' for k, v in config.items()]))"; \
		echo "$(ENV_FILE) file generated!"; \
	else \
		echo "$(ENV_FILE) already exists. Skipping..."; \
	fi

swagger:
	swag init -g cmd/app/main.go -o docs

deps:
	@echo "Установка зависимостей..."
	go mod download
	go mod tidy

seed:
	@echo "Заполнение базы данных тестовыми данными..."
	@echo "Запуск базы данных..."
	docker compose up -d db auth_redis
	@echo "Ждем инициализации БД..."
	@sleep 3
	@echo "Запуск скрипта заполнения..."
	docker compose run --rm todo-app ./seed
	@echo "✅ База данных заполнена тестовыми данными!"

stop:
	docker compose stop

start:
	docker compose up --build

start-background:
	docker compose up --build -d

clear:
	@echo "Полная очистка приложения и БД..."
	docker compose down -v
	docker compose up -d db auth_redis
	@echo "Ждем инициализации БД..."
	@sleep 2
	docker compose run --rm todo-app ./migrate down
	docker compose run --rm todo-app ./migrate up
	docker compose down
	@echo "✅ Очистка завершена"

generate-mocks:
	@echo "Генерация моков через go generate..."
	go generate ./...

test:
	go test -v ./...

# Запуск тестов с покрытием кода для internal пакетов, исключая моки
test-coverage:
	@echo "Очистка кэша и старых файлов покрытия..."
	@rm -f coverage.out coverage_filtered.out
	go clean -testcache
	@echo "Запуск тестов с покрытием кода для internal пакетов..."
	go test -v -coverprofile=coverage.out ./internal/...
	@echo "Исключаем моки и сгенерированные файлы из покрытия..."
	@grep -v -E "(mock.*\.go|docs\.go|fill\.go|generate\.go)" coverage.out > coverage_filtered.out 2>/dev/null || true
	@echo "Результаты покрытия для internal пакетов:"
	go tool cover -func=coverage_filtered.out | grep -v "mock" | tail -1; 
	
# Создание HTML отчета о покрытии
test-coverage-html: test-coverage
	@echo "Создание HTML отчета..."
	go tool cover -html=coverage_filtered.out -o coverage.html
	@echo "HTML отчет создан: coverage.html"