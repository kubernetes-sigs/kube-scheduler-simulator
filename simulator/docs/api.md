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

## Watch the simulator's resources

Watch individual changes to all k8s resources in the simulator. This endpoint uses `Server-Sent Events`.
Once this API is called, the server will be continuously sending WatchEvent every time the event happens.

### HTTP Request

`GET /api/v1/listwatchresources`

#### Parameter
You can specify the `lastResourceVersion` of each resource, which can be retrieved using the `list` API of each resource.
If you won't specify it, this API calls the `list` and returns the result as "ADDED" Events before starting watch.  

We recommend to call the `list` API before the calling and to use `XXXlastResourceVersion` parameters.

The `ResourceVersion` must be treated as opaque by clients and passed unmodified back to the server.
See also [this page](https://kubernetes.io/docs/reference/using-api/api-concepts/#resource-versions).


| parameter                | requirement | description                                                                                   |
|--------------------------|-------------|-----------------------------------------------------------------------------------------------|
| podslastResourceVersion  | OPTIONAL    | If not specified, all resources are returned as `ADDED` Events first and then start to watch. |
| nodeslastResourceVersion | OPTIONAL    | If not specified, all resources are returned as `ADDED` Events first and then start to watch. |
| pvslastResourceVersion   | OPTIONAL    | If not specified, all resources are returned as `ADDED` Events first and then start to watch. |
| pvcslastResourceVersion  | OPTIONAL    | If not specified, all resources are returned as `ADDED` Events first and then start to watch. |
| scslastResourceVersion   | OPTIONAL    | If not specified, all resources are returned as `ADDED` Events first and then start to watch. |
| pcslastResourceVersion   | OPTIONAL    | If not specified, all resources are returned as `ADDED` Events first and then start to watch. |

e.g.)
```
/api/v1/listwatchresources?podslastResourceVersion=213&nodeslastResourceVersion=213&pvslastResourceVersion=213&pvcslastResourceVersion=213&scslastResourceVersion=213&pcslastResourceVersion=213
```

### Response

[WatchEvent](/simulator/resourcewatcher/streamwriter/streamwriter.go#L18)

| code  | description |
| ----- | -------- |
| 200   | The response is server push. You should catch the WatchEvent and then handle the data each by each.|

