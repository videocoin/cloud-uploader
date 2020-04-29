package server

import (
	"github.com/sirupsen/logrus"
	clientv1 "github.com/videocoin/cloud-api/client/v1"
	"github.com/videocoin/cloud-uploader/datastore"
	"github.com/videocoin/cloud-uploader/downloader"
	"github.com/videocoin/cloud-uploader/splitter"
)

type Option func(*Server)

func WithAddr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

func WithLogger(logger *logrus.Entry) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

func WithDownloader(d *downloader.Downloader) Option {
	return func(s *Server) {
		s.downloader = d
	}
}

func WithAuthTokenSecret(secret string) Option {
	return func(s *Server) {
		s.authTokenSecret = secret
	}
}

func WithServiceClient(sc *clientv1.ServiceClient) Option {
	return func(s *Server) {
		s.sc = sc
	}
}

func WithDatastore(ds datastore.Datastore) Option {
	return func(s *Server) {
		s.ds = ds
	}
}

func WithSplitter(splitter *splitter.Splitter) Option {
	return func(s *Server) {
		s.splitter = splitter
	}
}
