package replayer

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/recorder"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/replayer/mock_resourceapplier"
)

func TestService_Replay(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		records       []recorder.Record
		prepareMockFn func(*mock_resourceapplier.MockResourceApplier)
		wantErr       bool
	}{
		{
			name: "no error when Create is successful",
			records: []recorder.Record{
				{
					Event: recorder.Add,
					Resource: unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"name":      "pod-1",
								"namespace": "default",
							},
						},
					},
				},
			},
			prepareMockFn: func(applier *mock_resourceapplier.MockResourceApplier) {
				applier.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "should return error if Create raise an error",
			records: []recorder.Record{
				{
					Event: recorder.Add,
					Resource: unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"name":      "pod-1",
								"namespace": "default",
							},
						},
					},
				},
			},
			prepareMockFn: func(applier *mock_resourceapplier.MockResourceApplier) {
				applier.EXPECT().Create(gomock.Any(), gomock.Any()).Return(xerrors.Errorf("failed to create resource"))
			},
			wantErr: true,
		},
		{
			name: "ignore AlreadyExists error when Create raise an error",
			records: []recorder.Record{
				{
					Event: recorder.Add,
					Resource: unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"name":      "pod-1",
								"namespace": "default",
							},
						},
					},
				},
			},
			prepareMockFn: func(applier *mock_resourceapplier.MockResourceApplier) {
				applier.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.NewAlreadyExists(schema.GroupResource{}, "resource already exists"))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockApplier := mock_resourceapplier.NewMockResourceApplier(ctrl)
			tt.prepareMockFn(mockApplier)

			fileName := strings.ReplaceAll(tt.name, " ", "_") + ".jsonl"
			filePath := path.Join(os.TempDir(), fileName)
			tempFile, err := os.Create(filePath)
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer os.Remove(filePath)

			err = writeRecordsToFile(tempFile, tt.records)
			if err != nil {
				t.Fatalf("failed to marshal records: %v", err)
			}

			err = tempFile.Close()
			if err != nil {
				t.Fatalf("failed to close temp file: %v", err)
			}

			service := New(mockApplier, Options{RecordFile: filePath})

			err = service.Replay(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.Replay() error = %v", err)
			}
		})
	}
}

func writeRecordsToFile(file *os.File, records []recorder.Record) error {
	for _, record := range records {
		b, err := json.Marshal(&record)
		if err != nil {
			return xerrors.Errorf("failed to marshal record: %w", err)
		}

		if _, err := file.Write(append(b, '\n')); err != nil {
			return xerrors.Errorf("failed to write record to file: %w", err)
		}
	}

	return nil
}
