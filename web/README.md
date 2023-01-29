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
Please see [environment-variables.md](./docs/environment-variables.md)

## For developer

```bash
# serve with hot reload at localhost:3000
$ yarn dev
# format the code
$ yarn format
# lint the code
$ yarn lint
```
