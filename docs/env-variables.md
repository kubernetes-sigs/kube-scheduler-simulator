# Different Environment Variables used

This page describes about the different environment variables that are
used to configure kube-scheduler-simulator.

`Port`: This is the port number on which kube-scheduler-simulator is started and it's value is 1212 which is set [here](https://github.com/kubernetes-sigs/kube-scheduler-simulator/blob/3111ec419ccb8c49197b75385fbd166f7f159435/docker-compose.yml#L7)

`EtcdURL`: This is the URL for kube-scheduler-simulator etcd which runs locally and it's value is set [here](https://github.com/kubernetes-sigs/kube-scheduler-simulator/blob/3111ec419ccb8c49197b75385fbd166f7f159435/docker-compose.yml#L8)

`FrontendURL`: This URL is on which kube-scheduler-simulator is reachable locally and it's value is set [here](https://github.com/kubernetes-sigs/kube-scheduler-simulator/blob/3111ec419ccb8c49197b75385fbd166f7f159435/docker-compose.yml#L9).

`KUBE_API_HOST`: This is the host on which the kube-apiserver serves
for the simulator server.

`KUBE_API_SERVER_URL`: This is the kube-apiserver URL and it's value is set [here](https://github.com/kubernetes-sigs/kube-scheduler-simulator/blob/master/docker-compose.yml#L26)

`KUBE_SCHEDULER_CONFIG_PATH`: A KubeSchedulerConfiguration file can be
passed via this environment varia
ble and the simulator will start the
scheduler with that configuration.
