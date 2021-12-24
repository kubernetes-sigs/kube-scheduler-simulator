<template>
  <v-sheet class="transparent">
    <v-btn
      v-for="(rn, i) in resourceNames"
      :key="i"
      color="primary ma-2"
      dark
      @click="create(rn)"
    >
      New {{ rn }}
    </v-btn>
  </v-sheet>
</template>

<script lang="ts">
import { ref, inject, defineComponent } from "@nuxtjs/composition-api";
import {
  podTemplate,
  nodeTemplate,
  pvTemplate,
  pvcTemplate,
  storageclassTemplate,
  priorityclassTemplate,
} from "./lib/template";
import {} from "./lib/util";
import PodStoreKey from "./StoreKey/PodStoreKey";
import NodeStoreKey from "./StoreKey/NodeStoreKey";
import PersistentVolumeStoreKey from "./StoreKey/PVStoreKey";
import PersistentVolumeClaimStoreKey from "./StoreKey/PVCStoreKey";
import StorageClassStoreKey from "./StoreKey/StorageClassStoreKey";
import PriorityClassStoreKey from "./StoreKey/PriorityClassStoreKey";
import {
  V1Node,
  V1PersistentVolumeClaim,
  V1PersistentVolume,
  V1Pod,
  V1StorageClass,
  V1PriorityClass,
} from "@kubernetes/client-node";

type Resource =
  | V1Pod
  | V1Node
  | V1PersistentVolumeClaim
  | V1PersistentVolume
  | V1StorageClass
  | V1PriorityClass;

interface Store {
  readonly selected: object | null;
  readonly count: number;
  select(_resource: Resource | null, _isNew: boolean): void;
}

export default defineComponent({
  setup() {
    var store: Store | null = null;

    const podstore = inject(PodStoreKey);
    if (!podstore) {
      throw new Error(`${PodStoreKey} is not provided`);
    }

    const nodestore = inject(NodeStoreKey);
    if (!nodestore) {
      throw new Error(`${NodeStoreKey} is not provided`);
    }

    const pvstore = inject(PersistentVolumeStoreKey);
    if (!pvstore) {
      throw new Error(`${pvstore} is not provided`);
    }

    const pvcstore = inject(PersistentVolumeClaimStoreKey);
    if (!pvcstore) {
      throw new Error(`${pvcstore} is not provided`);
    }

    const storageclassstore = inject(StorageClassStoreKey);
    if (!storageclassstore) {
      throw new Error(`${StorageClassStoreKey} is not provided`);
    }

    const priorityclassstore = inject(PriorityClassStoreKey);
    if (!priorityclassstore) {
      throw new Error(`${PriorityClassStoreKey} is not provided`);
    }

    const dialog = ref(false);
    const resourceNames = [
      "StorageClass",
      "PersistentVolumeClaim",
      "PersistentVolume",
      "Node",
      "Pod",
      "PriorityClass",
    ];

    const create = (rn: string) => {
      var targetTemplate: Resource | null = null;
      switch (rn) {
        case "Pod":
          store = podstore;
          // if store.count = 0, name suffix is 1.
          targetTemplate = podTemplate((store.count + 1).toString());
          break;
        case "Node":
          store = nodestore;
          // if store.count = 0, name suffix is 1.
          targetTemplate = nodeTemplate((store.count + 1).toString());
          break;
        case "PersistentVolume":
          store = pvstore;
          // if store.count = 0, name suffix is 1.
          targetTemplate = pvTemplate((store.count + 1).toString());
          break;
        case "PersistentVolumeClaim":
          store = pvcstore;
          // if store.count = 0, name suffix is 1.
          targetTemplate = pvcTemplate((store.count + 1).toString());
          break;
        case "StorageClass":
          store = storageclassstore;
          // if store.count = 0, name suffix is 1.
          targetTemplate = storageclassTemplate((store.count + 1).toString());
          break;
        case "PriorityClass":
          store = priorityclassstore;
          // if store.count = 2, name suffix is 1.
          targetTemplate = priorityclassTemplate((store.count - 1).toString());
          break;
      }

      if (store) {
        store.select(targetTemplate, true);
      }
      dialog.value = false;
    };

    return {
      create,
      dialog,
      resourceNames,
    };
  },
});
</script>
