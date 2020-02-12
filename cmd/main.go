package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	splitterv1 "github.com/videocoin/cloud-api/splitter/v1"
	privatev1 "github.com/videocoin/cloud-api/streams/private/v1"
	"github.com/videocoin/cloud-pkg/grpcutil"
	"github.com/videocoin/cloud-pkg/logger"
	"github.com/videocoin/cloud-pkg/tracer"
	"github.com/videocoin/cloud-uploader/service"
)

var (
	ServiceName string = "uploader"
	Version     string = "dev"
)

func main() {
	logger.Init(ServiceName, Version)  //nolint

	log := logrus.NewEntry(logrus.New())
	log = logrus.WithFields(logrus.Fields{
		"service": ServiceName,
		"version": Version,
	})

	closer, err := tracer.NewTracer(ServiceName)
	if err != nil {
		log.Info(err.Error())
	} else {
		defer closer.Close()
	}

	config := &service.Config{
		Name:    ServiceName,
		Version: Version,
	}

	err = envconfig.Process(ServiceName, config)
	if err != nil {
		log.Fatal(err.Error())
	}

	config.Logger = log

	conn, err := grpcutil.Connect(config.StreamsRPCAddr, config.Logger.WithField("system", "streamscli"))
	if err != nil {
		log.Fatal(err.Error())
	}
	streams := privatev1.NewStreamsServiceClient(conn)

	conn, err = grpcutil.Connect(config.SplitterRPCAddr, config.Logger.WithField("system", "splittercli"))
	if err != nil {
		log.Fatal(err.Error())
	}
	splitter := splitterv1.NewSplitterServiceClient(conn)

	svc, err := service.NewService(config, streams, splitter)
	if err != nil {
		log.Fatal(err.Error())
	}

	signals := make(chan os.Signal, 1)
	exit := make(chan bool, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals

		log.Infof("recieved signal %s", sig)
		exit <- true
	}()

	log.Info("starting")
	go log.Fatal( svc.Start())

	<-exit

	log.Info("stopping")
	err = svc.Stop()
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("stopped")
}
