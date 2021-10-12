// +build !test

package upload

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/livekit/protocol/logger"

	"github.com/livekit/livekit-recorder/pkg/config"
)

const (
	chunkSize  = 5 * 1024 * 1024
	maxRetries = 5
)

type Uploader struct {
	reader *Reader

	url string
	svc *s3.S3
	res *s3.CreateMultipartUploadOutput

	abort  chan struct{}
	result chan error
}

func NewUploader(s3Conf config.S3Config, s3Url, filename string) (*Uploader, error) {
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

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			s3Conf.AccessKey,
			s3Conf.Secret,
			"",
		),
		Region: aws.String(s3Conf.Region),
	})
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)
	input := &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String("video/mp4"),
	}

	res, err := svc.CreateMultipartUpload(input)
	if err != nil {
		return nil, err
	}

	reader, err := NewReader(filename, chunkSize)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			f, err := os.Open(filename)
			if err == nil {
				info, err := f.Stat()
				if err == nil {
					logger.Debugw("file opened", "size", info.Size())
				}
				f.Close()
			}
			time.Sleep(time.Second)
		}
	}()

	return &Uploader{
		reader: reader,
		url:    s3Url,
		svc:    svc,
		res:    res,
		abort:  make(chan struct{}, 1),
		result: make(chan error, 1),
	}, nil
}

// Run blocks until completion
func (u *Uploader) Run() {
	err := u.run()
	if err != nil {
		logger.Errorw("upload failed", err)
		u.abortUpload()
	}
	u.result <- err
}

func (u *Uploader) run() error {
	partNum := 1
	completed := make([]*s3.CompletedPart, 0)

	eof := false
	for !eof {
		select {
		case <-u.abort:
			u.abortUpload()
			return nil
		default:
			chunk, err := u.reader.Next()
			if err == nil || err == io.EOF {
				if err == io.EOF {
					logger.Debugw("Uploader hit EOF")
					eof = true
				}
				part, err := u.uploadPart(chunk, partNum)
				if err != nil {
					return err
				}
				completed = append(completed, part)
				partNum++
			} else {
				return err
			}
		}
	}

	logger.Debugw("resend first chunk")
	// resend first part
	chunk, err := u.reader.GetFirst()
	part, err := u.uploadPart(chunk, 1)
	if err != nil {
		return err
	}
	completed[0] = part

	// finished
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   u.res.Bucket,
		Key:      u.res.Key,
		UploadId: u.res.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completed,
		},
	}

	logger.Debugw("complete upload")
	_, err = u.svc.CompleteMultipartUpload(completeInput)
	if err != nil {
		return err
	}

	return nil
}

func (u *Uploader) Abort() {
	u.abort <- struct{}{}
}

func (u *Uploader) Finish() (string, error) {
	logger.Debugw("finish uploading")
	u.reader.Drain()
	logger.Debugw("waiting for result")
	err := <-u.result
	u.reader.Close()
	return u.url, err
}

func (u *Uploader) uploadPart(chunk []byte, partNumber int) (*s3.CompletedPart, error) {
	start := time.Now()
	logger.Debugw("uploading part", "num", partNumber, "size", len(chunk))
	partInput := &s3.UploadPartInput{
		Body:          bytes.NewReader(chunk),
		Bucket:        u.res.Bucket,
		Key:           u.res.Key,
		PartNumber:    aws.Int64(int64(partNumber)),
		UploadId:      u.res.UploadId,
		ContentLength: aws.Int64(int64(len(chunk))),
	}

	var partOutput *s3.UploadPartOutput
	var err error
	for retry := 0; retry <= maxRetries; retry++ {
		partOutput, err = u.svc.UploadPart(partInput)
		if err == nil {
			logger.Debugw("upload part finished", "time", fmt.Sprint(time.Since(start)))
			return &s3.CompletedPart{
				ETag:       partOutput.ETag,
				PartNumber: aws.Int64(int64(partNumber)),
			}, nil
		}
	}

	return nil, err
}

func (u *Uploader) abortUpload() {
	logger.Debugw("aborting upload")
	if u.res == nil {
		return
	}

	abortInput := &s3.AbortMultipartUploadInput{
		Bucket:   u.res.Bucket,
		Key:      u.res.Key,
		UploadId: u.res.UploadId,
	}
	_, err := u.svc.AbortMultipartUpload(abortInput)
	if err != nil {
		logger.Errorw("failed to abort upload", err)
	}
}
