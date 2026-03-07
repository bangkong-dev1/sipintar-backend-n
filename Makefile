.PHONY: run rerun

run:
	go run cmd/app/main.go

rerun:
	pkill -f "go run cmd/app/main.go" || true
	go run cmd/app/main.go



.DEFAULT_GOAL := help
