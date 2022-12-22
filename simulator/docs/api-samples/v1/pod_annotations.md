# Pod Annotations

The simulator server provides access to unmarshalled scheduler-simulator annotations with simple GET requests producing JSON output.

In addition, all pod information can be accessed as provided by apimachinery. The simulator server provides endpoints for a specific pod, all pods in a namespace, or all pods in all namespaces.

## Get a specific scheduler-simulator annotation for a specific pod

* For example: **"scheduler-simulator/filter-result"**
* **Annotations are unmarshalled**
* Data structure comes from apimachinery

### curl command
```bash
curl http://localhost:1212/api/v1/namespaces/default/pods/pod-w64pz/metadata/annotations/scheduler-simulator/filter-result
```

### Output

```json
{
 "node-pprb7": {
  "AzureDiskLimits": "passed",
  "EBSLimits": "passed",
  "GCEPDLimits": "passed",
  "InterPodAffinity": "passed",
  "NodeAffinity": "passed",
  "NodeName": "passed",
  "NodePorts": "passed",
  "NodeResourcesFit": "passed",
  "NodeUnschedulable": "passed",
  "NodeVolumeLimits": "passed",
  "PodTopologySpread": "passed",
  "TaintToleration": "passed",
  "VolumeBinding": "passed",
  "VolumeRestrictions": "passed",
  "VolumeZone": "passed"
 }
}
```

## Get all scheduler-simulator annotations for a specific pod

* **Annotations are unmarshalled**
* Data structure comes from apimachinery
* This only includes annotations that have a "scheduler-simulator/" prefix.

### curl command
```bash
curl http://localhost:1212/api/v1/namespaces/default/pods/pod-w64pz/metadata/annotations/scheduler-simulator
```

### Output

```json
{
 "scheduler-simulator/filter-result": {
  "node-pprb7": {
   "AzureDiskLimits": "passed",
   "EBSLimits": "passed",
   "GCEPDLimits": "passed",
   "InterPodAffinity": "passed",
   "NodeAffinity": "passed",
   "NodeName": "passed",
   "NodePorts": "passed",
   "NodeResourcesFit": "passed",
   "NodeUnschedulable": "passed",
   "NodeVolumeLimits": "passed",
   "PodTopologySpread": "passed",
   "TaintToleration": "passed",
   "VolumeBinding": "passed",
   "VolumeRestrictions": "passed",
   "VolumeZone": "passed"
  }
 },
 "scheduler-simulator/finalscore-result": {},
 "scheduler-simulator/postFilter-result": {},
 "scheduler-simulator/score-result": {}
}
```

## Get all annotations for a specific pod

* **Annotations are not filtered or unmarshalled**
* Data structure comes from apimachinery
* This would include annotations that are not scoped by "scheduler-simulator" (if any were added).

### curl command

For example, pod **"pod-w64pz"** in **"default"** namespace:

```bash
curl http://localhost:1212/api/v1/namespaces/default/pods/pod-w64pz/metadata/annotations
```

### Output

```json
{
 "scheduler-simulator/filter-result": "{\"node-pprb7\":{\"AzureDiskLimits\":\"passed\",\"EBSLimits\":\"passed\",\"GCEPDLimits\":\"passed\",\"InterPodAffinity\":\"passed\",\"NodeAffinity\":\"passed\",\"NodeName\":\"passed\",\"NodePorts\":\"passed\",\"NodeResourcesFit\":\"passed\",\"NodeUnschedulable\":\"passed\",\"NodeVolumeLimits\":\"passed\",\"PodTopologySpread\":\"passed\",\"TaintToleration\":\"passed\",\"VolumeBinding\":\"passed\",\"VolumeRestrictions\":\"passed\",\"VolumeZone\":\"passed\"}}",
 "scheduler-simulator/finalscore-result": "{}",
 "scheduler-simulator/postFilter-result": "{}",
 "scheduler-simulator/score-result": "{}"
}
```

## Get a specific pod

* Annotations are not filtered or unmarshalled
* Data structure comes from apimachinery

### curl command

For example, pod **"pod-w64pz"** in **"default"** namespace:

```bash
curl http://localhost:1212/api/v1/namespaces/default/pods/pod-w64pz
```

### Output

