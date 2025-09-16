ifneq (,$(wildcard ./.env))
	include .env
	export
endif

goose-up:
	goose up

goose-down:
	goose down

goose-create:
ifndef name
	${error Usage: make goose-create name=new_migration_name}
endif
	goose create $(name) sql