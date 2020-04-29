package server

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	otecho "github.com/opentracing-contrib/echo"
	echologrus "github.com/plutov/echo-logrus"
	"github.com/sirupsen/logrus"
	clientv1 "github.com/videocoin/cloud-api/client/v1"
	"github.com/videocoin/cloud-uploader/datastore"
	"github.com/videocoin/cloud-uploader/downloader"
	"github.com/videocoin/cloud-uploader/splitter"
)

type Server struct {
	addr            string
	authTokenSecret string
	logger          *logrus.Entry
	e               *echo.Echo
	downloader      *downloader.Downloader
	splitter        *splitter.Splitter
	sc              *clientv1.ServiceClient
	ds              datastore.Datastore
}

func NewServer(ctx context.Context, opts ...Option) (*Server, error) {
	s := &Server{
		logger: ctxlogrus.Extract(ctx).WithField("system", "api"),
		e:      echo.New(),
	}

	echologrus.Logger = s.logger.Logger

	s.e.HideBanner = true
	s.e.HidePort = true
	s.e.DisableHTTP2 = true
	s.e.Logger = echologrus.GetEchoLogger()

	for _, o := range opts {
		o(s)
	}

	s.route()

	return s, nil
}

func (s *Server) route() {
	s.e.Use(otecho.Middleware("uploader"))
	s.e.Use(middleware.CORS())
	s.e.Use(echologrus.Hook())

	s.e.GET("/healthz", s.getHealth)

	r := s.e.Group("/api/v1/upload/")
	r.Use(middleware.JWT([]byte(s.authTokenSecret)))

	r.POST("local/:id", s.uploadFromFile)
	r.POST("url/:id", s.uploadFromURL)
	r.GET("url/:id", s.checkUploadFromURL)
}

func (s *Server) Start() error {
	return s.e.Start(s.addr)
}

func (s *Server) Stop() error {
	return s.e.Shutdown(context.Background())
}

func (s *Server) E() *echo.Echo {
	return s.e
}
