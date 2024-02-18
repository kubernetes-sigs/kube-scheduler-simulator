package syncer_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/e2e-framework/support/kwok"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/app"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/config"
)

func Test_SyncNodesAndPods(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cluster := kwok.NewCluster("kwok").WithVersion("v0.5.0")
	_, err := cluster.Create(ctx,
		"--disable-kube-scheduler",
	)
	if err != nil {
		t.Fatalf("failed to create kwok cluster: %v", err)
	}

	defer func() {
		err := cluster.Destroy(ctx)
		if err != nil {
			t.Logf("failed to destroy kwok cluster: %v", err)
		}
	}()

	realclient, err := kubernetes.NewForConfig(cluster.KubernetesRestConfig())
	if err != nil {
		t.Fatalf("failed to initiate client: %v", err)
	}

	// create 5 nodes
	for i := 0; i < 5; i++ {
		_, err = realclient.CoreV1().Nodes().Create(ctx, &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("node-%v", i),
			},
		}, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("failed to create node: %v", err)
		}
	}

	// create 5 unscheduled pods
	for i := 0; i < 5; i++ {
		_, err = realclient.CoreV1().Pods("default").Create(ctx, &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("pod-%v", i),
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "app",
						Image: "fake",
					},
				},
			},
		}, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("failed to create unscheduled pod: %v", err)
		}
	}

	// create 5 scheduled pods
	for i := 0; i < 5; i++ {
		_, err = realclient.CoreV1().Pods("default").Create(ctx, &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("scheduled-pod-%v", i),
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "app",
						Image: "fake",
					},
				},
				NodeName: fmt.Sprintf("node-%v", i),
			},
		}, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("failed to create scheduled pod: %v", err)
		}
	}

	// simulatorCluster is the simulator's cluster
	simulatorCluster := kwok.NewCluster("kwok-simulator").WithVersion("v0.5.0")
	_, err = simulatorCluster.Create(ctx,
		"--disable-kube-scheduler",
		"--etcd-prefix=/kube-scheduler-simulator",
		"--etcd-port=2379",
	)
	if err != nil {
		t.Fatalf("failed to create simulator's kwok cluster: %v", err)
	}
	defer func() {
		err := simulatorCluster.Destroy(ctx)
		if err != nil {
			t.Logf("failed to destroy simulator's kwok cluster: %v", err)
		}
	}()

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("failed to get simulator config: %v", err)
	}
	cfg.ResourceSyncEnabled = true
	cfg.ExternalKubeClientCfg = cluster.KubernetesRestConfig()
	cfg.KubeAPIServerURL = simulatorCluster.KubernetesRestConfig().Host
	cfg.EtcdURL = "http://localhost:2379"

	restCfg := simulatorCluster.KubernetesRestConfig()

	simulatorclient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		t.Fatalf("failed to initiate client: %v", err)
	}

	cr, err := simulatorclient.RbacV1().ClusterRoles().Create(ctx, &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{GenerateName: "annoymoususerrole"},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs: []string{
					"*",
				},
				APIGroups: []string{
					"*",
				},
				Resources: []string{
					"*",
				},
			}, {
				NonResourceURLs: []string{
					"*",
				},
				Verbs: []string{
					"*",
				},
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("")
	}

	_, err = simulatorclient.RbacV1().ClusterRoleBindings().Create(ctx, &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{GenerateName: "annoymoususerrolebinding"},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.UserKind,
				APIGroup:  rbacv1.GroupName,
				Name:      user.Anonymous,
				Namespace: "",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     cr.Name,
		},
	}, metav1.CreateOptions{})

	if err != nil {
		t.Fatalf("")
	}

	go func() {
		err = app.StartSimulator(ctx, cfg)
		if err != nil {
			t.Fatalf("failed to initiate the simulator: %v", err)
		}
	}()

	// wait for the syncer to finish to sync resources.
	time.Sleep(20 * time.Second)

	nodes, err := simulatorclient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("failed to list the nodes: %v", err)
	}
	if len(nodes.Items) != 5 {
		t.Fatalf("5 nodes should be created, but %v nodes are created", len(nodes.Items))
	}
	pods, err := simulatorclient.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("failed to list the pods: %v", err)
	}
	if len(pods.Items) != 5 {
		t.Fatalf("5 pods should be created, but %v pods are created", len(pods.Items))
	}
}
