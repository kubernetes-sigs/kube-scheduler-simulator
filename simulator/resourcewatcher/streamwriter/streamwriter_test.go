package streamwriter

import (
	"testing"

	"github.com/golang/mock/gomock"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/resourcewatcher/streamwriter/mock_streamwriter"
)

const (
	Pods ResourceKind = "pods"
)

var (
	dummyWatchEvent1 = WatchEvent{
		Kind:      Pods,
		EventType: watch.Added,
		Obj:       Pod1,
	}
	Pod1 = corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod1",
		},
	}
)

func TestStreamWriter_Writer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                        string
		prepareResponseStreamMockFn func(ws *mock_streamwriter.MockResponseStream)
		wantErr                     bool
	}{
		{
			name: "should success when ResponseWriter's Write method returns no error",
			prepareResponseStreamMockFn: func(ws *mock_streamwriter.MockResponseStream) {
				ws.EXPECT().Flush()
				ws.EXPECT().Write(gomock.Any()).Return(0, nil)
			},
			wantErr: false,
		},
		{
			name: "should failed when ResponseWriter's Write method returns an error",
			prepareResponseStreamMockFn: func(ws *mock_streamwriter.MockResponseStream) {
				ws.EXPECT().Write(gomock.Any()).Return(0, xerrors.Errorf("call write"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockResponseStream := mock_streamwriter.NewMockResponseStream(ctrl)

			sw := NewStreamWriter(mockResponseStream)
			tt.prepareResponseStreamMockFn(mockResponseStream)

			if err := sw.Write(&dummyWatchEvent1); (err != nil) != tt.wantErr {
				t.Fatalf("Writer %v test, \nerror = %v", tt.name, err)
			}
		})
	}
}
