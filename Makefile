DIR = app
APP = $(shell grep 'Use' ./main.go | awk -F '"' '{print $$2}')
VERSION = $(shell grep 'Version' ./main.go | awk -F '"' '{print $$2}')

.PHONY: run build clean docker

run: clean
	@go build -o ./$(DIR)/$(APP) .
	GOMQ_ENV=dev ./$(DIR)/$(APP) start


build: clean
	@./build.sh -d $(DIR) -a $(APP) -v $(VERSION)


docker: clean
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ./$(DIR)/$(APP)-amd64 .
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o ./$(DIR)/$(APP)-arm64 .
	#docker run --privileged --rm tonistiigi/binfmt --install all
	-docker buildx create --use --name gomq
	docker buildx build --platform linux/amd64,linux/arm64 --build-arg APP=$(APP) -t laomar/gomq:$(VERSION) . --push
	@rm -f ./$(DIR)/$(APP)-amd64 ./$(DIR)/$(APP)-arm64

clean:
	@go clean
	@rm -rf ./$(DIR)/data/*
	@rm -rf ./$(DIR)/$(APP)*
