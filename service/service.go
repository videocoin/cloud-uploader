package service

import (
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	splitterv1 "github.com/videocoin/cloud-api/splitter/v1"
	privatev1 "github.com/videocoin/cloud-api/streams/private/v1"
	"gopkg.in/redis.v5"
)

type UploaderService struct {
	config       *Config
	logger       *logrus.Entry
	api          *echo.Echo
	cli          *redis.Client
	streams      privatev1.StreamsServiceClient
	splitter     splitterv1.SplitterServiceClient
	ProcessErrCh chan error
}

func NewService(
	config *Config,
	streams privatev1.StreamsServiceClient,
	splitter splitterv1.SplitterServiceClient,
) (*UploaderService, error) {
	api := echo.New()
	api.HideBanner = true
	api.HidePort = true
	api.DisableHTTP2 = true

	processErrCh := make(chan error, 10)

	opts, err := redis.ParseURL(config.RedisURI)
	if err != nil {
		return nil, err
	}

	opts.MaxRetries = 3
	opts.PoolSize = 10

	cli := redis.NewClient(opts)
	if err != nil {
		return nil, err
	}

	err = cli.Ping().Err()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(config.DownloadDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		mkdirErr := os.MkdirAll(config.DownloadDir, os.ModeDir)
		if mkdirErr != nil {
			return nil, err
		}
	}

	return &UploaderService{
		config:       config,
		logger:       config.Logger,
		api:          api,
		cli:          cli,
		streams:      streams,
		splitter:     splitter,
		ProcessErrCh: processErrCh,
	}, nil
}

func (s *UploaderService) Start(errCh chan error) {
	s.logger.Infof("starting api server on %s", s.config.Addr)

	s.route()

	go func() {
		errCh <- s.api.Start(s.config.Addr)
	}()

	go func() {
		for err := range s.ProcessErrCh {
			if err != nil {
				errCh <- err
			}
		}
	}()
}

func (s *UploaderService) Stop() error {
	return nil
}
func (s *UploaderService) route() {
	s.api.Use(middleware.CORS())
	s.api.GET("/healthz", s.getHealth)

	r := s.api.Group("/api/v1/upload/")
	r.Use(middleware.JWT([]byte(s.config.AuthTokenSecret)))
	r.POST("local/:id", s.uploadFromFile)
	r.POST("url/:id", s.uploadFromURL)
	r.GET("url/:id", s.checkUploadFromURL)
}
