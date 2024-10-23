### Environment Variables
These describe the environment variables that are used to configure the simulator's frontend.

Please refer [compose.yml](./../compose.yml) as an example use.

`KUBE_API_SERVER_URL`: This is the kube-apiserver URL. Its default
value is `http://localhost:3131`.

`BASE_URL`: This is the URL for the kube-scheduler-simulator
server. Its default value is `http://localhost:1212`.

`ALPHA_TABLE_VIEWS`: This variable enables the alpha feature `table
view`. Its value is either 0(default) or 1 (0 means disabled, 1
meaning enabled). We can see the resource status in the table.
