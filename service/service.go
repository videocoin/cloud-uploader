package service

import (
	"context"

	"github.com/videocoin/cloud-uploader/datastore"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	clientv1 "github.com/videocoin/cloud-api/client/v1"
	"github.com/videocoin/cloud-uploader/downloader"
	"github.com/videocoin/cloud-uploader/server"
)

type Service struct {
	cfg        *Config
	logger     *logrus.Entry
	server     *server.Server
	downloader *downloader.Downloader
	ds         datastore.Datastore
}

func NewService(ctx context.Context, cfg *Config) (*Service, error) {
	sc, err := clientv1.NewServiceClientFromEnvconfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	ds, err := datastore.NewDatastore(ctx, cfg.RedisURI)
	if err != nil {
		return nil, err
	}

	downloaderCtx := downloader.NewContextWithGDriveKey(ctx, cfg.GDriveKey)
	downloader, err := downloader.NewDownloader(
		downloaderCtx,
		cfg.DownloadDir,
		downloader.WithDatastore(ds),
	)
	if err != nil {
		return nil, err
	}

	serverOpts := []server.Option{
		server.WithAddr(cfg.Addr),
		server.WithAuthTokenSecret(cfg.AuthTokenSecret),
		server.WithDownloader(downloader),
		server.WithServiceClient(sc),
		server.WithDatastore(ds),
	}
	server, err := server.NewServer(ctx, serverOpts...)
	if err != nil {
		return nil, err
	}

	return &Service{
		cfg:        cfg,
		logger:     ctxlogrus.Extract(ctx),
		server:     server,
		downloader: downloader,
		ds:         ds,
	}, nil
}

func (s *Service) Start(errCh chan error) {
	go func() {
		s.logger.WithField("addr", s.cfg.Addr).Info("starting api server")
		errCh <- s.server.Start()
	}()

	go func() {
		s.logger.WithField("dst", s.cfg.DownloadDir).Info("starting downloader")
		err := s.downloader.Start()
		if err != nil {
			errCh <- err
		}
	}()

	go func() {
		s.logger.WithField("uri", s.cfg.RedisURI).Info("starting datastore")
		err := s.ds.Start()
		if err != nil {
			errCh <- err
		}
	}()
}

func (s *Service) Stop() error {
	err := s.server.Stop()
	if err != nil {
		return err
	}

	err = s.downloader.Stop()
	if err != nil {
		return err
	}

	err = s.ds.Stop()
	if err != nil {
		return err
	}

	return nil
}
