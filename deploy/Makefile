.PHONY: build stop log
build:
	docker-compose up -d --build
	
stop:
	docker-compose down -v

service ?= *

log:
	docker compose logs -f $(service)
	