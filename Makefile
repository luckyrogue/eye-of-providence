.PHONY: help doctor deploy deploy-api deploy-dashboard infra-up infra-down backend-ingest backend-auth backend-reports proto-gen agent-dev dashboard-dev clean

doctor:
	@./scripts/doctor.sh

deploy: deploy-api deploy-dashboard

deploy-api:
	@./infra/fly/deploy-api.sh

deploy-dashboard:
	@./infra/fly/deploy-dashboard.sh

help:
	@echo "Eye of Providence — dev targets"
	@echo "  make doctor             — проверить установленные зависимости"
	@echo "  make deploy             — задеплоить backend и dashboard на Fly"
	@echo "  make deploy-api         — задеплоить только backend"
	@echo "  make deploy-dashboard   — задеплоить только dashboard"
	@echo "  make infra-up           — поднять postgres/clickhouse/redis"
	@echo "  make infra-down         — погасить infra"
	@echo "  make backend-ingest     — запустить ingest service"
	@echo "  make backend-auth       — запустить auth service"
	@echo "  make backend-reports    — запустить reports (Gemini) service"
	@echo "  make proto-gen          — сгенерировать proto-код для Go и TS"
	@echo "  make agent-dev          — запустить Tauri agent в dev"
	@echo "  make dashboard-dev      — запустить web dashboard"

infra-up:
	docker compose -f infra/docker-compose.yml up -d

infra-down:
	docker compose -f infra/docker-compose.yml down

backend-ingest:
	cd backend && go run ./cmd/ingest

backend-auth:
	cd backend && go run ./cmd/auth

backend-reports:
	cd backend && go run ./cmd/reports

proto-gen:
	cd proto && buf generate

agent-dev:
	pnpm -F @eop/agent tauri dev

dashboard-dev:
	pnpm -F @eop/dashboard dev

clean:
	rm -rf node_modules */node_modules */dist agent/src-tauri/target backend/bin
