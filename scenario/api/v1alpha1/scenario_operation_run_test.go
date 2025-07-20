package v1alpha1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const ns = "default"

func podToUnstructured(p *corev1.Pod) *unstructured.Unstructured {
	objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(p)
	Expect(err).NotTo(HaveOccurred())
	return &unstructured.Unstructured{Object: objMap}
}

var _ = Describe("ScenarioOperation.Run()", func() {

	It("executes CreateOperation", func() {
		pod := &corev1.Pod{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
			ObjectMeta: metav1.ObjectMeta{Name: "run-create-pod", Namespace: ns},
			Spec: corev1.PodSpec{Containers: []corev1.Container{
				{Name: "pause", Image: "registry.k8s.io/pause:3.9"},
			}},
		}

		op := &ScenarioOperation{
			ID: "create-" + string(uuid.NewUUID()),
			Create: &CreateOperation{
				Object: podToUnstructured(pod),
			},
		}

		done, err := op.Run(context.TODO(), cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(done).To(BeFalse())

		Eventually(func() error {
			return k8sCli.Get(context.TODO(),
				types.NamespacedName{Name: "run-create-pod", Namespace: ns},
				&corev1.Pod{})
		}).Should(Succeed())
	})

	It("executes PatchOperation", func() {
		base := &corev1.Pod{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
			ObjectMeta: metav1.ObjectMeta{Name: "run-patch-pod", Namespace: ns},
			Spec: corev1.PodSpec{Containers: []corev1.Container{
				{Name: "pause", Image: "registry.k8s.io/pause:3.9"},
			}},
		}
		Expect(k8sCli.Create(context.TODO(), base)).To(Succeed())

		op := &ScenarioOperation{
			ID: "patch-" + string(uuid.NewUUID()),
			Patch: &PatchOperation{
				TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
				ObjectMeta: metav1.ObjectMeta{Name: "run-patch-pod", Namespace: ns},
				PatchType:  types.MergePatchType,
				Patch:      `{"metadata":{"labels":{"patched":"true"}}}`,
			},
		}

		_, err := op.Run(context.TODO(), cfg)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() map[string]string {
			tmp := &corev1.Pod{}
			_ = k8sCli.Get(context.TODO(),
				types.NamespacedName{Name: "run-patch-pod", Namespace: ns}, tmp)
			return tmp.Labels
		}).Should(HaveKeyWithValue("patched", "true"))
	})

	It("executes DeleteOperation", func() {
		p := &corev1.Pod{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
			ObjectMeta: metav1.ObjectMeta{Name: "run-delete-pod", Namespace: ns},
			Spec: corev1.PodSpec{Containers: []corev1.Container{
				{Name: "pause", Image: "registry.k8s.io/pause:3.9"},
			}},
		}
		Expect(k8sCli.Create(context.TODO(), p)).To(Succeed())

		op := &ScenarioOperation{
			ID: "delete-" + string(uuid.NewUUID()),
			Delete: &DeleteOperation{
				TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
				ObjectMeta: metav1.ObjectMeta{Name: "run-delete-pod", Namespace: ns},
			},
		}

		_, err := op.Run(context.TODO(), cfg)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() error {
			return k8sCli.Get(context.TODO(),
				types.NamespacedName{Name: "run-delete-pod", Namespace: ns},
				&corev1.Pod{})
		}).Should(MatchError(ContainSubstring("not found")))
	})

	It("executes DoneOperation", func() {
		op := &ScenarioOperation{
			ID:   "done-" + string(uuid.NewUUID()),
			Done: &DoneOperation{},
		}
		done, err := op.Run(context.TODO(), cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(done).To(BeTrue())
	})
})
