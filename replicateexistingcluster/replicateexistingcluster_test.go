package replicateexistingcluster

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"golang.org/x/xerrors"

	"github.com/kubernetes-sigs/kube-scheduler-simulator/export"
	"github.com/kubernetes-sigs/kube-scheduler-simulator/replicateexistingcluster/mock_replicateexistingcluster"
)

func TestService_ImportFromExistingCluster(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		prepareEachServiceMockFn func(existingExport *mock_replicateexistingcluster.MockExportService, clusterExport *mock_replicateexistingcluster.MockExportService)
		wantErr                  bool
	}{
		{
			name: "no error when Import and Export are successful",
			prepareEachServiceMockFn: func(simulatorExport *mock_replicateexistingcluster.MockExportService, clusterExport *mock_replicateexistingcluster.MockExportService) {
				dummyOption := new(export.Option)
				clusterExport.EXPECT().Export(gomock.Any()).Return(&export.ResourcesForExport{}, nil)
				simulatorExport.EXPECT().Import(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				simulatorExport.EXPECT().IgnoreErr().Return(*dummyOption).Times(1)
				simulatorExport.EXPECT().IgnoreSchedulerConfiguration().Return(*dummyOption).Times(1)
			},
			wantErr: false,
		},
		{
			name: "should return error if Import raise an error",
			prepareEachServiceMockFn: func(simulatorExport *mock_replicateexistingcluster.MockExportService, clusterExport *mock_replicateexistingcluster.MockExportService) {
				dummyOption := new(export.Option)
				clusterExport.EXPECT().Export(gomock.Any()).Return(&export.ResourcesForExport{}, nil)
				simulatorExport.EXPECT().Import(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(xerrors.Errorf("failed to Import"))
				simulatorExport.EXPECT().IgnoreErr().Return(*dummyOption).Times(1)
				simulatorExport.EXPECT().IgnoreSchedulerConfiguration().Return(*dummyOption).Times(1)
			},
			wantErr: true,
		},
		{
			name: "should return error if Export raise an error",
			prepareEachServiceMockFn: func(simulatorExport *mock_replicateexistingcluster.MockExportService, clusterExport *mock_replicateexistingcluster.MockExportService) {
				dummyOption := new(export.Option)
				clusterExport.EXPECT().Export(gomock.Any(), gomock.Any()).Return(nil, xerrors.Errorf("failed to Import"))
				simulatorExport.EXPECT().Import(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
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

			mockSimulatorExportService := mock_replicateexistingcluster.NewMockExportService(ctrl)
			mockClusterExportService := mock_replicateexistingcluster.NewMockExportService(ctrl)

			s := NewReplicateExistingClusterService(mockSimulatorExportService, mockClusterExportService)
			tt.prepareEachServiceMockFn(mockSimulatorExportService, mockClusterExportService)

			if err := s.ImportFromExistingCluster(context.Background()); (err != nil) != tt.wantErr {
				t.Fatalf("ImportFromExistingCluster() %v test, \nerror = %v", tt.name, err)
			}
		})
	}
}
