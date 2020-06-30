GOOS?=linux
GOARCH?=amd64

NAME=uploader
VERSION?=$$(git describe --abbrev=0)-$$(git rev-parse --abbrev-ref HEAD)-$$(git rev-parse --short HEAD)

REGISTRY_SERVER?=registry.videocoin.net
REGISTRY_PROJECT?=cloud

ENV?=dev

.PHONY: deploy

default: build

version:
	@echo ${VERSION}

build:
	GOOS=${GOOS} GOARCH=${GOARCH} \
		go build \
		    -mod vendor \
			-ldflags="-w -s -X main.Version=${VERSION}" \
			-o bin/${NAME} \
			./cmd/main.go

deps:
	GO111MODULE=on go mod vendor

lint: docker-lint

docker-lint:
	docker build -f Dockerfile.lint .

test-integration:
	go test -v -tags=integration ./...

docker-test-build:
	docker build -t tests -f Dockerfile.test .

docker-test-run:
	docker run --net=host -e "GDRIVE_KEY=${GDRIVE_KEY}" -e "AUTH_TOKEN_SECRET=${AUTH_TOKEN_SECRET}" tests make test-integration

docker-build:
	docker build -t ${REGISTRY_SERVER}/${REGISTRY_PROJECT}/${NAME}:${VERSION} -f Dockerfile .

docker-push:
	docker push ${REGISTRY_SERVER}/${REGISTRY_PROJECT}/${NAME}:${VERSION}

release: docker-build docker-push

deploy:
	cd deploy && helm upgrade -i --wait --timeout 30s --set image.tag="${VERSION}" -n console uploader ./helm