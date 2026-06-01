.PHONY: build stop log
build:
	docker-compose up -d --build
	
stop:
	docker stop $$(docker ps -a -q)
	docker rm $$(docker ps -a -q)

service ?= *

log:
	docker compose logs -f $(service)
	