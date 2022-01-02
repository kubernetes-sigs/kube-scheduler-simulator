# Export immediately after startup (empty)
## Request
```
GET /api/v1/export HTTP/1.1
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
```

## Response
```
HTTP/1.1 200 OK
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: http://localhost:3000
Content-Type: application/json; charset=UTF-8
Vary: Origin
Date: Sun, 02 Jan 2022 15:06:58 GMT
Connection: close
Content-Length: 3216

{"pods":[],"nodes":[],"pvs":[],"pvcs":[],"storageClasses":[],"schedulerConfig":{"parallelism":16,"leaderElection":{"leaderElect":true,"leaseDuration":"15s","renewDeadline":"10s","retryPeriod":"2s","resourceLock":"leases","resourceName":"kube-scheduler","resourceNamespace":"kube-system"},"clientConnection":{"kubeconfig":"","acceptContentTypes":"","contentType":"application/vnd.kubernetes.protobuf","qps":50,"burst":100},"healthzBindAddress":"0.0.0.0:10251","metricsBindAddress":"0.0.0.0:10251","enableProfiling":true,"enableContentionProfiling":true,"percentageOfNodesToScore":0,"podInitialBackoffSeconds":1,"podMaxBackoffSeconds":10,"profiles":[{"schedulerName":"default-scheduler","plugins":{"queueSort":{"enabled":[{"name":"PrioritySort"}]},"preFilter":{"enabled":[{"name":"NodeResourcesFit"},{"name":"NodePorts"},{"name":"VolumeRestrictions"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"},{"name":"VolumeBinding"},{"name":"NodeAffinity"}]},"filter":{"enabled":[{"name":"NodeUnschedulable"},{"name":"NodeName"},{"name":"TaintToleration"},{"name":"NodeAffinity"},{"name":"NodePorts"},{"name":"NodeResourcesFit"},{"name":"VolumeRestrictions"},{"name":"EBSLimits"},{"name":"GCEPDLimits"},{"name":"NodeVolumeLimits"},{"name":"AzureDiskLimits"},{"name":"VolumeBinding"},{"name":"VolumeZone"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"}]},"postFilter":{"enabled":[{"name":"DefaultPreemption"}]},"preScore":{"enabled":[{"name":"InterPodAffinity"},{"name":"PodTopologySpread"},{"name":"TaintToleration"},{"name":"NodeAffinity"}]},"score":{"enabled":[{"name":"NodeResourcesBalancedAllocation","weight":1},{"name":"ImageLocality","weight":1},{"name":"InterPodAffinity","weight":1},{"name":"NodeResourcesFit","weight":1},{"name":"NodeAffinity","weight":1},{"name":"PodTopologySpread","weight":2},{"name":"TaintToleration","weight":1}]},"reserve":{"enabled":[{"name":"VolumeBinding"}]},"permit":{},"preBind":{"enabled":[{"name":"VolumeBinding"}]},"bind":{"enabled":[{"name":"DefaultBinder"}]},"postBind":{}},"pluginConfig":[{"name":"DefaultPreemption","args":{"kind":"DefaultPreemptionArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","minCandidateNodesPercentage":10,"minCandidateNodesAbsolute":100}},{"name":"InterPodAffinity","args":{"kind":"InterPodAffinityArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","hardPodAffinityWeight":1}},{"name":"NodeAffinity","args":{"kind":"NodeAffinityArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2"}},{"name":"NodeResourcesBalancedAllocation","args":{"kind":"NodeResourcesBalancedAllocationArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","resources":[{"name":"cpu","weight":1},{"name":"memory","weight":1}]}},{"name":"NodeResourcesFit","args":{"kind":"NodeResourcesFitArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","scoringStrategy":{"type":"LeastAllocated","resources":[{"name":"cpu","weight":1},{"name":"memory","weight":1}]}}},{"name":"PodTopologySpread","args":{"kind":"PodTopologySpreadArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","defaultingType":"System"}},{"name":"VolumeBinding","args":{"kind":"VolumeBindingArgs","apiVersion":"kubescheduler.config.k8s.io/v1beta2","bindTimeoutSeconds":600}}]}]}}
```

# Export resources containing PersistentVolume and PersistentVolumeClaim
## Request
```
GET /api/v1/export HTTP/1.1
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
```

