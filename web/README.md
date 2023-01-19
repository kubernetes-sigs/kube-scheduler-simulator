# Scheduler-Simulator

This is the frontend of Kubernetes scheduler simulator.

## Run frontend

You have to install node.js and yarn.

- for yarn, see: [Installation | Yarn](https://classic.yarnpkg.com/en/docs/install/#mac-stable)
- for node.js, see: [Downloads | Node.js](https://nodejs.org/en/download/)
  Note: Nodejs 16 is suggested, other version may cause problems.

### Build Setup

```bash
# install dependencies
$ yarn install

# build for production and launch server
$ yarn build
$ yarn start
```

For detailed explanation on how things work, check out [Nuxt.js docs](https://nuxtjs.org).

### Environment Variables
These describe the environment variables that are used to configure the simulator's frontend.

Please refer [docker-compose.yml](./../docker-compose.yml) as an example use.

`KUBE_API_SERVER_URL`: This is the kube-apiserver URL. Its default
value is `http://localhost:3131`.

`BASE_URL`: This is the URL for the kube-scheduler-simulator
server. Its default value is `http://localhost:1212`.

`ALPHA_TABLE_VIEWS`: This variable enables the alpha feature `table
view`. Its value is either 0(default) or 1 (0 means disabled, 1
meaning enabled). We can see the resource status in the table.

## For developer

```bash
# serve with hot reload at localhost:3000
$ yarn dev
# format the code
$ yarn format
# lint the code
$ yarn lint
```
