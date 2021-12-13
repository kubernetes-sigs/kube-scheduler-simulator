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

## Reset scheduler configuration

restart scheduler with default configuration.

### HTTP Request

`PUT /api/v1/schedulerconfiguration`

### Request Body

empty

### Response

empty

| code  | description |
| ----- | -------- | 
| 202   | |
| 500 | something went wrong (see logs of the simulator server) |

## Apply Node

apply nodes.

### HTTP Request

`POST /api/v1/nodes`

### Request Body

[v1.NodeApplyConfiguration](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/client-go/applyconfigurations/core/v1/node.go#L32)

### Response

[v1.Node](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/api/core/v1/types.go#L5027)

| code  | description |
| ----- | -------- | 
| 202   | |
| 500 | something went wrong (see logs of the simulator server) |

## List Nodes

list all nodes.

### HTTP Request

`GET /api/v1/nodes`

### Response

[v1.NodeList](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/api/core/v1/types.go#L5050)

| code  | description |
| ----- | -------- | 
| 200   | |
| 500 | something went wrong (see logs of the simulator server) |

## Get Node

get a node with name.

### HTTP Request

`GET /api/v1/nodes/{name}`

### Path Parameters

| parameter | description |
| --- | ------- |
| name | node name|


### Response

[v1.Node](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/api/core/v1/types.go#L5027)

| code  | description |
| ----- | -------- | 
| 200   | |
| 404   | not found |
| 500 | something went wrong (see logs of the simulator server) |

## Delete Node

delete a node.

### HTTP Request

`DELETE /api/v1/nodes/{name}`

### Path Parameters

| parameter | description |
| --- | ------- |
| name | node name|

### Response

empty

| code  | description |
| ----- | -------- | 
| 200   | |
| 500 | something went wrong (see logs of the simulator server) |

## Apply pod

apply pods.

### HTTP Request

`POST /api/v1/pods`

### Request Body

[v1.PodApplyConfiguration](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/client-go/applyconfigurations/core/v1/pod.go#L32)

### Response

[v1.pod](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/api/core/v1/types.go#L3720)

| code  | description |
| ----- | -------- | 
| 202   | |
| 500 | something went wrong (see logs of the simulator server) |

## List pods

list all pods.

### HTTP Request

`GET /api/v1/pods`

### Response

[v1.podList](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/api/core/v1/types.go#L3744)

| code  | description |
| ----- | -------- | 
| 200   | |
| 500 | something went wrong (see logs of the simulator server) |

## Get pod

get a pod with name.

### HTTP Request

`GET /api/v1/pods/{name}`

### Path Parameters

| parameter | description |
| --- | ------- |
| name | pod name|


### Response

[v1.pod](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/api/core/v1/types.go#L3720)

| code  | description |
| ----- | -------- | 
| 200   | |
| 404   | not found |
| 500 | something went wrong (see logs of the simulator server) |

## Delete pod

delete a pod.

### HTTP Request

`DELETE /api/v1/pods/{name}`

### Path Parameters

| parameter | description |
| --- | ------- |
| name | pod name|

### Response

empty

| code  | description |
| ----- | -------- | 
| 200   | |
| 500 | something went wrong (see logs of the simulator server) |

## Apply persistent volume

apply persistent volumes.

### HTTP Request

`POST /api/v1/persistentvolumes`

### Request Body

[v1.PersistentVolumeApplyConfiguration](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/client-go/applyconfigurations/core/v1/persistentvolume.go#L32)

### Response

[v1.PersistentVolume](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/api/core/v1/types.go#L305)

| code  | description |
| ----- | -------- | 
| 202   | |
| 500 | something went wrong (see logs of the simulator server) |

## List persistent volumes

list all persistent volumes.

### HTTP Request

`GET /api/v1/persistentvolumes`

### Response

[v1.PersistentVolumeList](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/api/core/v1/types.go#L419)

| code  | description |
| ----- | -------- | 
| 200   | |
| 500 | something went wrong (see logs of the simulator server) |

## Get persistent volume

get a persistent volume with name.

### HTTP Request

`GET /api/v1/persistentvolumes/{name}`

### Path Parameters

| parameter | description |
| --- | ------- |
| name | persistent volume name|


### Response

[v1.PersistentVolume](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/api/core/v1/types.go#L305)

| code  | description |
| ----- | -------- | 
| 200   | |
| 404   | not found |
| 500 | something went wrong (see logs of the simulator server) |

## Delete persistent volume

delete a persistent volume.

### HTTP Request

`DELETE /api/v1/persistentvolumes/{name}`

### Path Parameters

| parameter | description |
| --- | ------- |
| name | persistent volume name|

### Response

empty

| code  | description |
| ----- | -------- | 
| 200   | |
| 500 | something went wrong (see logs of the simulator server) |


## Apply persistent volume claim

apply persistent volume claims.

### HTTP Request

`POST /api/v1/persistentvolumeclaims`

### Request Body

[v1.PersistentVolumeClaimApplyConfiguration](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/client-go/applyconfigurations/core/v1/persistentvolumeclaim.go#L32)

### Response

[v1.PersistentVolumeClaim](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/api/core/v1/types.go#L434)

| code  | description |
| ----- | -------- | 
| 202   | |
| 500 | something went wrong (see logs of the simulator server) |

## List persistent volume claims

list all persistent volume claims.

### HTTP Request

`GET /api/v1/persistentvolumeclaims`

### Response

[v1.PersistentVolumeClaimList](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/api/core/v1/types.go#L456)

| code  | description |
| ----- | -------- | 
| 200   | |
| 500 | something went wrong (see logs of the simulator server) |

## Get persistent volume claim

get a persistent volume claim with name.

### HTTP Request

`GET /api/v1/persistentvolumeclaims/{name}`

### Path Parameters

| parameter | description |
| --- | ------- |
| name | persistent volume claim name|


### Response

[v1.PersistentVolumeClaim](https://github.com/kubernetes/kubernetes/blob/release-1.22/staging/src/k8s.io/api/core/v1/types.go#L434)

| code  | description |
| ----- | -------- | 
| 200   | |
| 404   | not found |
| 500 | something went wrong (see logs of the simulator server) |

## Delete persistent volume claim

delete a persistent volume claim.

### HTTP Request

`DELETE /api/v1/persistentvolumeclaims/{name}`

### Path Parameters

| parameter | description |
| --- | ------- |
| name | persistent volume claim name|

### Response

empty

| code  | description |
| ----- | -------- | 
| 200   | |
| 500 | something went wrong (see logs of the simulator server) |
