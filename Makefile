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