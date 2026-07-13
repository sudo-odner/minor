LIB ?= github.com/sudo-odner/minor-shared
SERVICES := $(shell find .. -mindepth 2 -name "go.mod" -exec dirname {} \;)

.PHONY: build stop log all update $(SERVICES)
	
build:
	sudo docker-compose up -d --build
	
stop:
	sudo docker-compose down -v

service ?= *

log:
	docker compose logs -f $(service)

all: update

update: $(SERVICES)

$(SERVICES):
	@echo "---------------------------------------------------"
	@echo "Updating $(LIB) to $(VERSION) in $@..."
	@cd $@ && go get $(LIB) && go mod tidy
	@echo "Done for $@"