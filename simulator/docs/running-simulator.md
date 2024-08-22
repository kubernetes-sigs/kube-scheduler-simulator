## Running simulator

The simulator requires docker installed in your laptop.
We have [docker-compose.yml](../../docker-compose.yml) to run the simulator easily.

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

Additionally, you can run a kwok cluster that acts as a fake source cluster to try out [the resource importing feature](./import-cluster-resources.md).

```
make docker_build_and_up -e COMPOSE_PROFILES=externalImportEnabled
```
