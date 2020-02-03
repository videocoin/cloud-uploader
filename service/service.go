package service

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	splitterv1 "github.com/videocoin/cloud-api/splitter/v1"
	privatev1 "github.com/videocoin/cloud-api/streams/private/v1"
	"github.com/videocoin/cloud-pkg/grpcutil"
	"gopkg.in/redis.v5"
	"net/http"
	"os"
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

	conn, err := grpcutil.Connect(config.StreamsRPCAddr, config.Logger.WithField("system", "streamscli"))
	if err != nil {
		return nil, err
	}
	streams := privatev1.NewStreamsServiceClient(conn)

	conn, err = grpcutil.Connect(config.SplitterRPCAddr, config.Logger.WithField("system", "splittercli"))
	if err != nil {
		return nil, err
	}
	splitter := splitterv1.NewSplitterServiceClient(conn)

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

func (s *UploaderService) Start() error {
	s.logger.Infof("starting api server on %s", s.config.Addr)

	s.route()

	go s.api.Start(s.config.Addr)
	go func() {
		for err := range s.ProcessErrCh {
			if err != nil {
				s.logger.Error(err)
			}
		}
	}()

	return nil
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

func (s *UploaderService) health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"alive": "OK"})
}
