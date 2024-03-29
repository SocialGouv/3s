all: build

.PHONY: build
build: preflight
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -mod=vendor -o 3s .

.PHONY: preflight
preflight:
	go mod vendor
	go fmt github.com/SocialGouv/3s

update:
	go mod tidy
	go mod vendor

docker-build:
	docker build . -t ghcr.io/socialgouv/3s

docker-push:
	docker push ghcr.io/socialgouv/3s

docker: docker-build docker-push