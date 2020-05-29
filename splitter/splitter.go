package splitter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"github.com/vansante/go-ffprobe"
)

type Splitter struct {
	logger      *logrus.Entry
	segmentTime int
	outputDir   string
	InputCh     chan *MediaFile
	OutputCh    chan *MediaFile
}

func NewSplitter(ctx context.Context, opts ...Option) (*Splitter, error) {
	s := &Splitter{
		logger:      ctxlogrus.Extract(ctx).WithField("system", "splitter"),
		segmentTime: 30,
		outputDir:   "/tmp",
		InputCh:     make(chan *MediaFile, 1),
		OutputCh:    make(chan *MediaFile, 1),
	}

	for _, o := range opts {
		o(s)
	}

	return s, nil
}

func (s *Splitter) dispatch() {
	for f := range s.InputCh {
		go func(f *MediaFile) {
			if f == nil {
				return
			}

			logger := s.logger.WithField("path", f.Path).WithField("stream_id", f.StreamID)
			logger.Info("splitting")

			ctx := ctxlogrus.ToContext(context.Background(), logger)
			err := s.Split(ctx, f)
			if err != nil {
				f.Error = err
			}

			if f.Error == nil {
				logger.Info("splitting has been completed")
			}

			go func() {
				s.OutputCh <- f
			}()
		}(f)
	}
}

func (s *Splitter) Start() error {
	_, err := os.Stat(s.outputDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		mkdirErr := os.MkdirAll(s.outputDir, 0777)
		if mkdirErr != nil {
			return mkdirErr
		}

		return err
	}

	s.dispatch()

	return nil
}

func (s *Splitter) Stop() error {
	close(s.InputCh)
	close(s.OutputCh)
	return nil
}

func (s *Splitter) Split(ctx context.Context, f *MediaFile) error {
	logger := ctxlogrus.Extract(ctx)

	mediadata, err := ffprobe.GetProbeData(f.Path, time.Second*10)
	if err != nil {
		return fmt.Errorf("failed to get probe data: %s", err)
	}

	stream := mediadata.GetFirstVideoStream()
	if stream == nil {
		return fmt.Errorf("failed to get stream data: %s", err)
	}

	if stream.Duration == "" {
		if mediadata.Format != nil {
			f.Duration = mediadata.Format.DurationSeconds
		} else {
			return fmt.Errorf("failed to get duration: %s", err)
		}
	} else {
		duration, err := strconv.ParseFloat(stream.Duration, 64)
		if err != nil {
			return fmt.Errorf("failed to parse duration: %s", err)
		}

		f.Duration = duration
	}

	outputDir := path.Join(s.outputDir, f.StreamID)
	_, err = os.Stat(outputDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		mkdirErr := os.MkdirAll(outputDir, 0777)
		if mkdirErr != nil {
			return mkdirErr
		}

		return err
	}

	args := []string{
		"-i",
		f.Path,
		"-codec",
		"copy",
		"-f",
		"segment",
		"-segment_time",
		strconv.Itoa(s.segmentTime),
		"-segment_list",
		path.Join(s.outputDir, f.StreamID, "index.m3u8"),
		path.Join(s.outputDir, f.StreamID, "%d.ts"),
	}

	logger.Infof("ffmpeg %s", strings.Join(args, " "))

	cmd := exec.Command("ffmpeg", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to split: %s: %s", err, out)
	}

	return nil
}
