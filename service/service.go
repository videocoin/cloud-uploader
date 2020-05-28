package service

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	clientv1 "github.com/videocoin/cloud-api/client/v1"
	pstreamsv1 "github.com/videocoin/cloud-api/streams/private/v1"
	"github.com/videocoin/cloud-uploader/datastore"
	"github.com/videocoin/cloud-uploader/downloader"
	"github.com/videocoin/cloud-uploader/server"
	"github.com/videocoin/cloud-uploader/splitter"
)

type Service struct {
	cfg        *Config
	logger     *logrus.Entry
	server     *server.Server
	downloader *downloader.Downloader
	splitter   *splitter.Splitter
	ds         datastore.Datastore
	sc         *clientv1.ServiceClient
	stop       chan bool
}

func NewService(ctx context.Context, cfg *Config) (*Service, error) {
	sc, err := clientv1.NewServiceClientFromEnvconfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	splitter, err := splitter.NewSplitter(ctx, splitter.WithOutputDir(cfg.DownloadDir))
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
		server.WithSplitter(splitter),
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
		splitter:   splitter,
		ds:         ds,
		sc:         sc,
		stop:       make(chan bool, 1),
	}, nil
}

func (s *Service) dispatch() {
	go func() {
		for outputFile := range s.downloader.OutputCh {
			if outputFile != nil && s.splitter != nil {
				logger := s.logger.WithField("stream_id", outputFile.StreamID)
				logger.Info("recieved output file from downloader")
				go func() {
					logger.Info("sending file to splitter")
					s.splitter.InputCh <- &splitter.MediaFile{
						StreamID: outputFile.StreamID,
						Path:     outputFile.Path,
					}
				}()
			}
		}
	}()

	go func() {
		for mf := range s.splitter.OutputCh {
			logger := s.logger.WithField("stream_id", mf.StreamID).WithField("path", mf.Path)
			logger.Info("recieved output from splitter")

			ctx := ctxlogrus.ToContext(context.Background(), logger)

			if mf.Error != nil {
				logger.WithError(mf.Error).Info("failed to split")
				if s.sc != nil && s.sc.Streams != nil {
					s.stopStream(ctx, mf.StreamID)
				}
				continue
			}

			s.logger.
				WithField("stream_id", mf.StreamID).
				WithField("path", mf.Path).
				Info("file has been splitted")

			if s.sc != nil && s.sc.Streams != nil {
				logger.Info("stream publishing")

				streamReq := &pstreamsv1.StreamRequest{
					Id:       mf.StreamID,
					Duration: mf.Duration,
				}
				_, err := s.sc.Streams.Publish(context.Background(), streamReq)
				if err != nil {
					logger.WithError(err).Error("failed to publish stream")
					s.stopStream(ctx, mf.StreamID)
					continue
				}
			}
		}
	}()

	<-s.stop
}

func (s *Service) stopStream(ctx context.Context, streamID string) {
	logger := ctxlogrus.Extract(ctx)
	streamReq := &pstreamsv1.StreamRequest{Id: streamID}
	_, err := s.sc.Streams.Stop(context.Background(), streamReq)
	if err != nil {
		logger.WithError(err).Info("failed to stop stream")
	}
}

func (s *Service) Start(errCh chan error) {
	go func() {
		s.logger.WithField("addr", s.cfg.Addr).Info("starting api server")
		errCh <- s.server.Start()
	}()

	go func() {
		s.logger.WithField("dst", s.cfg.DownloadDir).Info("starting splitter")
		err := s.splitter.Start()
		if err != nil {
			errCh <- err
		}
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

	s.dispatch()
}

func (s *Service) Stop() error {
	err := s.server.Stop()
	if err != nil {
		return err
	}

	err = s.splitter.Stop()
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

	s.stop <- true

	return nil
}
