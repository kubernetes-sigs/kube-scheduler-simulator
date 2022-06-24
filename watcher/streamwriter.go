package watcher

//go:generate mockgen -destination=./mock_$GOPACKAGE/streamWriter.go . StreamWriter

import (
	"encoding/json"
	"sync"

	"golang.org/x/xerrors"
)

type StreamWriter interface {
	Write(we *WatchEvent) error
}

type streamWriter struct {
	sync.Mutex
	stream  ResponseStream
	encoder *json.Encoder
}

func newStreamWriter(stream ResponseStream) *streamWriter {
	return &streamWriter{
		stream:  stream,
		encoder: json.NewEncoder(stream),
	}
}

func (sw *streamWriter) Write(we *WatchEvent) error {
	sw.Lock()
	defer sw.Unlock()
	if err := sw.encoder.Encode(we); err != nil {
		return xerrors.Errorf("encode a WatchEvent: %w", err)
	}
	sw.stream.Flush()
	return nil
}
