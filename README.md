# Kubernetes scheduler simulator

Hello world. Here is Kubernetes scheduler simulator.

Nowadays, the scheduler is configurable/extendable in the multiple ways:
- configure with [KubeSchedulerConfiguration](https://kubernetes.io/docs/reference/scheduling/config/)
- add Plugins of [Scheduling Framework](https://kubernetes.io/docs/concepts/scheduling-eviction/scheduling-framework/)
- add [Extenders](https://github.com/kubernetes/enhancements/tree/5320deb4834c05ad9fb491dcd361f952727ece3e/keps/sig-scheduling/1819-scheduler-extender)
- etc...

But, unfortunately, not all configurations/expansions yield good results.
Those who customize the scheduler need to make sure their scheduler is working as expected, and doesn't have an unacceptably negative impact on the scheduling. 

In real Kubernetes, we cannot know the results of scheduling in detail without reading the logs, which usually require privileged access to the control plane.
That's way we are developing a simulator for kube-scheduler -- you can try out the behavior of the scheduler with web UI while checking which plugin made what decision for which Node.

## Simulator's architecture

We have several components:
- Simulator (in `/simulator`)
- Web UI (in `/web`)
- Coming soon... :)  (see [./keps](./keps) to see some nice ideas we're working on)

### Simulator

Simulator internally has kube-apiserver, scheduler, and HTTP server.

You can create any resources by communicating with kube-apiserver via kubectl, k8s client library, or web UI.

See the following docs to know more about simulator:
- [how-it-works.md](simulator/docs/how-it-works.md): describes about how the simulator works.
- [kube-apiserver.md](simulator/docs/kube-apiserver.md): describe about kube-apiserver in simulator. (how you can configure and access) 
- [api.md](simulator/docs/api.md): describes about HTTP server the simulator has.

### Web UI

Web UI is one of the clients of simulator, but it's optimized for simulator.

From the web, you can create/edit/delete these resources to simulate a cluster.

- Nodes
- Pods
- Persistent Volumes
- Persistent Volume Claims
- Storage Classes
- Priority Classes

![list resources](simulator/docs/images/resources.png)

You can create resources with yaml file as usual.

![create node](simulator/docs/images/create-node.png)

And, after pods are scheduled, you can see the results of

- Each Filter plugins
- Each Score plugins
- Final score (normalized and applied Plugin Weight)

![result](simulator/docs/images/result.jpg)

You can configure the scheduler on the simulator through KubeSchedulerConfiguration.

[Scheduler Configuration | Kubernetes](https://kubernetes.io/docs/reference/scheduling/config/)

You can pass a KubeSchedulerConfiguration file via the environment variable `KUBE_SCHEDULER_CONFIG_PATH` and the simulator will start kube-scheduler with that configuration.

Note: changes to any fields other than `.profiles` are disabled on simulator, since they do not affect the results of the scheduling.

![configure scheduler](simulator/docs/images/schedulerconfiguration.png)

If you want to use your custom plugins as out-of-tree plugins in the simulator, please follow [this doc](simulator/docs/how-to-use-custom-plugins/README.md).

## Getting started

Read more about environment variables being used in simulator server
[here.](./simulator/docs/env-variables.md)
### Run simulator with Docker

We have [docker-compose.yml](docker-compose.yml) to run the simulator easily.

You can use the following command.

```bash
# build the images for web frontend and simulator server, then start the containers.
make docker_build_and_up
```

Then, you can access the simulator with http://localhost:3000. If you want to deploy the simulator on a remote server and access it via a specific IP (e.g: like http://10.0.0.1:3000/), please make sure that you have executed `export SIMULATOR_EXTERNAL_IP=your.server.ip` before running `docker-compose up -d`.

Note: Insufficient memory allocation may cause problems in building the image.
Please allocate enough memory in that case.

### Run simulator locally

You have to run frontend, server and etcd.

#### Run simulator server and etcd

To run this simulator's server, you have to install Go and etcd.

You can install etcd with [kubernetes/kubernetes/hack/install-etcd.sh](https://github.com/kubernetes/kubernetes/blob/master/hack/install-etcd.sh).

```bash
cd simulator
make start
```

It starts etcd and simulator-server locally.

#### Run simulator frontend

To run the frontend, please see [README.md](web/README.md) on ./web dir.

## Existing cluster importing
The simulator can import existing clusters.
This allows for batch inclusion of resources from external clusters.
It is enabled by an `EXTERNAL_IMPORT_ENABLED` environment variables is `1`.

The following environment variables configure this function:
- `EXTERNAL_IMPORT_ENABLED`:If it is `1`, this importing function is enabled.
  The scheduler starts importing from external clusters as soon as it starts.
- `KUBECONFIG`:The cluster from which you are importing is indicated by this environment variable.
  [Read more](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/#the-kubeconfig-environment-variable)
- `KUBE_SCHEDULER_CONFIG_PATH`:This is an optional. The simulator can't import the scheduler configuration from existing cluster via `kube-apiserver`.
  If you set the file path to this variable, then you could import the scheduler configuration.

## Contributing

see [CONTRIBUTING.md](CONTRIBUTING.md)

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](http://slack.k8s.io/)
- [Mailing List](https://groups.google.com/forum/#!forum/kubernetes-dev)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

[owners]: https://git.k8s.io/community/contributors/guide/owners.md
[creative commons 4.0]: https://git.k8s.io/website/LICENSE
