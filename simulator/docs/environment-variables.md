## [deprecated] Environment Variables

**Deprecation notice**: We're planning to remove the configuration via environment variables.
Until deprecation, the simulator will read the configuration in the environment variable first,
if the environment variable is not set, it will read the configuration in the configuration file.
For config file, please refer to the simulator [config.yaml](./../config.yaml).

> [!WARNING]
> To reduce the maintenance burden, the simulator no longer run kube-apiserver internally 
> and we made the simulator require the kube-apiserver outside.
> `KUBE_API_HOST` and `KUBE_API_PORT` are deprecated and will be removed in the near future, meanwhile they're just ignored.

---

This page describes the environment variables that are used to configure the simulator.

Please refer to [docker-compose.yml](./../../docker-compose.yml) as an example use.

### For Simulator

`PORT`: (required) This is the port number on which kube-scheduler-simulator
server is started.

`KUBE_SCHEDULER_SIMULATOR_ETCD_URL`: (required) This is the URL for
etcd. The simulator runs kube-apiserver internally, and the
kube-apiserver uses this etcd.

`CORS_ALLOWED_ORIGIN_LIST`: This URL represents the URL once web UI is
started. The simulator and internal kube-apiserver set the allowed
origin for `CORS_ALLOWED_ORIGIN_LIST`.

`KUBECONFIG`: This is for the beta feature "Importing cluster's 
resources". This variable is used to find Kubeconfig required to
access your cluster for importing resources to scheduler simulator.

`KUBE_APISEVER_URL`: This is the URL of kube-apiserver which the
simulator uses. This variable is used to connect to external kube-apiserver.

`KUBE_API_HOST`: This is the host of kube-apiserver which the
simulator starts internally. Its default value is `127.0.0.1`.
This variable is deprecated and will be removed in the near future, but you can use `KUBE_APISEVER_URL` instead.

`KUBE_API_PORT`: This is the port of kube-apiserver. Its default
value is `3131`.
This variable is deprecated and will be removed in the near future, but you can use `KUBE_APISEVER_URL` instead.

`KUBE_SCHEDULER_CONFIG_PATH`: The path to a KubeSchedulerConfiguration
file.  If passed, the simulator will start the scheduler with that
configuration. Or, if you use web UI, you can change the
configuration from the web UI as well.

`EXTERNAL_IMPORT_ENABLED`: This variable indicates whether the simulator
will import resources from an user cluster's or not.
Note, this is still a beta feature.
