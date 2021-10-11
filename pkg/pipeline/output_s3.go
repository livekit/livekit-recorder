// +build !test

package pipeline

import (
	"fmt"
	"strings"
	"time"

	"github.com/tinyzimmer/go-gst/gst"
)

type S3OutputBin struct {
	*OutputBase
	sink *gst.Element
}

func NewS3OutputBin(region, s3Url string) (OutputBin, error) {
	var bucket, key string
	if strings.HasPrefix(s3Url, "s3://") {
		s3Url = s3Url[5:]
	}
	if idx := strings.Index(s3Url, "/"); idx != -1 {
		bucket = s3Url[:idx]
		key = s3Url[idx+1:]
	} else {
		bucket = s3Url
		key = fmt.Sprintf("recording-%s.mp4", time.Now().Format("20060102150405"))
	}

	// s3sink needs a large queue in front of it
	queue, err := gst.NewElement("queue")
	if err != nil {
		return nil, err
	}
	if err = queue.SetProperty("max-size-bytes", uint(10*1024*1024)); err != nil {
		return nil, err
	}

	sink, err := gst.NewElement("s3sink")
	if err != nil {
		return nil, err
	}
	if err = sink.SetProperty("region", region); err != nil {
		return nil, err
	}
	if err = sink.SetProperty("bucket", bucket); err != nil {
		return nil, err
	}
	if err = sink.SetProperty("key", key); err != nil {
		return nil, err
	}
	if err = sink.SetProperty("content-type", "video/mp4"); err != nil {
		return nil, err
	}

	// create bin
	bin := gst.NewBin("output")
	if err = bin.Add(sink); err != nil {
		return nil, err
	}

	// add ghost pad
	ghostPad := gst.NewGhostPad("sink", sink.GetStaticPad("sink"))
	if !bin.AddPad(ghostPad.Pad) {
		return nil, ErrGhostPadFailed
	}

	return &S3OutputBin{
		OutputBase: &OutputBase{bin},
		sink:       sink,
	}, nil
}
