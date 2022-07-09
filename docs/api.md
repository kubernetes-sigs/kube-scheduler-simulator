# API reference

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
Once this API is called, the server will continuously returns a WatchEvent response containing event information for a resource.

### HTTP Request

`GET /api/v1/watchresources`

#### Parameter
You must specify the `lastResourceVersion` of each resource, which can be retrieved using the `list` function of each resource.
If you won't specify it, the simulator failed to watch the resource.

|parameter|requirement|description|
| ----- | --- | -------- |
|podsLastResourceVersion|MUST|string|
|nodesLastResourceVersion|MUST|string|
|pvsLastResourceVersion|MUST|string|
|pvcsLastResourceVersion|MUST|string|
|scsLastResourceVersion|MUST|string|
|pcsLastResourceVersion|MUST|string|

These `resourceVersion` set a constraint on what resource versions a request may be served from. See https://kubernetes.io/docs/reference/using-api/api-concepts/#resource-versions for details. 

e.g.)
```
/api/v1/watchresources?podsLastResourceVersion=213&nodesLastResourceVersion=213&pvsLastResourceVersion=213&pvcsLastResourceVersion=213&scsLastResourceVersion=213&pcsLastResourceVersion=213
```

### Response

[WatchEvent](watcher/watcher.go#43)

| code  | description |
| ----- | -------- |
| 200   | The response is server push. You should catch the WatchEvent and then handle the data each by each.|

