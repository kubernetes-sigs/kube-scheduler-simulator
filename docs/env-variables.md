# Different Environment Variables used

This page describes about the different environment variables that are
used to configure kube-scheduler-simulator. 

Please refer [docker-compose.yml](./docker-compose.yml) as an example
use.

`Port`: This is the port number on which kube-scheduler-simulator
server is started.

`EtcdURL`: This is the URL for etcd. The simulator runs kube-apiserver
internally, and the kube-apiserver uses this etcd.

`FrontendURL`: This URL represents the URL web UI started. The
simulator and internal kube-apiserver set the origin as the allowed
origin for `CORS_ALLOWED_ORIGIN_LIST`.

`KUBECONFIG`: This is for the beta feature "Existing cluster
Importing". This variable is used to find Kubeconfig required to
access your cluster for importing resources to scheduler simulator.

`KUBE_API_HOST`: This is the host of kube-apiserver which the
simulator starts internally.

`KUBE_SCHEDULER_CONFIG_PATH`: The path to a KubeSchedulerConfiguration
file.  If passed, the simulator will start the scheduler with that
configuration.  Or, if you use web UI, you can change the
configuration from the web UI as well.

## For Web UI

`KUBE_API_SERVER_URL`: This is the kube-apiserver URL and it's value is set [here](https://github.com/kubernetes-sigs/kube-scheduler-simulator/blob/master/docker-compose.yml#L26)
