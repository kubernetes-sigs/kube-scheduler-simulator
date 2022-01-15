# /v1/import samples

## case1: import resources with some PVs and PVCs

### Request
```
POST /api/v1/import HTTP/1.1
Host: localhost:1212
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:93.0) Gecko/20100101 Firefox/93.0
Accept: application/json, text/plain, */*
Accept-Language: ja,en-US;q=0.7,en;q=0.3
Accept-Encoding: gzip, deflate
Origin: http://localhost:3000
Connection: close
Referer: http://localhost:3000/
Sec-Fetch-Dest: empty
Sec-Fetch-Mode: cors
Sec-Fetch-Site: cross-site
Content-Type: application/json; charset=UTF-8
Content-Length: 4580

{"pods":[],"nodes":[],"pvs":[{"metadata":{"name":"pv1","uid":"db4a5204-ef32-4ff4-b112-be4090b3e57e","resourceVersion":"1488","creationTimestamp":"2021-12-28T03:59:58Z","annotations":{"pv.kubernetes.io/bound-by-controller":"yes"},"managedFields":[{"manager":"simulator","operation":"Apply","apiVersion":"v1","time":"2021-12-28T03:59:58Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:accessModes":{},"f:capacity":{"f:storage":{}},"f:hostPath":{"f:path":{},"f:type":{}},"f:persistentVolumeReclaimPolicy":{},"f:volumeMode":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T03:59:58Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:pv.kubernetes.io/bound-by-controller":{}}},"f:spec":{"f:claimRef":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T03:59:58Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:phase":{}}},"subresource":"status"}]},"spec":{"capacity":{"storage":"1Gi"},"hostPath":{"path":"/tmp/data","type":"DirectoryOrCreate"},"accessModes":["ReadWriteOnce"],"claimRef":{"kind":"PersistentVolumeClaim","namespace":"default","name":"pvc1","uid":"eaeda676-2c00-425f-b4c2-ad090a2e3c5a","apiVersion":"v1","resourceVersion":"1481"},"persistentVolumeReclaimPolicy":"Delete","volumeMode":"Filesystem"},"status":{"phase":"Bound"}}],"pvcs":[{"metadata":{"name":"pvc1","namespace":"default","uid":"eaeda676-2c00-425f-b4c2-ad090a2e3c5a","resourceVersion":"1490","creationTimestamp":"2021-12-28T03:59:56Z","annotations":{"pv.kubernetes.io/bind-completed":"yes","pv.kubernetes.io/bound-by-controller":"yes"},"managedFields":[{"manager":"simulator","operation":"Apply","apiVersion":"v1","time":"2021-12-28T03:59:56Z","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{"f:accessModes":{},"f:resources":{"f:requests":{"f:storage":{}}},"f:volumeMode":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T03:59:58Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{".":{},"f:pv.kubernetes.io/bind-completed":{},"f:pv.kubernetes.io/bound-by-controller":{}}},"f:spec":{"f:volumeName":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2021-12-28T03:59:58Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:accessModes":{},"f:capacity":{".":{},"f:storage":{}},"f:phase":{}}},"subresource":"status"}]},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"volumeName":"pv1","volumeMode":"Filesystem"},"status":{"phase":"Bound","accessModes":["ReadWriteOnce"],"capacity":{"storage":"1Gi"}}}],"storageClasses":[],"priorityClasses":[],"schedulerConfig":{"parallelism":16,"leaderElection":{"leaderElect":true,"leaseDuration":"15s","renewDeadline":"10s","retryPeriod":"2s","resourceLock":"leases","resourceName":"kube-scheduler","resourceNamespace":"kube-system"},"clientConnection":{"kubeconfig":"","acceptContentTypes":"","contentType":"application/vnd.kubernetes.protobuf","qps":50,"burst":100},"healthzBindAddress":"0.0.0.0:10251","metricsBindAddress":"0.0.0.0:10251","enableProfiling":true,"enableContentionProfiling":true,"percentageOfNodesToScore":0,"podInitialBackoffSeconds":1,"podMaxBackoffSeconds":10,"profiles":[{"schedulerName":"default-scheduler","plugins":{"queueSort":{"enabled":[{"name":"PrioritySort"}]},"preFilter":{"enabled":[{"name":"NodeResourcesFit"},{"name":"NodePorts"},{"name":"VolumeRestrictions"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"},{"name":"VolumeBinding"},{"name":"NodeAffinity"}]},"filter":{"enabled":[{"name":"NodeUnschedulable"},{"name":"NodeName"},{"name":"TaintToleration"},{"name":"NodeAffinity"},{"name":"NodePorts"},{"name":"NodeResourcesFit"},{"name":"VolumeRestrictions"},{"name":"EBSLimits"},{"name":"GCEPDLimits"},{"name":"NodeVolumeLimits"},{"name":"AzureDiskLimits"},{"name":"VolumeBinding"},{"name":"VolumeZone"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"}]},"postFilter":{"enabled":[{"name":"DefaultPreemption"}]},"preScore":{"enabled":[{"name":"InterPodAffinity"},{"name":"PodTopologySpread"},{"name":"TaintToleration"},{"name":"NodeAffinity"}]},"score":{"enabled":[{"name":"NodeResourcesBalancedAllocation","weight":1},{"name":"ImageLocality","weight":1},{"name":"InterPodAffinity","weight":1},{"name":"NodeResourcesFit","weight":1},{"name":"NodeAffinity","weight":1},{"name":"PodTopologySpread","weight":2},{"name":"TaintToleration","weight":1}]},"reserve":{"enabled":[{"name":"VolumeBinding"}]},"permit":{},"preBind":{"enabled":[{"name":"VolumeBinding"}]},"bind":{"enabled":[{"name":"DefaultBinder"}]},"postBind":{}}}]}}
```

### Response
```
HTTP/1.1 200 OK
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: http://localhost:3000
Vary: Origin
Date: Sun, 02 Jan 2022 15:10:37 GMT
Content-Length: 0
Connection: close

```