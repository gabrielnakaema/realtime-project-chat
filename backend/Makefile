ifneq (,$(wildcard ./.env))
	include .env
	export
endif

.PHONY: dev dev-api dev-notification dev-websocket dev-chat dev-tasks dev-mcp dev-relay
.PHONY: generate-proto goose-up goose-down goose-create

AIR_CONFIG_DIR := dev/air

dev:
	goreman start

dev-api:
	air -c $(AIR_CONFIG_DIR)/api.toml

dev-notification:
	air -c $(AIR_CONFIG_DIR)/notification.toml

dev-websocket:
	air -c $(AIR_CONFIG_DIR)/websocket.toml

dev-chat:
	air -c $(AIR_CONFIG_DIR)/chat.toml

dev-tasks:
	air -c $(AIR_CONFIG_DIR)/tasks.toml

dev-mcp:
	air -c $(AIR_CONFIG_DIR)/mcp.toml

dev-relay:
	air -c $(AIR_CONFIG_DIR)/relay.toml

generate-proto:
	go run github.com/bufbuild/buf/cmd/buf@v1.57.2 generate

goose-up:
	goose up

goose-down:
	goose down

goose-create:
ifndef name
	${error Usage: make goose-create name=new_migration_name}
endif
	goose create $(name) sql
