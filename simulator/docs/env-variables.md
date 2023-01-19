# Environment Variables

This page describes the environment variables that are used to configure the simulator.

Please refer [docker-compose.yml](./../../docker-compose.yml) as an example use.

`PORT`: (required) This is the port number on which kube-scheduler-simulator
server is started.

`KUBE_SCHEDULER_SIMULATOR_ETCD_URL`: (required) This is the URL for
etcd. The simulator runs kube-apiserver internally, and the
kube-apiserver uses this etcd.

`CORS_ALLOWED_ORIGIN_LIST`: This URL represents the URL once web UI is
started. The simulator and internal kube-apiserver set the allowed
origin for `CORS_ALLOWED_ORIGIN_LIST`.

`KUBECONFIG`: This is for the beta feature "Existing cluster
Importing". This variable is used to find Kubeconfig required to
access your cluster for importing resources to scheduler simulator.

`KUBE_API_HOST`: This is the host of kube-apiserver which the
simulator starts internally. Its default value is `127.0.0.1`.

`KUBE_API_PORT`: This is the port of kube-apiserver. Its default
value is `3131`.

`KUBE_SCHEDULER_CONFIG_PATH`: The path to a KubeSchedulerConfiguration
file.  If passed, the simulator will start the scheduler with that
configuration.  Or, if you use web UI, you can change the
configuration from the web UI as well.

`EXTERNAL_IMPORT_ENABLED`: This variable indicates whether the simulator
will import resources from an existing cluster or not. Note, this is
still a beta feature.
