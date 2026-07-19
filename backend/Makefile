ifneq (,$(wildcard ./.env))
	include .env
	export
endif

dev:
	goreman start

dev-api:
	air -c .air.toml

dev-notification:
	air -c .air.notification.toml

dev-websocket:
	air -c .air.websocket.toml

dev-chat:
	air -c .air.chat.toml

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
