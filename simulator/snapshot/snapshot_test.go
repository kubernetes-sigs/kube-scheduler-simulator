package snapshot

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	schedulingcfgv1 "k8s.io/client-go/applyconfigurations/scheduling/v1"
	confstoragev1 "k8s.io/client-go/applyconfigurations/storage/v1"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	configv1 "k8s.io/kube-scheduler/config/v1"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler"
	schedulerCfg "sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/config"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/snapshot/mock_snapshot"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/util"
)

const (
	testNamespace1 = "default1"
	testNamespace2 = "default2"
)

type (
	resourceName         string
	SettingClientFunc    func(context.Context, *fake.Clientset)
	SettingClientFuncMap map[resourceName]SettingClientFunc
)

var (
	ns   resourceName = "namespace"
	node resourceName = "node"
	pod  resourceName = "pod"
	pv   resourceName = "persistentVolume"
	pvc  resourceName = "persistentVolumeClaim"
	sc   resourceName = "storageClass"
	pc   resourceName = "priorityClass"
	// order indicates calling order in the invokeResourcesFn method.
	order = []resourceName{
		ns, node, pod, pv, pvc, sc, pc,
	}
	// defaultResForSnapFn returns default expected ResourcesForSnap.
	defaultResForSnapFn = func() *ResourcesForSnap {
		return &ResourcesForSnap{
			Pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: testNamespace1,
					},
					Spec: corev1.PodSpec{
						NodeName: "node1",
					},
				},
			},
			Nodes: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				},
			},
			Pvs: []corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pv1",
					},
				},
			},
			Pvcs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pvc1",
						Namespace: testNamespace1,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "pv1",
					},
				},
			},
			StorageClasses: []storagev1.StorageClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "sc1",
					},
				},
			},
			PriorityClasses: []schedulingv1.PriorityClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pc1",
					},
				},
			},
			SchedulerConfig: &configv1.KubeSchedulerConfiguration{},
			Namespaces: []corev1.Namespace{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: testNamespace1,
					},
				},
			},
		}
	}
	// defaultApplyFuncs returns default expected settings to fakeClientset for `Apply`.
	defaultApplyFuncs = SettingClientFuncMap{
		ns: func(ctx context.Context, c *fake.Clientset) {
			c.PrependReactor("patch", "namespaces", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &corev1.Namespace{}, nil
			})
		},
		node: func(ctx context.Context, c *fake.Clientset) {
			c.PrependReactor("patch", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &corev1.Node{}, nil
			})
		},
		pod: func(ctx context.Context, c *fake.Clientset) {
			c.PrependReactor("patch", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &corev1.Pod{}, nil
			})
		},
		pv: func(ctx context.Context, c *fake.Clientset) {
			c.PrependReactor("patch", "persistentvolumes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &corev1.PersistentVolume{}, nil
			})
		},
		pvc: func(ctx context.Context, c *fake.Clientset) {
			c.PrependReactor("patch", "persistentvolumeclaims", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &corev1.PersistentVolumeClaim{}, nil
			})
		},
		sc: func(ctx context.Context, c *fake.Clientset) {
			c.PrependReactor("patch", "storageclasses", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &storagev1.StorageClass{}, nil
			})
		},
		pc: func(ctx context.Context, c *fake.Clientset) {
			c.PrependReactor("patch", "priorityclasses", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &schedulingv1.PriorityClass{}, nil
			})
		},
	}
	// defaultFuncs returns default methods setting to fakeClientset for create each resources.
	defaultFuncs = SettingClientFuncMap{
		ns: func(ctx context.Context, c *fake.Clientset) {
			c.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: testNamespace1,
				},
			}, metav1.CreateOptions{})
		},
		node: func(ctx context.Context, c *fake.Clientset) {
			c.CoreV1().Nodes().Create(ctx, &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node1",
				},
			}, metav1.CreateOptions{})
		},
		pod: func(ctx context.Context, c *fake.Clientset) {
			c.CoreV1().Pods(testNamespace1).Create(ctx, &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod1",
				},
				Spec: corev1.PodSpec{
					NodeName: "node1",
				},
			}, metav1.CreateOptions{})
		},
		pv: func(ctx context.Context, c *fake.Clientset) {
			c.CoreV1().PersistentVolumes().Create(ctx, &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pv1",
				},
			}, metav1.CreateOptions{})
		},
		pvc: func(ctx context.Context, c *fake.Clientset) {
			c.CoreV1().PersistentVolumeClaims(testNamespace1).Create(ctx, &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pvc1",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "pv1",
				},
			}, metav1.CreateOptions{})
		},
		sc: func(ctx context.Context, c *fake.Clientset) {
			c.StorageV1().StorageClasses().Create(ctx, &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sc1",
				},
			}, metav1.CreateOptions{})
		},
		pc: func(ctx context.Context, c *fake.Clientset) {
			c.SchedulingV1().PriorityClasses().Create(ctx, &schedulingv1.PriorityClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pc1",
				},
			}, metav1.CreateOptions{})
		},
	}
	// invokeResourcesFn invokes specified individual methods or baseFuncs.
	invokeResourcesFn = func(ctx context.Context, c *fake.Clientset, funcs SettingClientFuncMap, baseFuncs SettingClientFuncMap) {
		for _, n := range order {
			if fn, ok := funcs[n]; ok {
				fn(ctx, c)
				continue
			}
			baseFuncs[n](ctx, c)
		}
	}
)

