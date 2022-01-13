.PHONY: build run

default: build

build:
	docker build --build-arg ARCH=amd64 -t militarpancho1/remadperbot:v1 .
run: build
	docker run --env-file .env militarpancho1/remadperbot:v1
