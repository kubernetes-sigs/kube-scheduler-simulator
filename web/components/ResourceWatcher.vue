<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, inject, onMounted } from "@nuxtjs/composition-api";
import PodStoreKey from "./StoreKey/PodStoreKey";
import NodeStoreKey from "./StoreKey/NodeStoreKey";
import PersistentVolumeClaimStoreKey from "./StoreKey/PVCStoreKey";
import PersistentVolumeStoreKey from "./StoreKey/PVStoreKey";
import PriorityClassStoreKey from "./StoreKey/PriorityClassStoreKey";
import StorageClassStoreKey from "./StoreKey/StorageClassStoreKey";
import SnackBarStoreKey from "./StoreKey/SnackBarStoreKey"
import { WatcherAPIKey } from "~/api/APIProviderKeys";
import { WatchEventType } from "@/types/resources";
import { LastResourceVersions } from "@/types/api/v1";
import {
  V1Node,
  V1PersistentVolume,
  V1PersistentVolumeClaim,
  V1Pod,
  V1PriorityClass,
  V1StorageClass,
} from "@kubernetes/client-node";

export default defineComponent({
  setup() {
    const watcherAPI = inject(WatcherAPIKey);
    if (!watcherAPI) {
      throw new Error(`${watcherAPI} is not provided`);
    }
    const pstore = inject(PodStoreKey);
    if (!pstore) {
      throw new Error(`${PodStoreKey} is not provided`);
    }
    const nstore = inject(NodeStoreKey);
    if (!nstore) {
      throw new Error(`${NodeStoreKey} is not provided`);
    }
    const pvcstore = inject(PersistentVolumeClaimStoreKey);
    if (!pvcstore) {
      throw new Error(`${pvcstore} is not provided`);
    }
    const pvstore = inject(PersistentVolumeStoreKey);
    if (!pvstore) {
      throw new Error(`${pvstore} is not provided`);
    }
    const priorityclassstore = inject(PriorityClassStoreKey);
    if (!priorityclassstore) {
      throw new Error(`${PriorityClassStoreKey} is not provided`);
    }
    const storageclassstore = inject(StorageClassStoreKey);
    if (!storageclassstore) {
      throw new Error(`${StorageClassStoreKey} is not provided`);
    }
    const snackbarstore = inject(SnackBarStoreKey);
    if (!snackbarstore) {
      throw new Error(`${SnackBarStoreKey} is not provided`);
    }

    // Initializes each resource and starts watching.
    onMounted(async () => {
      await pstore.initList();
      await nstore.initList();
      await pvcstore.initList();
      await pvstore.initList();
      await priorityclassstore.initList();
      await storageclassstore.initList();
      await watchAndUpdates();
    });

    const createLastResourceVersions = (): LastResourceVersions => {
      return {
        pods: pstore.lastResourceVersion,
        nodes: nstore.lastResourceVersion,
        pvs: pvstore.lastResourceVersion,
        pvcs: pvcstore.lastResourceVersion,
        storageClasses: storageclassstore.lastResourceVersion,
        priorityClasses: priorityclassstore.lastResourceVersion,
      } as LastResourceVersions;
    };

    // Call watch API and allocates the event to each resource's handler.
    const watchAndUpdates = () => {
      watcherAPI
        .watchResources(createLastResourceVersions() as LastResourceVersions)
        .then((response) => {
          if (!response.body) {
            return;
          }
          const stream = response.body.getReader();
          const utf8Decoder = new TextDecoder("utf-8");
          let buffer = "";

          return stream.read().then(function processText({ done, value }): any {
            if (done) {
              snackbarstore.setServerErrorMessage(
                "The watch stream is terminated. Please reload your browser if you don't expect this."
              );
              return;
            }
            buffer += utf8Decoder.decode(value);
            buffer = onNewLine(buffer, async (chunk: string) => {
              if (chunk.trim().length === 0) {
                return;
              }
              try {
                const event = JSON.parse(chunk) as WatchEvent;
                switch (event.Kind) {
                  case ResourceKind.PODS: {
                    pstore.watchEventHandler(
                      event.EventType,
                      event.Obj as V1Pod
                    );
                    pstore.setLastResourceVersion(event.Obj as V1Pod);
                    break;
                  }
                  case ResourceKind.NODES: {
                    nstore.watchEventHandler(
                      event.EventType,
                      event.Obj as V1Node
                    );
                    nstore.setLastResourceVersion(event.Obj as V1Node);
                    break;
                  }
                  case ResourceKind.PVS: {
                    pvstore.watchEventHandler(
                      event.EventType,
                      event.Obj as V1PersistentVolume
                    );
                    pvstore.setLastResourceVersion(
                      event.Obj as V1PersistentVolume
                    );
                    break;
                  }
                  case ResourceKind.PVCS: {
                    pvcstore.watchEventHandler(
                      event.EventType,
                      event.Obj as V1PersistentVolumeClaim
                    );
                    pvcstore.setLastResourceVersion(
                      event.Obj as V1PersistentVolumeClaim
                    );
                    break;
                  }
                  case ResourceKind.SCS: {
                    storageclassstore.watchEventHandler(
                      event.EventType,
                      event.Obj as V1StorageClass
                    );
                    storageclassstore.setLastResourceVersion(
                      event.Obj as V1StorageClass
                    );
                    break;
                  }
                  case ResourceKind.PCS: {
                    priorityclassstore.watchEventHandler(
                      event.EventType,
                      event.Obj as V1PriorityClass
                    );
                    priorityclassstore.setLastResourceVersion(
                      event.Obj as V1PriorityClass
                    );
                    break;
                  }
                }
              } catch (error) {
                console.log("Error while parsing", chunk, "\n", error);
              }
            });
            return stream.read().then(processText);
          });
        })
        .catch(() => {
          console.log("Error! Retrying in 5 seconds...");
          // Call the watch API again if some error occurs.
          setTimeout(() => watchAndUpdates(), 5000);
        });
    };
    return {};
  },
});

type WatchEvent = {
  Kind: ResourceKind;
  EventType: WatchEventType;
  Obj: Object;
};

enum ResourceKind {
  PODS = "pods",
  NODES = "nodes",
  PVS = "persistentvolumes",
  PVCS = "persistentvolumeclaims",
  SCS = "storageclasses",
  PCS = "priorityclasses",
}

function onNewLine(buffer: string, fn: Function): string {
  const newLineIndex = buffer.indexOf("\n");
  if (newLineIndex === -1) {
    return buffer;
  }
  const chunk = buffer.slice(0, buffer.indexOf("\n"));
  const newBuffer = buffer.slice(buffer.indexOf("\n") + 1);
  fn(chunk);
  return onNewLine(newBuffer, fn);
}
</script>