func TestService_Snap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		prepareEachServiceMockFn func(*mock_snapshot.MockSchedulerService)
		prepareFakeClientSetFn   func() *fake.Clientset
		wantReturn               func() *ResourcesForSnap
		wantErr                  error
	}{
		{
			name: "Snap all resources",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{}, defaultFuncs)
				return c
			},
			wantReturn: defaultResForSnapFn,
		},
		{
			name: "Snap failure on List Pod",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					pod: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("list", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to list Pod")
						})
					},
				}, defaultFuncs)
				return c
			},
			wantErr: xerrors.New("failed to get(): get resources all: call list Pod: failed to list Pod"),
		},
		{
			name: "Snap failure on List Node",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					node: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("list", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to list Node")
						})
					},
				}, defaultFuncs)
				return c
			},
			wantErr: xerrors.New("failed to get(): get resources all: call list Node: failed to list Node"),
		},
		{
			name: "Snap failure on List PersistentVolume",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					pv: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("list", "persistentvolumes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to list PersistentVolume")
						})
					},
				}, defaultFuncs)
				return c
			},
			wantErr: xerrors.New("failed to get(): get resources all: call list PersistentVolume: failed to list PersistentVolume"),
		},
		{
			name: "Snap failure on List PersistentVolumeClaim",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					pvc: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("list", "persistentvolumeclaims", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to list PersistentVolumeClaim")
						})
					},
				}, defaultFuncs)
				return c
			},
			wantErr: xerrors.New("failed to get(): get resources all: call list PersistentVolumeClaim: failed to list PersistentVolumeClaim"),
		},
		{
			name: "Snap failure on List of StorageClass",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					sc: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("list", "storageclasses", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to list StorageClass")
						})
					},
				}, defaultFuncs)
				return c
			},
			wantErr: xerrors.New("failed to get(): get resources all: call list StorageClass: failed to list StorageClass"),
		},
		{
			name: "Snap failure on List PriorityClass",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					pc: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("list", "priorityclasses", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to list PriorityClass")
						})
					},
				}, defaultFuncs)
				return c
			},
			wantErr: xerrors.New("failed to get(): get resources all: call list PriorityClass: failed to list PriorityClass"),
		},
		{
			name: "Snap failure on List Namespace",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					ns: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("list", "namespaces", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to list Namespace")
						})
					},
				}, defaultFuncs)
				return c
			},
			wantErr: xerrors.New("failed to get(): get resources all: call list Namespace: failed to list Namespace"),
		},
		{
			name: "Snap failure on get scheduler config",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, xerrors.New("failed to get config"))
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{}, defaultFuncs)
				return c
			},
			wantErr: xerrors.New("failed to get(): get resources all: get scheduler config: failed to get config"),
		},
		{
			name: "get ErrServiceDisabled on get scheduler config, but it's ignored.",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(nil, scheduler.ErrServiceDisabled)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{}, defaultFuncs)
				return c
			},
			wantReturn: func() *ResourcesForSnap {
				r := defaultResForSnapFn()
				r.SchedulerConfig = nil
				return r
			},
		},
		{
			name: "Snap all Pods with different namespaces",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					pod: func(ctx context.Context, c *fake.Clientset) {
						c.CoreV1().Pods(testNamespace1).Create(ctx, &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod1",
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						}, metav1.CreateOptions{})
						c.CoreV1().Pods(testNamespace2).Create(ctx, &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pod2",
							},
							Spec: corev1.PodSpec{
								NodeName: "node1",
							},
						}, metav1.CreateOptions{})
					},
				}, defaultFuncs)
				return c
			},
			wantReturn: func() *ResourcesForSnap {
				r := defaultResForSnapFn()
				r.Pods = []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: testNamespace1},
						Spec:       corev1.PodSpec{NodeName: "node1"},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: testNamespace2},
						Spec:       corev1.PodSpec{NodeName: "node1"},
					},
				}
				return r
			},
		},
		{
			name: "Snap all pvcs with different namespaces",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					pvc: func(ctx context.Context, c *fake.Clientset) {
						c.CoreV1().PersistentVolumeClaims(testNamespace1).Create(ctx, &corev1.PersistentVolumeClaim{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pvc1",
							},
							Spec: corev1.PersistentVolumeClaimSpec{
								VolumeName: "pv1",
							},
						}, metav1.CreateOptions{})
						c.CoreV1().PersistentVolumeClaims(testNamespace2).Create(ctx, &corev1.PersistentVolumeClaim{
							ObjectMeta: metav1.ObjectMeta{
								Name: "pvc2",
							},
							Spec: corev1.PersistentVolumeClaimSpec{
								VolumeName: "pv1",
							},
						}, metav1.CreateOptions{})
					},
				}, defaultFuncs)
				return c
			},
			wantReturn: func() *ResourcesForSnap {
				r := defaultResForSnapFn()
				r.Pvcs = []corev1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pvc1",
							Namespace: testNamespace1,
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							VolumeName: "pv1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "pvc2",
							Namespace: testNamespace2,
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							VolumeName: "pv1",
						},
					},
				}
				return r
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockSchedulerSvc := mock_snapshot.NewMockSchedulerService(ctrl)
			fakeClientset := tt.prepareFakeClientSetFn()

			s := NewService(fakeClientset, mockSchedulerSvc)
			tt.prepareEachServiceMockFn(mockSchedulerSvc)
			r, err := s.Snap(context.Background())

			var diffResponse string
			if tt.wantReturn != nil {
				diffResponse = cmp.Diff(tt.wantReturn(), r)
			}

			if diffResponse != "" || (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("Snap() %v test, \nerror = %v,\n%s", tt.name, err, diffResponse)
			}
			if tt.wantErr != nil {
				assert.EqualError(t, tt.wantErr, err.Error())
			}
		})
	}
}

