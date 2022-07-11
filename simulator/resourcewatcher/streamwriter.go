package resourcewatcher

//go:generate mockgen --build_flags=--mod=mod -destination=./mock_$GOPACKAGE/streamWriter.go . StreamWriter

import (
	"encoding/json"
	"sync"

	"golang.org/x/xerrors"
)

// StreamWriter is an interface that allows send a received WatchEvent to the frontend.
type StreamWriter interface {
	Write(we *WatchEvent) error
}

// streamWriter operates a given stream to send a received WatchEvent to the frontend.
type streamWriter struct {
	sync.Mutex
	// stream knows where to write a received WatchEvent and how to send it.
	stream ResponseStream
	// encoder is a json encoder and the result will be written to the above stream via io.Writer.
	encoder *json.Encoder
}

func newStreamWriter(stream ResponseStream) *streamWriter {
	return &streamWriter{
		stream:  stream,
		encoder: json.NewEncoder(stream),
	}
}

// Write encodes the an received WatchEvent and push it to the frontend.
func (sw *streamWriter) Write(we *WatchEvent) error {
	sw.Lock()
	defer sw.Unlock()
	if err := sw.encoder.Encode(we); err != nil {
		return xerrors.Errorf("encode a WatchEvent: %w", err)
	}
	sw.stream.Flush()
	return nil
}
