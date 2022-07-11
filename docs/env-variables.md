# Different Environment Variables used

This page describes about the different environment variables that are
used to configure kube-scheduler-simulator. 

Please refer [docker-compose.yml](./docker-compose.yml) as an example
use.

`Port`: (required) This is the port number on which kube-scheduler-simulator
server is started. It's default value is `1212`.

`EtcdURL`: (required) This is the URL for etcd. The simulator runs kube-apiserver
internally, and the kube-apiserver uses this etcd. It's default value is `http://simulator-etcd:2379`.

`FrontendURL`: This URL represents the URL web UI started. The
simulator and internal kube-apiserver set the origin as the allowed
origin for `CORS_ALLOWED_ORIGIN_LIST`.

`KUBECONFIG`: (required) This is for the beta feature "Existing cluster
Importing". This variable is used to find Kubeconfig required to
access your cluster for importing resources to scheduler simulator.

`KUBE_API_HOST`: (required) This is the host of kube-apiserver which the
simulator starts internally.

`KUBE_SCHEDULER_CONFIG_PATH`: (required) The path to a KubeSchedulerConfiguration
file.  If passed, the simulator will start the scheduler with that
configuration.  Or, if you use web UI, you can change the
configuration from the web UI as well.

`ExternalImportEnabled`: This variable indicates whether the simulator will import resources from an existing cluster or not.

## For Web UI

`KUBE_API_SERVER_URL`: This is the kube-apiserver URL and it's value is set [here](https://github.com/kubernetes-sigs/kube-scheduler-simulator/blob/master/docker-compose.yml#L26)