func TestService_Snap_IgnoreErrOption(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		prepareEachServiceMockFn func(*mock_snapshot.MockSchedulerService)
		prepareFakeClientSetFn   func() *fake.Clientset
		wantReturn               func() *ResourcesForSnap
		wantErr                  error
	}{
		{
			name: "snap all resources",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{}, defaultFuncs)
				return c
			},
			wantReturn: defaultResForSnapFn,
		},
		{
			name: "no error if failure to list Pod",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					pod: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("list", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to list Pod")
						})
					},
				}, defaultFuncs)
				return c
			},
			wantReturn: func() *ResourcesForSnap {
				r := defaultResForSnapFn()
				r.Pods = []corev1.Pod{}
				return r
			},
		},
		{
			name: "no error if failure to list Node",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					node: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("list", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to list Node")
						})
					},
				}, defaultFuncs)
				return c
			},
			wantReturn: func() *ResourcesForSnap {
				r := defaultResForSnapFn()
				r.Nodes = []corev1.Node{}
				return r
			},
		},
		{
			name: "no error if failure to list PersistentVolume",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					pv: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("list", "persistentvolumes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to list PersistentVolume")
						})
					},
				}, defaultFuncs)
				return c
			},
			wantReturn: func() *ResourcesForSnap {
				r := defaultResForSnapFn()
				r.Pvs = []corev1.PersistentVolume{}
				return r
			},
		},
		{
			name: "no error if failure to list PersistentVolumeClaims",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					pvc: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("list", "persistentvolumeclaims", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to list PersistentVolumeClaim")
						})
					},
				}, defaultFuncs)
				return c
			},
			wantReturn: func() *ResourcesForSnap {
				r := defaultResForSnapFn()
				r.Pvcs = []corev1.PersistentVolumeClaim{}
				return r
			},
		},
		{
			name: "no error if failure to list storageClasses",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					sc: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("list", "storageclasses", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to list StorageClass")
						})
					},
				}, defaultFuncs)
				return c
			},
			wantReturn: func() *ResourcesForSnap {
				r := defaultResForSnapFn()
				r.StorageClasses = []storagev1.StorageClass{}
				return r
			},
		},
		{
			name: "no error if failure to list priorityClasses",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					pc: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("list", "priorityclasses", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to list PriorityClass")
						})
					},
				}, defaultFuncs)
				return c
			},
			wantReturn: func() *ResourcesForSnap {
				r := defaultResForSnapFn()
				r.PriorityClasses = []schedulingv1.PriorityClass{}
				return r
			},
		},
		{
			name: "no error if failure to list Namespace",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().GetSchedulerConfig().Return(&configv1.KubeSchedulerConfiguration{}, nil)
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					ns: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("list", "namespaces", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to list Namespace")
						})
					},
				}, defaultFuncs)
				return c
			},
			wantReturn: func() *ResourcesForSnap {
				r := defaultResForSnapFn()
				r.Namespaces = []corev1.Namespace{}
				return r
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			fakeClientset := tt.prepareFakeClientSetFn()
			mockSchedulerSvc := mock_snapshot.NewMockSchedulerService(ctrl)
			s := NewService(fakeClientset, mockSchedulerSvc)
			tt.prepareEachServiceMockFn(mockSchedulerSvc)
			r, err := s.Snap(context.Background(), s.IgnoreErr())

			var diffResponse string
			if tt.wantReturn != nil {
				diffResponse = cmp.Diff(tt.wantReturn(), r)
			}

			if diffResponse != "" || (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("Snap() %v test, \nerror = %v,\n%s", tt.name, err, diffResponse)
			}
			if tt.wantErr != nil {
				assert.EqualError(t, tt.wantErr, err.Error())
			}
		})
	}
}

