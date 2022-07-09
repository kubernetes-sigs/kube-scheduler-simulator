# API reference

This page describe the simulator's HTTP API endpoint.

## Get scheduler configuration

get current scheduler configuration.

### HTTP Request

`GET /api/v1/schedulerconfiguration`

### Response

[v1beta2.KubeSchedulerConfiguration](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/kube-scheduler/config/v1beta2/types.go#L44)

| code  | description |
| ----- | -------- |
| 200   | |


## Update scheduler configuration

update scheduler configuration and restart scheduler with new configuration.

### HTTP Request

`POST /api/v1/schedulerconfiguration`

### Request Body

[v1beta2.KubeSchedulerConfiguration](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/kube-scheduler/config/v1beta2/types.go#L44)

### Response

empty

| code  | description |
| ----- | -------- |
| 202   | |
| 500 | something went wrong (see logs of the simulator server) |

## Reset all resources and scheduler configutarion

clean up all resources and restore the initial scheduler configuration.
(If you didn't pass the initial scheduler configuration via `KUBE_SCHEDULER_CONFIG_PATH`, the default scheduler configuration will be restored.)

### HTTP Request

`PUT /api/v1/reset`

### Request Body

empty

### Response

empty

| code  | description |
| ----- | -------- |
| 202   | |
| 500 | something went wrong (see logs of the simulator server) |

## Export

Get all resources and current scheduler configuration.

### HTTP Request

`GET /api/v1/export`


### Response

[ResourcesForImport](/simulator/server/handler/export.go#L21)

You can find sample requests/responses [here](api-samples/v1/export.md)

| code  | description |
| ----- | -------- |
| 200   | |
| 500 | something went wrong (see logs of the simulator server) |

## Import

Apply resources and scheduler configuration.

### HTTP Request

`POST /api/v1/import`

### Request Body

[ResourcesForImport](/simulator/server/handler/export.go#L21)

You can find sample requests/responses [here](api-samples/v1/import.md)
### Response

| code  | description |
| ----- | -------- |
| 200   | |
| 500 | something went wrong (see logs of the simulator server) |
