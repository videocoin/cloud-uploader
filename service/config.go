package service

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	Name    string `envconfig:"-"`
	Version string `envconfig:"-"`

	Addr         string `required:"true" default:"0.0.0.0:8090" envconfig:"ADDR"`
	UsersRPCAddr string `default:"0.0.0.0:5000" envconfig:"USERS_RPC_ADDR"`
	DownloadDir  string `required:"true" default:"/tmp"`
	EnableCORS   bool   `default:"true"`
	GDriveKey    string `required:"true" envconfig:"GDRIVE_KEY"`

	AuthTokenSecret string `required:"true" envconfig:"AUTH_TOKEN_SECRET"`

	Logger *logrus.Entry `envconfig:"-"`
}
