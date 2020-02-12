GOOS?=linux
GOARCH?=amd64

GCP_PROJECT=videocoin-network

NAME=uploader
VERSION=$$(git describe --abbrev=0)-$$(git rev-parse --abbrev-ref HEAD)-$$(git rev-parse --short HEAD)

DBM_MSQLURI=root:@tcp(127.0.0.1:3306)/videocoin?charset=utf8&parseTime=True&loc=Local
ENV?=dev

.PHONY: deploy

default: build

version:
	@echo ${VERSION}

build:
	GOOS=${GOOS} GOARCH=${GOARCH} \
		go build \
			-ldflags="-w -s -X main.Version=${VERSION}" \
			-o bin/${NAME} \
			./cmd/main.go

deps:
	GO111MODULE=on go mod vendor

lint:
	golangci-lint run -v

docker-lint:
	docker build -f Dockerfile.lint .

test-integration:
	go test -v -tags=integration ./...

docker-test-build:
	docker build -t tests -f Dockerfile.test .

docker-test-run:
	docker run --net=host -e "GDRIVE_KEY=${GDRIVE_KEY}" -e "AUTH_TOKEN_SECRET=${AUTH_TOKEN_SECRET}" tests make test-integration

docker-build:
	@ if [ "${CI}" = "" ]; then \
        make lint; \
    fi
	docker build -t gcr.io/${GCP_PROJECT}/${NAME}:${VERSION} -f Dockerfile .

docker-push:
	docker push gcr.io/${GCP_PROJECT}/${NAME}:${VERSION}

dbm-status:
	goose -dir migrations -table ${NAME} mysql "${DBM_MSQLURI}" status

dbm-up:
	goose -dir migrations -table ${NAME} mysql "${DBM_MSQLURI}" up

dbm-down:
	goose -dir migrations -table ${NAME} mysql "${DBM_MSQLURI}" down

release: docker-build docker-push

deploy:
	ENV=${ENV} deploy/deploy.sh