package pipeline

import "github.com/tinyzimmer/go-gst/gst"

type OutputBin interface {
	Bin() *gst.Element
	Link() error
	AddRtmpSink(url string) error
	RemoveRtmpSink(url string) error
}

type OutputBase struct {
	bin *gst.Bin
}

func (b *OutputBase) Bin() *gst.Element {
	return b.bin.Element
}

func (b *OutputBase) Link() error {
	return nil
}

func (b *OutputBase) AddRtmpSink(url string) error {
	return ErrNotSupported
}

func (b *OutputBase) RemoveRtmpSink(url string) error {
	return ErrNotSupported
}