```json
{
 "metadata": {
  "name": "pod-w64pz",
  "generateName": "pod-",
  "namespace": "default",
  "uid": "c5a5b6bf-2de8-4cf6-97d1-8c76ab075eb4",
  "resourceVersion": "244",
  "creationTimestamp": "2022-12-22T05:38:40Z",
  "annotations": {
   "scheduler-simulator/filter-result": "{\"node-pprb7\":{\"AzureDiskLimits\":\"passed\",\"EBSLimits\":\"passed\",\"GCEPDLimits\":\"passed\",\"InterPodAffinity\":\"passed\",\"NodeAffinity\":\"passed\",\"NodeName\":\"passed\",\"NodePorts\":\"passed\",\"NodeResourcesFit\":\"passed\",\"NodeUnschedulable\":\"passed\",\"NodeVolumeLimits\":\"passed\",\"PodTopologySpread\":\"passed\",\"TaintToleration\":\"passed\",\"VolumeBinding\":\"passed\",\"VolumeRestrictions\":\"passed\",\"VolumeZone\":\"passed\"}}",
   "scheduler-simulator/finalscore-result": "{}",
   "scheduler-simulator/postFilter-result": "{}",
   "scheduler-simulator/score-result": "{}"
  },
  "managedFields": [
   {
    "manager": "simulator",
    "operation": "Update",
    "apiVersion": "v1",
    "time": "2022-12-22T05:38:40Z",
    "fieldsType": "FieldsV1",
    "fieldsV1": {
     "f:status": {
      "f:conditions": {
       ".": {},
       "k:{\"type\":\"PodScheduled\"}": {
        ".": {},
        "f:lastProbeTime": {},
        "f:lastTransitionTime": {},
        "f:message": {},
        "f:reason": {},
        "f:status": {},
        "f:type": {}
       }
      }
     }
    },
    "subresource": "status"
   },
   {
    "manager": "simulator",
    "operation": "Update",
    "apiVersion": "v1",
    "time": "2022-12-22T05:38:43Z",
    "fieldsType": "FieldsV1",
    "fieldsV1": {
     "f:metadata": {
      "f:annotations": {
       ".": {},
       "f:scheduler-simulator/filter-result": {},
       "f:scheduler-simulator/finalscore-result": {},
       "f:scheduler-simulator/postFilter-result": {},
       "f:scheduler-simulator/score-result": {}
      },
      "f:generateName": {}
     },
     "f:spec": {
      "f:containers": {
       "k:{\"name\":\"pause\"}": {
        ".": {},
        "f:image": {},
        "f:imagePullPolicy": {},
        "f:name": {},
        "f:resources": {
         ".": {},
         "f:limits": {
          ".": {},
          "f:cpu": {},
          "f:memory": {}
         },
         "f:requests": {
          ".": {},
          "f:cpu": {},
          "f:memory": {}
         }
        },
        "f:terminationMessagePath": {},
        "f:terminationMessagePolicy": {}
       }
      },
      "f:dnsPolicy": {},
      "f:enableServiceLinks": {},
      "f:restartPolicy": {},
      "f:schedulerName": {},
      "f:securityContext": {},
      "f:terminationGracePeriodSeconds": {}
     }
    }
   }
  ]
 },
 "spec": {
  "containers": [
   {
    "name": "pause",
    "image": "k8s.gcr.io/pause:3.5",
    "resources": {
     "limits": {
      "cpu": "100m",
      "memory": "16Gi"
     },
     "requests": {
      "cpu": "100m",
      "memory": "16Gi"
     }
    },
    "terminationMessagePath": "/dev/termination-log",
    "terminationMessagePolicy": "File",
    "imagePullPolicy": "IfNotPresent"
   }
  ],
  "restartPolicy": "Always",
  "terminationGracePeriodSeconds": 30,
  "dnsPolicy": "ClusterFirst",
  "nodeName": "node-pprb7",
  "securityContext": {},
  "schedulerName": "default-scheduler",
  "priority": 0,
  "enableServiceLinks": true,
  "preemptionPolicy": "PreemptLowerPriority"
 },
 "status": {
  "phase": "Pending",
  "conditions": [
   {
    "type": "PodScheduled",
    "status": "True",
    "lastProbeTime": null,
    "lastTransitionTime": "2022-12-22T05:38:43Z"
   }
  ],
  "qosClass": "Guaranteed"
 }
}
```

## Get all pods for a specific namespace

* Annotations are not filtered or unmarshalled
* Data structure comes from apimachinery

### curl command

For example, **"default"** namespace:

```bash
curl http://localhost:1212/api/v1/namespaces/default/pods
```

### Output

