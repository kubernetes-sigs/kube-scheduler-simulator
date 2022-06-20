package priorityclass

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	schedulingv1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8sTesting "k8s.io/client-go/testing"
)

// deleteCollectionReaction returns a ReactionFunc that applies DeleteCollectionAction
// to the given tracker.
func deleteCollectionReaction(tracker k8sTesting.ObjectTracker) k8sTesting.ReactionFunc {
	return func(action k8sTesting.Action) (bool, runtime.Object, error) {
		ns := action.GetNamespace()
		gvr := action.GetResource()
		gvk := action.GetResource().GroupVersion().WithKind("PriorityClass")

		switch action := action.(type) {

		case k8sTesting.DeleteCollectionActionImpl:
			obj, err := tracker.List(gvr, gvk, ns)
			list := obj.(*schedulingv1.PriorityClassList)
			for _, class := range list.Items {
				nameField := fields.Set{"metadata.name": class.Name}
				if action.GetListRestrictions().Fields.Matches(nameField) {
					err := tracker.Delete(gvr, ns, class.Name)
					if err != nil {
						return true, nil, err
					}
				}
			}
			return true, obj, err

		default:
			return false, nil, fmt.Errorf("no reaction implemented for %s", action)
		}
	}
}

func addReactorForPriorityClasses(c *fake.Clientset) {
	c.AddReactor("delete-collection", "priorityclasses", deleteCollectionReaction(c.Tracker()))
}

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
				addReactorForPriorityClasses(c)
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
				addReactorForPriorityClasses(c)
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
				addReactorForPriorityClasses(c)
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