func TestService_Load(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		prepareEachServiceMockFn func(*mock_snapshot.MockSchedulerService)
		prepareFakeClientSetFn   func() *fake.Clientset
		applyConfiguration       func() *ResourcesForLoad
		wantErr                  error
	}{
		{
			name: "load all success",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("StorageClass1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{}, defaultApplyFuncs)
				return c
			},
		},
		{
			name: "load all success (with external scheduler enabled.)",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(scheduler.ErrServiceDisabled)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("StorageClass1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{}, defaultApplyFuncs)
				return c
			},
		},
		{
			name: "load failure to apply Pod",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("something wrong", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("StorageClass1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{
					pod: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("patch", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to apply Pod")
						})
					},
				}, defaultApplyFuncs)
				return c
			},
			wantErr: xerrors.New("failed to apply(): apply resources: apply Pod: failed to apply Pod"),
		},
		{
			name: "load failure to apply Node",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("something wrong"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("StorageClass1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{
					node: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("patch", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to apply Node")
						})
					},
				}, defaultApplyFuncs)
				return c
			},
			wantErr: xerrors.New("failed to apply(): apply resources: apply Node: failed to apply Node"),
		},
		{
			name: "load failure to apply PersistentVolume",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("something wrong").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("StorageClass1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{
					pv: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("patch", "persistentvolumes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to apply PersistentVolume")
						})
					},
				}, defaultApplyFuncs)
				return c
			},
			wantErr: xerrors.New("failed to apply(): apply PVs: apply PersistentVolume: failed to apply PersistentVolume"),
		},
		{
			name: "load failure to apply PersistentVolumeClaim",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("something wrong", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("StorageClass1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{
					pvc: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("patch", "persistentvolumeclaims", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to apply PersistentVolumeClaim")
						})
					},
				}, defaultApplyFuncs)
				return c
			},
			wantErr: xerrors.New("failed to apply(): apply resources: apply PersistentVolumeClaims: failed to apply PersistentVolumeClaim"),
		},
		{
			name: "load failure to apply StorageClass",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("something wrong"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{
					sc: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("patch", "storageclasses", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to apply StorageClass")
						})
					},
				}, defaultApplyFuncs)
				return c
			},
			wantErr: xerrors.New("failed to apply(): apply resources: apply StorageClass: failed to apply StorageClass"),
		},
		{
			name: "load failure to apply PriorityClass",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("SC1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("something wrong"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{
					pc: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("patch", "priorityclasses", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to apply PriorityClass")
						})
					},
				}, defaultApplyFuncs)
				return c
			},
			wantErr: xerrors.New("failed to apply(): apply resources: apply PriorityClass: failed to apply PriorityClass"),
		},
		{
			name: "load failure to apply Namespace",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace("something wrong"),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("SC1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PC1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{
					ns: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("patch", "namespaces", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to apply Namespace")
						})
					},
				}, defaultApplyFuncs)
				return c
			},
			wantErr: xerrors.New("failed to apply(): apply resources: apply Namespace: failed to apply Namespace"),
		},
		{
			name: "load success when PersistentVolumeClaim was not found (Get() return err)",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace("something wrong"),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Bound")).WithSpec(v1.PersistentVolumeSpec().WithClaimRef(v1.ObjectReference().WithName("PVC1").WithNamespace(testNamespace1))),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("SC1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PC1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{
					pvc: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("patch", "persistentvolumeclaims", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, &corev1.PersistentVolumeClaim{}, nil
						})
						c.PrependReactor("get", "persistentvolumeclaims", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							a, ok := action.(k8stesting.GetAction)
							assert.Equal(t, true, ok)
							assert.Equal(t, "PVC1", a.GetName())
							return true, nil, xerrors.New("pvc not found")
						})
					},
				}, defaultApplyFuncs)
				return c
			},
		},
		{
			name: "load success when PersistentVolumeClaim was found (Get() success)",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Bound")).WithSpec(v1.PersistentVolumeSpec().WithClaimRef(v1.ObjectReference().WithName("PVC1").WithNamespace(testNamespace1))),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("SC1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PC1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{}, defaultApplyFuncs)
				c.PrependReactor("get", "persistentvolumeclaims", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					a, ok := action.(k8stesting.GetAction)
					assert.Equal(t, true, ok)
					assert.Equal(t, "PVC1", a.GetName())
					o := &corev1.PersistentVolumeClaim{
						ObjectMeta: metav1.ObjectMeta{
							UID:  "testUID",
							Name: "PVC1",
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							VolumeName: "PV1",
						},
					}
					return true, o, nil
				})
				return c
			},
		},
		{
			name: "success load when pv.Status is not exist",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				j := `{"pods":[],"nodes":[],`
				// delete status object
				j += `"pvs":[{"metadata":{"name":"pv1","uid":"b0184e68-5ba6-4533-b3fd-bde9416ad03d","resourceVersion":"565","creationTimestamp":"2021-12-28T01:06:35Z","annotations":{"pv.kubernetes.io/bound-by-controller":"yes"},"managedFields":[{"manager":"simulator","operation":"Apply","apiVersion":"v1","time":"2021-12-28T01:06:35Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:accessModes":{},"f:capacity":{"f:storage":{}},"f:hostPath":{"f:path":{},"f:type":{}},"f:persistentVolumeReclaimPolicy":{},"f:volumeMode":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:35Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:phase":{}}},"subresource":"status"},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:pv.kubernetes.io/bound-by-controller":{}}},"f:spec":{"f:claimRef":{}}}}]},"spec":{"capacity":{"storage":"1Gi"},"hostPath":{"path":"/tmp/data","type":"DirectoryOrCreate"},"accessModes":["ReadWriteOnce"],"claimRef":{"kind":"PersistentVolumeClaim","namespace":"default","name":"pvc1","uid":"fb6d1964-41e3-4541-a200-4d76f62b2254","apiVersion":"v1","resourceVersion":"557"},"persistentVolumeReclaimPolicy":"Delete","volumeMode":"Filesystem"}}]`
				j += `,"pvcs":[{"metadata":{"name":"pvc1","namespace":"default","uid":"fb6d1964-41e3-4541-a200-4d76f62b2254","resourceVersion":"567","creationTimestamp":"2021-12-28T01:06:32Z","annotations":{"pv.kubernetes.io/bind-completed":"yes","pv.kubernetes.io/bound-by-controller":"yes"},"managedFields":[{"manager":"simulator","operation":"Apply","apiVersion":"v1","time":"2021-12-28T01:06:32Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:accessModes":{},"f:resources":{"f:requests":{"f:storage":{}}},"f:volumeMode":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:pv.kubernetes.io/bind-completed":{},"f:pv.kubernetes.io/bound-by-controller":{}}},"f:spec":{"f:volumeName":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:accessModes":{},"f:capacity":{".":{},"f:storage":{}},"f:phase":{}}},"subresource":"status"}]},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"volumeName":"pv1","volumeMode":"Filesystem"},"status":{"phase":"Bound","accessModes":["ReadWriteOnce"],"capacity":{"storage":"1Gi"}}}],"storageClasses":[],"priorityClasses":[],"schedulerConfig":{"parallelism":16,"leaderElection":{"leaderElect":true,"leaseDuration":"15s","renewDeadline":"10s","retryPeriod":"2s","resourceLock":"leases","resourceName":"kube-scheduler","resourceNamespace":"kube-system"},"clientConnection":{"kubeconfig":"","acceptContentTypes":"","contentType":"application/vnd.kubernetes.protobuf","qps":50,"burst":100},"healthzBindAddress":"0.0.0.0:10251","metricsBindAddress":"0.0.0.0:10251","enableProfiling":true,"enableContentionProfiling":true,"percentageOfNodesToScore":0,"podInitialBackoffSeconds":1,"podMaxBackoffSeconds":10,"profiles":[{"schedulerName":"default-scheduler","plugins":{"queueSort":{"enabled":[{"name":"PrioritySort"}]},"preFilter":{"enabled":[{"name":"NodeResourcesFit"},{"name":"NodePorts"},{"name":"VolumeRestrictions"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"},{"name":"VolumeBinding"},{"name":"NodeAffinity"}]},"filter":{"enabled":[{"name":"NodeUnschedulable"},{"name":"NodeName"},{"name":"TaintToleration"},{"name":"NodeAffinity"},{"name":"NodePorts"},{"name":"NodeResourcesFit"},{"name":"VolumeRestrictions"},{"name":"EBSLimits"},{"name":"GCEPDLimits"},{"name":"NodeVolumeLimits"},{"name":"AzureDiskLimits"},{"name":"VolumeBinding"},{"name":"VolumeZone"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"}]},"postFilter":{"enabled":[{"name":"DefaultPreemption"}]},"preScore":{"enabled":[{"name":"InterPodAffinity"},{"name":"PodTopologySpread"},{"name":"TaintToleration"},{"name":"NodeAffinity"}]},"score":{"enabled":[{"name":"NodeResourcesBalancedAllocation","weight":1},{"name":"ImageLocality","weight":1},{"name":"InterPodAffinity","weight":1},{"name":"NodeResourcesFit","weight":1},{"name":"NodeAffinity","weight":1},{"name":"PodTopologySpread","weight":2},{"name":"TaintToleration","weight":1}]},"reserve":{"enabled":[{"name":"VolumeBinding"}]},"permit":{},"preBind":{"enabled":[{"name":"VolumeBinding"}]},"bind":{"enabled":[{"name":"DefaultBinder"}]},"postBind":{}},"pluginConfig":[{"name":"DefaultPreemption","args":{"kind":"DefaultPreemptionArgs","apiVersion":"kubescheduler.config.k8s.io/v1","minCandidateNodesPercentage":10,"minCandidateNodesAbsolute":100}},{"name":"InterPodAffinity","args":{"kind":"InterPodAffinityArgs","apiVersion":"kubescheduler.config.k8s.io/v1","hardPodAffinityWeight":1}},{"name":"NodeAffinity","args":{"kind":"NodeAffinityArgs","apiVersion":"kubescheduler.config.k8s.io/v1"}},{"name":"NodeResourcesBalancedAllocation","args":{"kind":"NodeResourcesBalancedAllocationArgs","apiVersion":"kubescheduler.config.k8s.io/v1","resources":[{"name":"cpu","weight":1},{"name":"memory","weight":1}]}},{"name":"NodeResourcesFit","args":{"kind":"NodeResourcesFitArgs","apiVersion":"kubescheduler.config.k8s.io/v1","scoringStrategy":{"type":"LeastAllocated","resources":[{"name":"cpu","weight":1},{"name":"memory","weight":1}]}}},{"name":"PodTopologySpread","args":{"kind":"PodTopologySpreadArgs","apiVersion":"kubescheduler.config.k8s.io/v1","defaultingType":"System"}},{"name":"VolumeBinding","args":{"kind":"VolumeBindingArgs","apiVersion":"kubescheduler.config.k8s.io/v1","bindTimeoutSeconds":600}}]}]}}`
				b := []byte(j)
				r := ResourcesForLoad{}
				if err := json.Unmarshal(b, &r); err != nil {
					panic(err)
				}
				return &r
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{}, defaultApplyFuncs)
				c.PrependReactor("get", "persistentvolumeclaims", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					t.Fatal("This will not be called.")
					return false, nil, nil
				})
				return c
			},
		},
		{
			name: "success load when pv.Status.Phase is not exist",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				j := `{"pods":[],"nodes":[],`
				// delete Phase key&value
				j += `"pvs":[{"metadata":{"name":"pv1","uid":"b0184e68-5ba6-4533-b3fd-bde9416ad03d","resourceVersion":"565","creationTimestamp":"2021-12-28T01:06:35Z","annotations":{"pv.kubernetes.io/bound-by-controller":"yes"},"managedFields":[{"manager":"simulator","operation":"Apply","apiVersion":"v1","time":"2021-12-28T01:06:35Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:accessModes":{},"f:capacity":{"f:storage":{}},"f:hostPath":{"f:path":{},"f:type":{}},"f:persistentVolumeReclaimPolicy":{},"f:volumeMode":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:35Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:phase":{}}},"subresource":"status"},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:pv.kubernetes.io/bound-by-controller":{}}},"f:spec":{"f:claimRef":{}}}}]},"spec":{"capacity":{"storage":"1Gi"},"hostPath":{"path":"/tmp/data","type":"DirectoryOrCreate"},"accessModes":["ReadWriteOnce"],"claimRef":{"kind":"PersistentVolumeClaim","namespace":"default","name":"pvc1","uid":"fb6d1964-41e3-4541-a200-4d76f62b2254","apiVersion":"v1","resourceVersion":"557"},"persistentVolumeReclaimPolicy":"Delete","volumeMode":"Filesystem"},"status":{}}]`
				j += `,"pvcs":[{"metadata":{"name":"pvc1","namespace":"default","uid":"fb6d1964-41e3-4541-a200-4d76f62b2254","resourceVersion":"567","creationTimestamp":"2021-12-28T01:06:32Z","annotations":{"pv.kubernetes.io/bind-completed":"yes","pv.kubernetes.io/bound-by-controller":"yes"},"managedFields":[{"manager":"simulator","operation":"Apply","apiVersion":"v1","time":"2021-12-28T01:06:32Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:accessModes":{},"f:resources":{"f:requests":{"f:storage":{}}},"f:volumeMode":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:pv.kubernetes.io/bind-completed":{},"f:pv.kubernetes.io/bound-by-controller":{}}},"f:spec":{"f:volumeName":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T01:06:36Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:accessModes":{},"f:capacity":{".":{},"f:storage":{}},"f:phase":{}}},"subresource":"status"}]},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"volumeName":"pv1","volumeMode":"Filesystem"},"status":{"phase":"Bound","accessModes":["ReadWriteOnce"],"capacity":{"storage":"1Gi"}}}],"storageClasses":[],"priorityClasses":[],"schedulerConfig":{"parallelism":16,"leaderElection":{"leaderElect":true,"leaseDuration":"15s","renewDeadline":"10s","retryPeriod":"2s","resourceLock":"leases","resourceName":"kube-scheduler","resourceNamespace":"kube-system"},"clientConnection":{"kubeconfig":"","acceptContentTypes":"","contentType":"application/vnd.kubernetes.protobuf","qps":50,"burst":100},"healthzBindAddress":"0.0.0.0:10251","metricsBindAddress":"0.0.0.0:10251","enableProfiling":true,"enableContentionProfiling":true,"percentageOfNodesToScore":0,"podInitialBackoffSeconds":1,"podMaxBackoffSeconds":10,"profiles":[{"schedulerName":"default-scheduler","plugins":{"queueSort":{"enabled":[{"name":"PrioritySort"}]},"preFilter":{"enabled":[{"name":"NodeResourcesFit"},{"name":"NodePorts"},{"name":"VolumeRestrictions"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"},{"name":"VolumeBinding"},{"name":"NodeAffinity"}]},"filter":{"enabled":[{"name":"NodeUnschedulable"},{"name":"NodeName"},{"name":"TaintToleration"},{"name":"NodeAffinity"},{"name":"NodePorts"},{"name":"NodeResourcesFit"},{"name":"VolumeRestrictions"},{"name":"EBSLimits"},{"name":"GCEPDLimits"},{"name":"NodeVolumeLimits"},{"name":"AzureDiskLimits"},{"name":"VolumeBinding"},{"name":"VolumeZone"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"}]},"postFilter":{"enabled":[{"name":"DefaultPreemption"}]},"preScore":{"enabled":[{"name":"InterPodAffinity"},{"name":"PodTopologySpread"},{"name":"TaintToleration"},{"name":"NodeAffinity"}]},"score":{"enabled":[{"name":"NodeResourcesBalancedAllocation","weight":1},{"name":"ImageLocality","weight":1},{"name":"InterPodAffinity","weight":1},{"name":"NodeResourcesFit","weight":1},{"name":"NodeAffinity","weight":1},{"name":"PodTopologySpread","weight":2},{"name":"TaintToleration","weight":1}]},"reserve":{"enabled":[{"name":"VolumeBinding"}]},"permit":{},"preBind":{"enabled":[{"name":"VolumeBinding"}]},"bind":{"enabled":[{"name":"DefaultBinder"}]},"postBind":{}},"pluginConfig":[{"name":"DefaultPreemption","args":{"kind":"DefaultPreemptionArgs","apiVersion":"kubescheduler.config.k8s.io/v1","minCandidateNodesPercentage":10,"minCandidateNodesAbsolute":100}},{"name":"InterPodAffinity","args":{"kind":"InterPodAffinityArgs","apiVersion":"kubescheduler.config.k8s.io/v1","hardPodAffinityWeight":1}},{"name":"NodeAffinity","args":{"kind":"NodeAffinityArgs","apiVersion":"kubescheduler.config.k8s.io/v1"}},{"name":"NodeResourcesBalancedAllocation","args":{"kind":"NodeResourcesBalancedAllocationArgs","apiVersion":"kubescheduler.config.k8s.io/v1","resources":[{"name":"cpu","weight":1},{"name":"memory","weight":1}]}},{"name":"NodeResourcesFit","args":{"kind":"NodeResourcesFitArgs","apiVersion":"kubescheduler.config.k8s.io/v1","scoringStrategy":{"type":"LeastAllocated","resources":[{"name":"cpu","weight":1},{"name":"memory","weight":1}]}}},{"name":"PodTopologySpread","args":{"kind":"PodTopologySpreadArgs","apiVersion":"kubescheduler.config.k8s.io/v1","defaultingType":"System"}},{"name":"VolumeBinding","args":{"kind":"VolumeBindingArgs","apiVersion":"kubescheduler.config.k8s.io/v1","bindTimeoutSeconds":600}}]}]}}`
				b := []byte(j)
				r := ResourcesForLoad{}
				if err := json.Unmarshal(b, &r); err != nil {
					panic(err)
				}
				return &r
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{}, defaultApplyFuncs)
				c.PrependReactor("get", "persistentvolumeclaims", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					t.Fatal("This will not be called.")
					return false, nil, nil
				})
				return c
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockSchedulerSve := mock_snapshot.NewMockSchedulerService(ctrl)
			c := tt.prepareFakeClientSetFn()

			s := NewService(c, mockSchedulerSve)
			tt.prepareEachServiceMockFn(mockSchedulerSve)

			err := s.Load(context.Background(), tt.applyConfiguration())
			if err != nil && tt.wantErr == nil {
				t.Fatalf("Expect no error, but Load() returns error = %+v", err)
			}
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error(), "expected: %+v\nactual: %+v", tt.wantErr, err)
			}
		})
	}
}