```json
{
 "metadata": {
  "resourceVersion": "359"
 },
 "items": [
  {
   "metadata": {
    "name": "pod-w64pz",
    "generateName": "pod-",
    "namespace": "default",
    "uid": "c5a5b6bf-2de8-4cf6-97d1-8c76ab075eb4",
    "resourceVersion": "244",
    "creationTimestamp": "2022-12-22T05:38:40Z",
    "annotations": {
     "scheduler-simulator/filter-result": "{\"node-pprb7\":{\"AzureDiskLimits\":\"passed\",\"EBSLimits\":\"passed\",\"GCEPDLimits\":\"passed\",\"InterPodAffinity\":\"passed\",\"NodeAffinity\":\"passed\",\"NodeName\":\"passed\",\"NodePorts\":\"passed\",\"NodeResourcesFit\":\"passed\",\"NodeUnschedulable\":\"passed\",\"NodeVolumeLimits\":\"passed\",\"PodTopologySpread\":\"passed\",\"TaintToleration\":\"passed\",\"VolumeBinding\":\"passed\",\"VolumeRestrictions\":\"passed\",\"VolumeZone\":\"passed\"}}",
     "scheduler-simulator/finalscore-result": "{}",
     "scheduler-simulator/postFilter-result": "{}",
     "scheduler-simulator/score-result": "{}"
    },
    "managedFields": [
     {
      "manager": "simulator",
      "operation": "Update",
      "apiVersion": "v1",
      "time": "2022-12-22T05:38:40Z",
      "fieldsType": "FieldsV1",
      "fieldsV1": {
       "f:status": {
        "f:conditions": {
         ".": {},
         "k:{\"type\":\"PodScheduled\"}": {
          ".": {},
          "f:lastProbeTime": {},
          "f:lastTransitionTime": {},
          "f:message": {},
          "f:reason": {},
          "f:status": {},
          "f:type": {}
         }
        }
       }
      },
      "subresource": "status"
     },
     {
      "manager": "simulator",
      "operation": "Update",
      "apiVersion": "v1",
      "time": "2022-12-22T05:38:43Z",
      "fieldsType": "FieldsV1",
      "fieldsV1": {
       "f:metadata": {
        "f:annotations": {
         ".": {},
         "f:scheduler-simulator/filter-result": {},
         "f:scheduler-simulator/finalscore-result": {},
         "f:scheduler-simulator/postFilter-result": {},
         "f:scheduler-simulator/score-result": {}
        },
        "f:generateName": {}
       },
       "f:spec": {
        "f:containers": {
         "k:{\"name\":\"pause\"}": {
          ".": {},
          "f:image": {},
          "f:imagePullPolicy": {},
          "f:name": {},
          "f:resources": {
           ".": {},
           "f:limits": {
            ".": {},
            "f:cpu": {},
            "f:memory": {}
           },
           "f:requests": {
            ".": {},
            "f:cpu": {},
            "f:memory": {}
           }
          },
          "f:terminationMessagePath": {},
          "f:terminationMessagePolicy": {}
         }
        },
        "f:dnsPolicy": {},
        "f:enableServiceLinks": {},
        "f:restartPolicy": {},
        "f:schedulerName": {},
        "f:securityContext": {},
        "f:terminationGracePeriodSeconds": {}
       }
      }
     }
    ]
   },
   "spec": {
    "containers": [
     {
      "name": "pause",
      "image": "k8s.gcr.io/pause:3.5",
      "resources": {
       "limits": {
        "cpu": "100m",
        "memory": "16Gi"
       },
       "requests": {
        "cpu": "100m",
        "memory": "16Gi"
       }
      },
      "terminationMessagePath": "/dev/termination-log",
      "terminationMessagePolicy": "File",
      "imagePullPolicy": "IfNotPresent"
     }
    ],
    "restartPolicy": "Always",
    "terminationGracePeriodSeconds": 30,
    "dnsPolicy": "ClusterFirst",
    "nodeName": "node-pprb7",
    "securityContext": {},
    "schedulerName": "default-scheduler",
    "priority": 0,
    "enableServiceLinks": true,
    "preemptionPolicy": "PreemptLowerPriority"
   },
   "status": {
    "phase": "Pending",
    "conditions": [
     {
      "type": "PodScheduled",
      "status": "True",
      "lastProbeTime": null,
      "lastTransitionTime": "2022-12-22T05:38:43Z"
     }
    ],
    "qosClass": "Guaranteed"
   }
  }
 ]
}
```

## Get all pods

* All pods across all namespaces.
* Annotations are not filtered or unmarshalled
* Data structure comes from apimachinery

### curl command
```bash
curl http://localhost:1212/api/v1/pods
```

