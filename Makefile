CONFIG_SOURCE=config.yml
ENV_FILE=.env

create-env:
	@if [ ! -f .env ]; then \
		echo "Generating .env from config.yml..."; \
		python3 -c "import yaml; config = yaml.safe_load(open('config.yml')); open('.env', 'w').write('\n'.join([f'{k}={v}' for k, v in config.items()]))"; \
		echo ".env file generated!"; \
	else \
		echo ".env already exists. Skipping..."; \
	fi

swagger:
	swag init -g cmd/app/main.go -o docs

deps:
	@echo "Установка зависимостей..."
	go mod download
	go mod tidy

start:
	docker compose up --build

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