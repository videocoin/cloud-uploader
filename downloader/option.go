package downloader

import (
	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-uploader/datastore"
)

type Option func(*Downloader)

func WithLogger(logger *logrus.Entry) Option {
	return func(d *Downloader) {
		d.logger = logger
	}
}

func WithDatastore(ds datastore.Datastore) Option {
	return func(d *Downloader) {
		d.ds = ds
	}
}
