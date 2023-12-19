APP := gomq

.PHONY: dev
dev: clean
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ./app/$(APP)
	./app/$(APP) start

.PHONY: docker
docker: clean
	docker buildx build --platform linux/amd64 -t $(APP) .


.PHONY: clean
clean:
	go clean
	rm -f ./app/data/log/gomq.log