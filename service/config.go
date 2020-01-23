package service

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	Name    string `envconfig:"-"`
	Version string `envconfig:"-"`

	Addr            string `required:"true" default:"0.0.0.0:8090" envconfig:"ADDR"`
	UsersRPCAddr    string `default:"0.0.0.0:5000" envconfig:"USERS_RPC_ADDR"`
	StreamsRPCAddr  string `required:"true" envconfig:"STREAMS_RPC_ADDR" default:"127.0.0.1:5102"`
	SplitterRPCAddr string `required:"true" envconfig:"SPLITTER_RPC_ADDR" default:"127.0.0.1:5103"`
	RedisURI        string `default:"redis://:@127.0.0.1:6379/1" envconfig:"REDISURI"`
	DownloadDir     string `required:"true" default:"/tmp" envconfig:"DOWNLOAD_DIR"`
	EnableCORS      bool   `default:"true"`
	GDriveKey       string `required:"true" envconfig:"GDRIVE_KEY"`

	AuthTokenSecret string `required:"true" envconfig:"AUTH_TOKEN_SECRET"`

	Logger *logrus.Entry `envconfig:"-"`
}
