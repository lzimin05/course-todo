start:
	docker compose up --build

clear: 
	docker compose down -v


swagger:
	swag init -g cmd/app/main.go -o docs