func TestService_Load_WithIgnoreErrOption(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		prepareEachServiceMockFn func(*mock_snapshot.MockSchedulerService)
		prepareFakeClientSetFn   func() *fake.Clientset
		applyConfiguration       func() *ResourcesForLoad
		wantErr                  error
	}{
		{
			name: "load all success",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("StorageClass1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{}, defaultApplyFuncs)
				return c
			},
		},
		{
			name: "no error if failure to apply Pod",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("something wrong", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("StorageClass1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{
					pod: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("patch", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to apply Pod")
						})
					},
				}, defaultApplyFuncs)
				return c
			},
		},
		{
			name: "no error if failure to apply Node",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("something wrong"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("StorageClass1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{
					node: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("patch", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to apply Node")
						})
					},
				}, defaultApplyFuncs)
				return c
			},
		},
		{
			name: "no error if failed to apply PersistentVolume",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("something wrong").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("StorageClass1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{
					pv: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("patch", "persistentvlumes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to apply PersistentVolume")
						})
					},
				}, defaultApplyFuncs)
				return c
			},
		},
		{
			name: "no error if failed to apply PersistentVolumeClaim",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("something wrong", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("StorageClass1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{
					pvc: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("patch", "persistentvlumeclaims", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to apply PersistentVolumeClaims")
						})
					},
				}, defaultApplyFuncs)
				return c
			},
		},
		{
			name: "no error if failed to apply StorageClass",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("something wrong"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{
					sc: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("patch", "storageclasses", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to apply StorageClass")
						})
					},
				}, defaultApplyFuncs)
				return c
			},
		},
		{
			name: "no error if failed to apply PriorityClass",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("something wrong"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{
					pc: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("patch", "priorityclasses", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to apply PriorityClass")
						})
					},
				}, defaultApplyFuncs)
				return c
			},
		},
		{
			name: "no error if failed to apply Namespace",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				ss.EXPECT().RestartScheduler(gomock.Any()).Return(nil)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace("something wrong"),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("SC1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{
					ns: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("patch", "namespaces", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, nil, xerrors.New("failed to apply Namespace")
						})
					},
				}, defaultApplyFuncs)
				return c
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockSchedulerSve := mock_snapshot.NewMockSchedulerService(ctrl)
			c := tt.prepareFakeClientSetFn()

			s := NewService(c, mockSchedulerSve)
			tt.prepareEachServiceMockFn(mockSchedulerSve)

			err := s.Load(context.Background(), tt.applyConfiguration(), s.IgnoreErr())
			if err != nil && tt.wantErr == nil {
				t.Fatalf("Expect no error, but Load() returns error = %+v", err)
			}
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error(), "expected: %+v\nactual: %+v", tt.wantErr, err)
			}
		})
	}
}