### Output

```json
{
 "metadata": {
  "resourceVersion": "258"
 },
 "items": [
  {
   "metadata": {
    "name": "pod-w64pz",
    "generateName": "pod-",
    "namespace": "default",
    "uid": "c5a5b6bf-2de8-4cf6-97d1-8c76ab075eb4",
    "resourceVersion": "244",
    "creationTimestamp": "2022-12-22T05:38:40Z",
    "annotations": {
     "scheduler-simulator/filter-result": "{\"node-pprb7\":{\"AzureDiskLimits\":\"passed\",\"EBSLimits\":\"passed\",\"GCEPDLimits\":\"passed\",\"InterPodAffinity\":\"passed\",\"NodeAffinity\":\"passed\",\"NodeName\":\"passed\",\"NodePorts\":\"passed\",\"NodeResourcesFit\":\"passed\",\"NodeUnschedulable\":\"passed\",\"NodeVolumeLimits\":\"passed\",\"PodTopologySpread\":\"passed\",\"TaintToleration\":\"passed\",\"VolumeBinding\":\"passed\",\"VolumeRestrictions\":\"passed\",\"VolumeZone\":\"passed\"}}",
     "scheduler-simulator/finalscore-result": "{}",
     "scheduler-simulator/postFilter-result": "{}",
     "scheduler-simulator/score-result": "{}"
    },
    "managedFields": [
     {
      "manager": "simulator",
      "operation": "Update",
      "apiVersion": "v1",
      "time": "2022-12-22T05:38:40Z",
      "fieldsType": "FieldsV1",
      "fieldsV1": {
       "f:status": {
        "f:conditions": {
         ".": {},
         "k:{\"type\":\"PodScheduled\"}": {
          ".": {},
          "f:lastProbeTime": {},
          "f:lastTransitionTime": {},
          "f:message": {},
          "f:reason": {},
          "f:status": {},
          "f:type": {}
         }
        }
       }
      },
      "subresource": "status"
     },
     {
      "manager": "simulator",
      "operation": "Update",
      "apiVersion": "v1",
      "time": "2022-12-22T05:38:43Z",
      "fieldsType": "FieldsV1",
      "fieldsV1": {
       "f:metadata": {
        "f:annotations": {
         ".": {},
         "f:scheduler-simulator/filter-result": {},
         "f:scheduler-simulator/finalscore-result": {},
         "f:scheduler-simulator/postFilter-result": {},
         "f:scheduler-simulator/score-result": {}
        },
        "f:generateName": {}
       },
       "f:spec": {
        "f:containers": {
         "k:{\"name\":\"pause\"}": {
          ".": {},
          "f:image": {},
          "f:imagePullPolicy": {},
          "f:name": {},
          "f:resources": {
           ".": {},
           "f:limits": {
            ".": {},
            "f:cpu": {},
            "f:memory": {}
           },
           "f:requests": {
            ".": {},
            "f:cpu": {},
            "f:memory": {}
           }
          },
          "f:terminationMessagePath": {},
          "f:terminationMessagePolicy": {}
         }
        },
        "f:dnsPolicy": {},
        "f:enableServiceLinks": {},
        "f:restartPolicy": {},
        "f:schedulerName": {},
        "f:securityContext": {},
        "f:terminationGracePeriodSeconds": {}
       }
      }
     }
    ]
   },
   "spec": {
    "containers": [
     {
      "name": "pause",
      "image": "k8s.gcr.io/pause:3.5",
      "resources": {
       "limits": {
        "cpu": "100m",
        "memory": "16Gi"
       },
       "requests": {
        "cpu": "100m",
        "memory": "16Gi"
       }
      },
      "terminationMessagePath": "/dev/termination-log",
      "terminationMessagePolicy": "File",
      "imagePullPolicy": "IfNotPresent"
     }
    ],
    "restartPolicy": "Always",
    "terminationGracePeriodSeconds": 30,
    "dnsPolicy": "ClusterFirst",
    "nodeName": "node-pprb7",
    "securityContext": {},
    "schedulerName": "default-scheduler",
    "priority": 0,
    "enableServiceLinks": true,
    "preemptionPolicy": "PreemptLowerPriority"
   },
   "status": {
    "phase": "Pending",
    "conditions": [
     {
      "type": "PodScheduled",
      "status": "True",
      "lastProbeTime": null,
      "lastTransitionTime": "2022-12-22T05:38:43Z"
     }
    ],
    "qosClass": "Guaranteed"
   }
  }
 ]
}
```