.PHONY: run rerun swagger-init swagger-clean

run:
	go run cmd/app/main.go

rerun:
	pkill -f "go run cmd/app/main.go" || true
	go run cmd/app/main.go

build:
	go build -o app cmd/app/main.go

# Swagger commands
swagger-init:
	swag init -g cmd/app/main.go

swagger-clean:
	rm -rf docs/
	swag init -g cmd/app/main.go

# Install swagger CLI
swagger-install:
	go install github.com/swaggo/swag/cmd/swag@latest


.DEFAULT_GOAL := help
