.PHONY: install-kev start-kev test-kev up down logs demo test

install-kev:
	chmod +x scripts/*.sh
	./scripts/install-kev.sh

start-kev:
	./scripts/start-kev.sh

test-kev:
	./scripts/test-kev.sh

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

demo:
	curl -fsS -X POST http://localhost:3080/api/demo/reset

test:
	cd backend && go test ./...
	cd frontend && npm test -- --run
