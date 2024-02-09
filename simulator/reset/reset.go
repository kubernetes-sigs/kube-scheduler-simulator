package reset

import (
	"context"
	"errors"

	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/xerrors"
	clientset "k8s.io/client-go/kubernetes"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/util"
)

type SchedulerService interface {
	ResetScheduler() error
}

// Service cleans up resources stored in etcd.
type Service struct {
	// initialData has the all resource data that are fetched when reset service is initialized.
	initialData map[string]string

	etcdClient   *clientv3.Client
	k8sClient    clientset.Interface
	schedService SchedulerService
}

// NewResetService initializes Service.
// ResetService always tries to restore the cluster to the initial state.
func NewResetService(
	etcdClient *clientv3.Client,
	k8sClient clientset.Interface,
	schedService SchedulerService,
) (*Service, error) {
	s := &Service{
		initialData:  map[string]string{},
		etcdClient:   etcdClient,
		k8sClient:    k8sClient,
		schedService: schedService,
	}

	result, err := etcdClient.Get(context.Background(), util.EtcdPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, xerrors.Errorf("get all data in etcd: %w", err)
	}

	for _, v := range result.Kvs {
		s.initialData[string(v.Key)] = string(v.Value)
	}

	return s, nil
}

// Reset resets all resources and scheduler configuration to the initial state.
func (s *Service) Reset(ctx context.Context) error {
	if _, err := s.etcdClient.Delete(ctx, util.EtcdPrefix, clientv3.WithPrefix()); err != nil {
		return xerrors.Errorf("delete all data in etcd: %w", err)
	}

	// restore initial data.
	eg := util.NewErrGroupWithSemaphore(ctx)
	for k, v := range s.initialData {
		k := k
		v := v
		err := eg.Go(func() error {
			if _, err := s.etcdClient.Put(ctx, k, v); err != nil {
				return xerrors.Errorf("put initial data in etcd: key: %s, value: %s, error: %w", k, v, err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	if err := s.schedService.ResetScheduler(); err != nil && !errors.Is(err, scheduler.ErrServiceDisabled) {
		return xerrors.Errorf("reset scheduler: %w", err)
	}
	return nil
}
