package upload

import (
	"io"
	"os"

	"github.com/livekit/protocol/logger"
	"github.com/nxadm/tail"
)

type Reader struct {
	tail      *tail.Tail
	filename  string
	chunkSize int
	total     int64
	drain     chan struct{}
	buffer    []byte
}

func NewReader(filename string, chunkSize int) (*Reader, error) {
	t, err := tail.TailFile(filename, tail.Config{
		MustExist:   false,
		Follow:      true,
		MaxLineSize: chunkSize / 4,
	})
	if err != nil {
		return nil, err
	}
	return &Reader{
		tail:      t,
		filename:  filename,
		chunkSize: chunkSize,
		buffer:    make([]byte, 0, chunkSize),
	}, nil
}

func (r *Reader) Next() ([]byte, error) {
	if r.chunkSize < chunkSize*100 {
		r.chunkSize *= 10
	} else {
		r.chunkSize = chunkSize
	}

	lines := r.tail.Lines
	for {
		select {
		case <-r.drain:
			go func() {
				err := r.tail.StopAtEOF()
				logger.Errorw("tail stop", err)
			}()
		case line := <-lines:
			if line.Err != nil && line.Err != io.EOF {
				logger.Errorw("line error", line.Err)
				return nil, line.Err
			}

			r.buffer = append(r.buffer, []byte(line.Text+"\n")...)
			logger.Debugw("file read", "len", len([]byte(line.Text)))
			if line.Err == io.EOF {
				return r.buffer, io.EOF
			}

			if len(r.buffer) < r.chunkSize {
				continue
			}

			if len(r.buffer) == r.chunkSize {
				b := r.buffer
				r.buffer = make([]byte, 0, r.chunkSize)
				return b, nil
			} else {
				b := r.buffer[:chunkSize]
				r.buffer = r.buffer[chunkSize:]
				return b, nil
			}
		}
	}
}

func (r *Reader) Drain() {
	logger.Debugw("draining reader")
	r.drain <- struct{}{}
}

func (r *Reader) GetFirst() ([]byte, error) {
	f, err := os.Open(r.filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buffer := make([]byte, r.chunkSize)
	n, err := f.Read(buffer)
	if err != nil {
		if err == io.EOF {
			return buffer[:n], nil
		}
		return nil, err
	}

	return buffer, nil
}

func (r *Reader) Close() {
	if r.tail != nil {
		r.tail.Cleanup()
	}
}
