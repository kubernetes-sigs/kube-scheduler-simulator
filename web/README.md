# scheduler-simulator

This is the frontend of Kubernetes scheduler simulator.

## Run frontend

You have to install node.js and yarn.

- for yarn, see: [Installation | Yarn](https://classic.yarnpkg.com/en/docs/install/#mac-stable)
- for node.js, see: [Downloads | Node.js](https://nodejs.org/en/download/)

### Build Setup

```bash
# install dependencies
$ yarn install

# build for production and launch server
$ yarn build
$ yarn start
```

For detailed explanation on how things work, check out [Nuxt.js docs](https://nuxtjs.org).

## for developer

```bash
# serve with hot reload at localhost:3000
$ yarn dev
# format the code
$ yarn format
# lint the code
$ yarn lint
```

## Importing Certificate

You should import apiserver's certificate on your browser,
since the simulator frontend uses `HTTP/2` connections for some communications.

If you won't import it, the resources fetch will be failed.

(Note: However, as a backdoor trick, you can avoid the problem by accessing `https://127.0.0.1:3131` directly from your browser and giving yourself permission to ignore the error at your own risk.)

### Default certificate

If you won't prepare your own certificate and won't specify the certificate file path, the simulator's apiserver use the test certificate of (`net/http/internal/testcert`)[https://pkg.go.dev/net/http/internal/testcert#pkg-variables].

Therefore, all you have to do is these 2 steps.

1. Create a local file and copy&paste the content of (`LocalhostCert`)[https://pkg.go.dev/net/http/internal/testcert#pkg-variables].
2. On browser settings or machine settings, import the certificate file of `1` as a trusted certificate. ※See also (this page)[https://docs.vmware.com/en/VMware-Adapter-for-SAP-Landscape-Management/index.html] as one example.

### Your own certificate

You can also use a certificate created on your own.
To achieve that, you will use two env variables `KUBE_API_CERT_PATH` and `KUBE_API_KEY_PATH`.

Please follow these steps.

1. Create your own private key and certificate files.
2. Set `KUBE_API_CERT_PATH` and `KUBE_API_KEY_PATH` env variables to each file path.
3. On browser settings or machine settings, import the certificate file of `2` as a trusted certificate. ※See also (this page)[https://docs.vmware.com/en/VMware-Adapter-for-SAP-Landscape-Management/index.html] as one example.
