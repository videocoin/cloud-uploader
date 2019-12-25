module github.com/videocoin/cloud-uploader

go 1.12

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dwbuiten/go-mediainfo v0.0.0-20150630175133-91f51f40c56a // indirect
	github.com/google/uuid v1.0.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/labstack/echo v3.3.10+incompatible
	github.com/nareix/joy4 v0.0.0-20181022032202-3ddbc8f9d431 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/sirupsen/logrus v1.4.2
	github.com/videocoin/cloud-api v0.0.17
	github.com/videocoin/cloud-pkg v0.0.6
	google.golang.org/api v0.14.0
)

replace github.com/videocoin/cloud-pkg => ../cloud-pkg

replace github.com/videocoin/cloud-api => ../cloud-api
