APP := gomq

.PHONY: dev
dev: clean
	go build -o ./app/$(APP)
	./app/$(APP) start

.PHONY: docker
docker: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./app/$(APP)
	#docker buildx build --platform linux/amd64 -t $(APP) .


.PHONY: clean
clean:
	go clean
	rm -f ./app/data/log/gomq.log