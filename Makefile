PACKAGES := $(shell go list ./...)
name := $(shell basename ${PWD})
DOCKER_USERNAME := "nzkiwikid"
TAG             := "latest"
DATA_VOLUME_NAME=sqlite-data


all: help

.PHONY: help
help: Makefile
	@echo
	@echo " Choose a make command to run"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## init: initialize project (make init module=github.com/user/project)
.PHONY: init
init:
	go mod init ${module}
	go install github.com/cosmtrek/air@latest
	asdf reshim golang

## vet: vet code
.PHONY: vet
vet:
	go vet $(PACKAGES)

## test: run unit tests
.PHONY: test
test:
	go test -race -cover $(PACKAGES)

## build: build a binary
.PHONY: build
build: test
	go build -o ./app -v

## docker-build: build project into a docker container image
.PHONY: docker-build
docker-build: test
	GOPROXY=direct docker buildx build -t ${name} .

.PHONY: docker-db-init
docker-db-init: 
	docker volume create --name=sqlite-data
	

## docker-push: push docker container to Docker Hub
.PHONY: docker-push
docker-push: docker-build
	docker tag ${name} ${DOCKER_USERNAME}/${name}:${TAG}
	docker push ${DOCKER_USERNAME}/${name}:${TAG}

## docker-run: run project in a container
.PHONY: docker-run
docker-run:
	docker run -it --rm -p 8122:8122 -v ${DATA_VOLUME_NAME}:/app/db ${name}

## start: build and run local project
.PHONY: start
start: build
	IS_DEV=true MQTT_URL=192.168.1.5 air

## css: build tailwindcss
.PHONY: css
css:
	tailwindcss -i css/input.css -o css/output.css --minify

## css-watch: watch build tailwindcss
.PHONY: css-watch
css-watch:
	tailwindcss -i css/input.css -o css/output.css --watch
