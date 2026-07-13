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

goose-up:
	goose up

goose-down:
	goose down

goose-create:
ifndef name
	${error Usage: make goose-create name=new_migration_name}
endif
	goose create $(name) sql