## Response
```
HTTP/1.1 200 OK
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: http://localhost:3000
Content-Type: application/json; charset=UTF-8
Vary: Origin
Date: Sun, 02 Jan 2022 15:18:25 GMT
Connection: close
Content-Length: 4126

{"pods":[],"nodes":[],"pvs":[{"metadata":{"name":"pv1","uid":"ae9d1168-ccd9-47f1-ac6f-3f6295b29753","resourceVersion":"2034","creationTimestamp":"2022-01-02T15:10:37Z","annotations":{"pv.kubernetes.io/bound-by-controller":"yes"},"managedFields":[{"manager":"simulator","operation":"Apply","apiVersion":"v1","time":"2022-01-02T15:10:37Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:pv.kubernetes.io/bound-by-controller":{}}},"f:spec":{"f:accessModes":{},"f:capacity":{"f:storage":{}},"f:claimRef":{},"f:hostPath":{"f:path":{},"f:type":{}},"f:persistentVolumeReclaimPolicy":{},"f:volumeMode":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2022-01-02T15:10:37Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:phase":{}}},"subresource":"status"}]},"spec":{"capacity":{"storage":"1Gi"},"hostPath":{"path":"/tmp/data","type":"DirectoryOrCreate"},"accessModes":["ReadWriteOnce"],"claimRef":{"kind":"PersistentVolumeClaim","namespace":"default","name":"pvc1","apiVersion":"v1","resourceVersion":"1481"},"persistentVolumeReclaimPolicy":"Delete","volumeMode":"Filesystem"},"status":{"phase":"Available"}}],"pvcs":[{"metadata":{"name":"pvc1","namespace":"default","uid":"e4f5d5f4-2372-4036-ad10-1cb063970656","resourceVersion":"2036","creationTimestamp":"2022-01-02T15:10:37Z","annotations":{"pv.kubernetes.io/bind-completed":"yes","pv.kubernetes.io/bound-by-controller":"yes"},"managedFields":[{"manager":"simulator","operation":"Apply","apiVersion":"v1","time":"2022-01-02T15:10:37Z","fieldsType":"FieldsV1","fieldsV1":{"f:metadata":{"f:annotations":{"f:pv.kubernetes.io/bind-completed":{},"f:pv.kubernetes.io/bound-by-controller":{}}},"f:spec":{"f:accessModes":{},"f:resources":{"f:requests":{"f:storage":{}}},"f:volumeMode":{},"f:volumeName":{}}}},{"manager":"simulator","operation":"Update","apiVersion":"v1","time":"2022-01-02T15:10:37Z","fieldsType":"FieldsV1","fieldsV1":{"f:status":{"f:phase":{}}},"subresource":"status"}]},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"volumeName":"pv1","volumeMode":"Filesystem"},"status":{"phase":"Lost"}}],"storageClasses":[],"schedulerConfig":{"parallelism":16,"leaderElection":{"leaderElect":true,"leaseDuration":"15s","renewDeadline":"10s","retryPeriod":"2s","resourceLock":"leases","resourceName":"kube-scheduler","resourceNamespace":"kube-system"},"clientConnection":{"kubeconfig":"","acceptContentTypes":"","contentType":"application/vnd.kubernetes.protobuf","qps":50,"burst":100},"healthzBindAddress":"0.0.0.0:10251","metricsBindAddress":"0.0.0.0:10251","enableProfiling":true,"enableContentionProfiling":true,"percentageOfNodesToScore":0,"podInitialBackoffSeconds":1,"podMaxBackoffSeconds":10,"profiles":[{"schedulerName":"default-scheduler","plugins":{"queueSort":{"enabled":[{"name":"PrioritySort"}]},"preFilter":{"enabled":[{"name":"NodeResourcesFit"},{"name":"NodePorts"},{"name":"VolumeRestrictions"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"},{"name":"VolumeBinding"},{"name":"NodeAffinity"}]},"filter":{"enabled":[{"name":"NodeUnschedulable"},{"name":"NodeName"},{"name":"TaintToleration"},{"name":"NodeAffinity"},{"name":"NodePorts"},{"name":"NodeResourcesFit"},{"name":"VolumeRestrictions"},{"name":"EBSLimits"},{"name":"GCEPDLimits"},{"name":"NodeVolumeLimits"},{"name":"AzureDiskLimits"},{"name":"VolumeBinding"},{"name":"VolumeZone"},{"name":"PodTopologySpread"},{"name":"InterPodAffinity"}]},"postFilter":{"enabled":[{"name":"DefaultPreemption"}]},"preScore":{"enabled":[{"name":"InterPodAffinity"},{"name":"PodTopologySpread"},{"name":"TaintToleration"},{"name":"NodeAffinity"}]},"score":{"enabled":[{"name":"NodeResourcesBalancedAllocation","weight":1},{"name":"ImageLocality","weight":1},{"name":"InterPodAffinity","weight":1},{"name":"NodeResourcesFit","weight":1},{"name":"NodeAffinity","weight":1},{"name":"PodTopologySpread","weight":2},{"name":"TaintToleration","weight":1}]},"reserve":{"enabled":[{"name":"VolumeBinding"}]},"permit":{},"preBind":{"enabled":[{"name":"VolumeBinding"}]},"bind":{"enabled":[{"name":"DefaultBinder"}]},"postBind":{}}}]}}

```