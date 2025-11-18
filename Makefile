SHELL := /bin/zsh

.PHONY: default dev stop

default: dev

DEV_APP_CMD := cd $(CURDIR)/app && nvm use 22 && wails dev
DEV_SERVER_CMD := cd $(CURDIR)/server && go run main.go
DEV_WEB_CMD := cd $(CURDIR)/web && nvm use 22 && pnpm dev

dev:
	@echo "Starting development environment..."
	@osascript -e 'tell application "Terminal" to do script "$(DEV_APP_CMD)"'
	@osascript -e 'tell application "Terminal" to do script "$(DEV_SERVER_CMD)"'
	@osascript -e 'tell application "Terminal" to do script "$(DEV_WEB_CMD)"'

stop:
	@echo "Stopping development processes..."
	@pkill -f "wails dev" || true
	@pkill -f "go run main.go" || true
	@pkill -f "pnpm dev" || true
	@echo "Development processes stopped."

