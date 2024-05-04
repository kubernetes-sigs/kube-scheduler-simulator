## Running simulator

There are multiple ways to run the simulator with web UI.


### Run simulator with Docker

We have [docker-compose.yml](../../docker-compose.yml) to run the simulator easily.
You should install [docker](https://docs.docker.com/engine/install/) and [docker-compose](https://docs.docker.com/compose/install/) at first.

You can use either of following commands.

```bash
# pull docker images from the registry and run them. 
# It's the easiest way to run up the simulator and web UI.
make docker_up

# build the images for web frontend and simulator server, then start the containers.
# You need to use this if you change the implementation of the simulator.
make docker_build_and_up
```

Then, you can access the simulator with http://localhost:3000.
If you want to deploy the simulator on a remote server and access it via a specific IP (e.g: like http://10.0.0.1:3000/),
please make sure that you have executed `export SIMULATOR_EXTERNAL_IP=your.server.ip` before running `docker compose up -d`.

### Run simulator without docker

You have to run web UI, server and etcd.

#### 1. Run simulator server and etcd

To run this simulator's server, you have to install Go and etcd.

You can install etcd with [kubernetes/kubernetes/hack/install-etcd.sh](https://github.com/kubernetes/kubernetes/blob/master/hack/install-etcd.sh).

```bash
cd simulator
# build and start the simulator + run up etcd.
make start
```

#### 2. Run web UI

To run the frontend, please see [README.md](../../web/README.md) on ./web dir.
