# Simulator server configuration

Simulator server configuration used to only support setting configurations 
through environment variables, and now adds configurations through configuration files. 
The simulator reads the configuration file in the path of [./config.yaml](./../config.yaml).

```
# This is an example config for scheduler-simulator.

apiVersion: kube-scheduler-simulator-config/v1alpha1
kind: SimulatorConfiguration

# This is the port number on which kube-scheduler-simulator
# server is started.
port: 1212

# This is the URL for etcd. The simulator runs kube-apiserver
# internally, and the kube-apiserver uses this etcd.
etcdURL: "http://127.0.0.1:2379"

# This URL represents the URL once web UI is started.
# The simulator and internal kube-apiserver set the allowed
# origin for CorsAllowedOriginList
corsAllowedOriginList:
  - "http://localhost:3000"

# This is for the beta feature "Importing cluster's resources".
# This variable is used to find Kubeconfig required to access your
# cluster for importing resources to scheduler simulator.
kubeConfig: "/kubeconfig.yaml"

# This is the URL of kube-apiserver which the simulator uses.
# This variable is used to connect to external kube-apiserver.
kubeAPIServerURL: ""

# The path to a KubeSchedulerConfiguration file.
# If passed, the simulator will start the scheduler
# with that configuration. Or, if you use web UI,
# you can change the configuration from the web UI as well.
kubeSchedulerConfigPath: ""

# This variable indicates whether the simulator will
# import resources from a user cluster specified by kubeConfig.
# Note that it only imports the resources once when the simulator is started.
# You cannot make both externalImportEnabled and resourceSyncEnabled true because those features would be conflicted.
# This is still a beta feature.
externalImportEnabled: false

# This variable indicates whether the simulator will
# keep syncing resources from an user cluster's or not.
# You cannot make both externalImportEnabled and resourceSyncEnabled true because those features would be conflicted.
# Note, this is still a beta feature.
resourceSyncEnabled: false

# This variable indicates whether an external scheduler
# is used.
externalSchedulerEnabled: false
```
