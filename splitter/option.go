package splitter

import (
	"github.com/sirupsen/logrus"
)

type Option func(*Splitter)

func WithLogger(logger *logrus.Entry) Option {
	return func(s *Splitter) {
		s.logger = logger
	}
}

func WithSegmentTime(segmentTime int) Option {
	return func(s *Splitter) {
		s.segmentTime = segmentTime
	}
}

func WithOutputDir(output string) Option {
	return func(s *Splitter) {
		s.outputDir = output
	}
}
