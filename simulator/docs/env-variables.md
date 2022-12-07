# Different Environment Variables used

This page describes about the different environment variables that are
used to configure kube-scheduler-simulator.

Please refer [docker-compose.yml](./../../docker-compose.yml) as an example
use.

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

## For Web UI

This part is deprecated. Please refer to the config file [user.config.js](./../../web/user.config.js).

`KUBE_API_SERVER_URL`: This is the kube-apiserver URL. Its default
value is `http://localhost:3131`.

`BASE_URL`: This is the URL for the kube-scheduler-simulator
server. Its default value is `http://localhost:1212`.

`ALPHA_TABLE_VIEWS`: This variable enables the alpha feature `table
view`. Its value is either 0(default) or 1 (0 meaning disalbed, 1
meaning enabled). We can see the resource status in the table.
