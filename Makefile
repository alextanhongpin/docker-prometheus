export
REVISION := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
VERSION := $(shell cat VERSION)


run:
	@go run main.go


up:
	@docker-compose up -d


down:
	@docker-compose down
