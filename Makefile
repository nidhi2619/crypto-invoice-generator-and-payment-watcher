.PHONY: run build

run:
	cd backend && go run cmd/api/main.go

build:
	cd backend && go build -o ../bin/api cmd/api/main.go

# Frontend commands
run-frontend:
	cd frontend && npm run dev
