// +build test

package upload

import "github.com/livekit/livekit-recorder/pkg/config"

type Uploader struct {
	url string
}

func NewUploader(s3Conf config.S3Config, s3Url, filename string) (*Uploader, error) {
	return &Uploader{url: s3Url}, nil
}

func (u *Uploader) Run() {}

func (u *Uploader) Abort() {}

func (u *Uploader) Finish() (string, error) {
	return u.url, nil
}
