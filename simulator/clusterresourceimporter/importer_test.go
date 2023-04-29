package clusterresourceimporter

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"golang.org/x/xerrors"

	m "sigs.k8s.io/kube-scheduler-simulator/simulator/clusterresourceimporter/mock_clusterresourceimporter"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/snapshot"
)

func TestService_ImportClusterResources(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		prepareEachServiceMockFn func(*m.MockReplicateService, *m.MockReplicateService)
		wantErr                  bool
	}{
		{
			name: "no error when Load and Save are successful",
			prepareEachServiceMockFn: func(simulatorExport *m.MockReplicateService, clusterExport *m.MockReplicateService) {
				dummyOption := new(snapshot.Option)
				clusterExport.EXPECT().Save(gomock.Any()).Return(&snapshot.ResourcesForSave{}, nil)
				simulatorExport.EXPECT().Load(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				simulatorExport.EXPECT().IgnoreErr().Return(*dummyOption).Times(1)
				simulatorExport.EXPECT().IgnoreSchedulerConfiguration().Return(*dummyOption).Times(1)
			},
			wantErr: false,
		},
		{
			name: "should return error if Load raise an error",
			prepareEachServiceMockFn: func(simulatorExport *m.MockReplicateService, clusterExport *m.MockReplicateService) {
				dummyOption := new(snapshot.Option)
				clusterExport.EXPECT().Save(gomock.Any()).Return(&snapshot.ResourcesForSave{}, nil)
				simulatorExport.EXPECT().Load(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(xerrors.Errorf("failed to Import"))
				simulatorExport.EXPECT().IgnoreErr().Return(*dummyOption).Times(1)
				simulatorExport.EXPECT().IgnoreSchedulerConfiguration().Return(*dummyOption).Times(1)
			},
			wantErr: true,
		},
		{
			name: "should return error if Save raise an error",
			prepareEachServiceMockFn: func(simulatorExport *m.MockReplicateService, clusterExport *m.MockReplicateService) {
				dummyOption := new(snapshot.Option)
				clusterExport.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil, xerrors.Errorf("failed to Import"))
				simulatorExport.EXPECT().Load(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
				simulatorExport.EXPECT().IgnoreSchedulerConfiguration().Return(*dummyOption).Times(0)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockSimulatorExportService := m.NewMockReplicateService(ctrl)
			mockClusterExportService := m.NewMockReplicateService(ctrl)

			s := NewService(mockSimulatorExportService, mockClusterExportService)
			tt.prepareEachServiceMockFn(mockSimulatorExportService, mockClusterExportService)

			if err := s.ImportClusterResources(context.Background()); (err != nil) != tt.wantErr {
				t.Fatalf("ImportClusterResources() %v test, \nerror = %v", tt.name, err)
			}
		})
	}
}
