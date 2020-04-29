module github.com/videocoin/cloud-uploader

go 1.12

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gogo/protobuf v1.3.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/labstack/echo/v4 v4.1.10
	github.com/opentracing-contrib/echo v0.0.0-20190807091611-5fe2e1308f06
	github.com/opentracing/opentracing-go v1.1.0
	github.com/plutov/echo-logrus v1.0.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/vansante/go-ffprobe v1.1.0
	github.com/videocoin/cloud-api v0.0.17
	github.com/videocoin/cloud-pkg v0.0.6
	google.golang.org/api v0.14.0
	google.golang.org/grpc v1.21.1
	gopkg.in/redis.v5 v5.2.9
)

replace github.com/videocoin/cloud-pkg => ../cloud-pkg

replace github.com/videocoin/cloud-api => ../cloud-api
