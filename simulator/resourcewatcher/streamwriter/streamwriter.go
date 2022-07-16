package streamwriter

//go:generate mockgen --build_flags=--mod=mod -destination=./mock_$GOPACKAGE/responseStream.go . ResponseStream
import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/watch"
)

// ResourceKind represents k8s resource name.
type ResourceKind string

// WatchEvent represents an event notified by the watched apiserver.
type WatchEvent struct {
	Kind      ResourceKind
	EventType watch.EventType
	// Obj is an object included in the event notified by the watched apiserver.
	Obj interface{}
}

// StreamWriter operates a given stream to send a received WatchEvent to the frontend.
type StreamWriter struct {
	sync.Mutex
	// stream knows where to write a received WatchEvent and how to send it.
	stream ResponseStream
	// encoder is a json encoder and the result will be written to the above stream via io.Writer.
	encoder *json.Encoder
}

func NewStreamWriter(stream ResponseStream) *StreamWriter {
	return &StreamWriter{
		stream:  stream,
		encoder: json.NewEncoder(stream),
	}
}

// Write encodes the an received WatchEvent and push it to the frontend.
func (sw *StreamWriter) Write(we *WatchEvent) error {
	sw.Lock()
	defer sw.Unlock()
	if err := sw.encoder.Encode(we); err != nil {
		return xerrors.Errorf("encode a WatchEvent: %w", err)
	}
	sw.stream.Flush()
	return nil
}

// ResponseStream is an interface that allows Server Push to a Service.
type ResponseStream interface {
	io.Writer
	http.Flusher
}