func TestFunction_listPcs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		prepareFakeClientSetFn func() *fake.Clientset
		wantReturn             func() *ResourcesForSnap
		wantErr                bool
	}{
		{
			name: "all pc which have name prefixed with `system-` should filter out",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				ctx := context.Background()
				// add test data.
				invokeResourcesFn(ctx, c, SettingClientFuncMap{
					pc: func(ctx context.Context, c *fake.Clientset) {
						c.PrependReactor("list", "priorityclasses", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
							return true, &schedulingv1.PriorityClassList{
								Items: []schedulingv1.PriorityClass{
									{
										ObjectMeta: metav1.ObjectMeta{
											Name: "system-cluster-critical",
										},
									},
									{
										ObjectMeta: metav1.ObjectMeta{
											Name: "priority-class1",
										},
									},
								},
							}, nil
						})
					},
				}, defaultFuncs)
				return c
			},
			wantReturn: func() *ResourcesForSnap {
				_pcs := []schedulingv1.PriorityClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "priority-class1",
						},
					},
				}
				return &ResourcesForSnap{
					Pods:            []corev1.Pod{},
					Nodes:           []corev1.Node{},
					Pvs:             []corev1.PersistentVolume{},
					Pvcs:            []corev1.PersistentVolumeClaim{},
					StorageClasses:  []storagev1.StorageClass{},
					PriorityClasses: _pcs,
					SchedulerConfig: &configv1.KubeSchedulerConfiguration{},
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockSchedulerSve := mock_snapshot.NewMockSchedulerService(ctrl)
			c := tt.prepareFakeClientSetFn()

			s := NewService(c, mockSchedulerSve)

			errgrp := util.NewErrGroupWithSemaphore(context.Background())
			resources := &ResourcesForSnap{
				Pods:            []corev1.Pod{},
				Nodes:           []corev1.Node{},
				Pvs:             []corev1.PersistentVolume{},
				Pvcs:            []corev1.PersistentVolumeClaim{},
				StorageClasses:  []storagev1.StorageClass{},
				PriorityClasses: []schedulingv1.PriorityClass{},
				SchedulerConfig: &configv1.KubeSchedulerConfiguration{},
			}

			err := s.listPcs(context.Background(), resources, errgrp, options{})
			if err := errgrp.Wait(); err != nil {
				t.Fatalf("listPcs: %v", err)
			}
			diffResponse := cmp.Diff(resources, tt.wantReturn())
			if diffResponse != "" || (err != nil) != tt.wantErr {
				t.Fatalf("listPcs() %v test, \nerror = %v, wantErr %v\n%s", tt.name, err, tt.wantErr, diffResponse)
			}
		})
	}
}

