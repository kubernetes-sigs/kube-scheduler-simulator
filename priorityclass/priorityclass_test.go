package priorityclass

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	schedulingv1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func createDefaultPriorityClasses(c *fake.Clientset) {
	c.SchedulingV1().PriorityClasses().Create(context.Background(), &schedulingv1.PriorityClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-cluster-critical",
		},
		Value: 2000000000,
	}, metav1.CreateOptions{})
	c.SchedulingV1().PriorityClasses().Create(context.Background(), &schedulingv1.PriorityClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-node-critical",
		},
		Value: 2000001000,
	}, metav1.CreateOptions{})
}

var listOfDefaultPriorityClases = []schedulingv1.PriorityClass{
	{
		ObjectMeta: metav1.ObjectMeta{Name: "system-cluster-critical"},
		Value:      2000000000,
	},
	{
		ObjectMeta: metav1.ObjectMeta{Name: "system-node-critical"},
		Value:      2000001000,
	},
}

func TestService_DeleteCollection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		prepareFakeClientSetFn func() *fake.Clientset
		lopt                   metav1.ListOptions
		wantErr                bool
		wantReturn             *schedulingv1.PriorityClassList
	}{
		{
			name: "delete all priorityclasses",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				createDefaultPriorityClasses(c)
				c.SchedulingV1().PriorityClasses().Create(context.Background(), &schedulingv1.PriorityClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "production",
					},
					Value: 1000000,
				}, metav1.CreateOptions{})
				c.SchedulingV1().PriorityClasses().Create(context.Background(), &schedulingv1.PriorityClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "staging",
					},
					Value: 800000,
				}, metav1.CreateOptions{})
				c.SchedulingV1().PriorityClasses().Create(context.Background(), &schedulingv1.PriorityClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "development",
					},
					Value: 600000,
				}, metav1.CreateOptions{})

				return c
			},
			lopt:    metav1.ListOptions{},
			wantErr: false,
			wantReturn: &schedulingv1.PriorityClassList{
				Items: listOfDefaultPriorityClases,
			},
		},
		{
			name: "delete all priorityclasses, only default classes",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				createDefaultPriorityClasses(c)
				return c
			},
			lopt:    metav1.ListOptions{},
			wantErr: false,
			wantReturn: &schedulingv1.PriorityClassList{
				Items: listOfDefaultPriorityClases,
			},
		},
		{
			name: "delete all priorityclasses with fieldSelector",
			prepareFakeClientSetFn: func() *fake.Clientset {
				c := fake.NewSimpleClientset()
				createDefaultPriorityClasses(c)
				c.SchedulingV1().PriorityClasses().Create(context.Background(), &schedulingv1.PriorityClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "production",
					},
					Value: 1000000,
				}, metav1.CreateOptions{})
				return c
			},
			lopt: metav1.ListOptions{
				FieldSelector: "metadata.name!=production",
			},
			wantErr: false,
			wantReturn: &schedulingv1.PriorityClassList{
				Items: append([]schedulingv1.PriorityClass{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "production"},
						Value:      1000000,
					},
				},
					listOfDefaultPriorityClases...),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fakeclientset := tt.prepareFakeClientSetFn()
			s := NewPriorityClassService(fakeclientset)
			err := s.DeleteCollection(context.Background(), tt.lopt)
			classes, _ := s.List(context.Background())
			diffResponse := cmp.Diff(classes, tt.wantReturn)
			if diffResponse != "" || (err != nil) != tt.wantErr {
				t.Fatalf("DeleteCollection() %s error = %v, wantErr %v\n%s\n%s", tt.name, err, tt.wantErr, classes, diffResponse)
			}
		})
	}
}