func TestFunction_applyPcs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		prepareFakeClientSetFn func() *fake.Clientset
		applyConfiguration     func() *ResourcesForLoad
		wantErr                bool
	}{
		{
			name: "all pc which have name prefixed with `system-` should filter out",
			applyConfiguration: func() *ResourcesForLoad {
				pods := []v1.PodApplyConfiguration{}
				nodes := []v1.NodeApplyConfiguration{}
				pvs := []v1.PersistentVolumeApplyConfiguration{}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{}
				storageclasses := []confstoragev1.StorageClassApplyConfiguration{}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{}
				pcs = append(pcs, *schedulingcfgv1.PriorityClass("system-PriorityClass1"))
				pcs = append(pcs, *schedulingcfgv1.PriorityClass("PriorityClass1"))
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  storageclasses,
					PriorityClasses: pcs,
					SchedulerConfig: config,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				c.PrependReactor("patch", "priorityclasses", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					a, ok := action.(k8stesting.PatchAction)
					assert.Equal(t, true, ok)
					// High priority PriorityClass will not come.
					assert.Equal(t, true, !strings.HasPrefix(a.GetName(), "system-"))
					return true, nil, nil
				})
				return c
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ctrl := gomock.NewController(t)

			c := tt.prepareFakeClientSetFn()
			mockSchedulerSve := mock_snapshot.NewMockSchedulerService(ctrl)
			s := NewService(c, mockSchedulerSve)

			errgrp := util.NewErrGroupWithSemaphore(ctx)
			err := s.applyPcs(ctx, tt.applyConfiguration(), errgrp, options{})
			if err := errgrp.Wait(); err != nil {
				t.Fatalf("applyPcs: %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("applyPcs() %v test, \nerror = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestService_Load_WithIgnoreSchedulerConfigurationOption(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		prepareEachServiceMockFn func(*mock_snapshot.MockSchedulerService)
		prepareFakeClientSetFn   func() *fake.Clientset
		applyConfiguration       func() *ResourcesForLoad
		wantErr                  error
	}{
		{
			name: "Load does not call RestartScheduler",
			prepareEachServiceMockFn: func(ss *mock_snapshot.MockSchedulerService) {
				// If the Load function call this, it will return the error.
				// RestartScheduler must be called zero times.
				ss.EXPECT().RestartScheduler(gomock.Any()).Times(0)
			},
			applyConfiguration: func() *ResourcesForLoad {
				ns := []v1.NamespaceApplyConfiguration{
					*v1.Namespace(testNamespace1),
				}
				pods := []v1.PodApplyConfiguration{
					*v1.Pod("Pod1", testNamespace1),
				}
				nodes := []v1.NodeApplyConfiguration{
					*v1.Node("Node1"),
				}
				pvs := []v1.PersistentVolumeApplyConfiguration{
					*v1.PersistentVolume("PV1").WithStatus(v1.PersistentVolumeStatus().WithPhase("Pending")).WithUID("test"),
				}
				pvcs := []v1.PersistentVolumeClaimApplyConfiguration{
					*v1.PersistentVolumeClaim("PVC1", testNamespace1),
				}
				scs := []confstoragev1.StorageClassApplyConfiguration{
					*confstoragev1.StorageClass("StorageClass1"),
				}
				pcs := []schedulingcfgv1.PriorityClassApplyConfiguration{
					*schedulingcfgv1.PriorityClass("PriorityClass1"),
				}
				config, _ := schedulerCfg.DefaultSchedulerConfig()
				return &ResourcesForLoad{
					Pods:            pods,
					Nodes:           nodes,
					Pvs:             pvs,
					Pvcs:            pvcs,
					StorageClasses:  scs,
					PriorityClasses: pcs,
					SchedulerConfig: config,
					Namespaces:      ns,
				}
			},
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				invokeResourcesFn(context.Background(), c, SettingClientFuncMap{}, defaultApplyFuncs)
				return c
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockSchedulerSve := mock_snapshot.NewMockSchedulerService(ctrl)
			c := tt.prepareFakeClientSetFn()

			s := NewService(c, mockSchedulerSve)
			tt.prepareEachServiceMockFn(mockSchedulerSve)

			if err := s.Load(context.Background(), tt.applyConfiguration(), s.IgnoreSchedulerConfiguration()); (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("Load() with ignoreSchedulerConfiguration option, %v test, \nerror = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
