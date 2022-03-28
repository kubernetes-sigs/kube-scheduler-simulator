<template>
  <DataTable
    :title="`PersistentVolumeClaims`"
    :headers="headers"
    :items="pvcs"
    :on-click="onClick"
  />
</template>

<script lang="ts">
import { V1PersistentVolumeClaim } from "@kubernetes/client-node";
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from "@nuxtjs/composition-api";
import DataTable from "./DataTable.vue";
import {} from "../../lib/util";
import PersistentVolumeClaimStoreKey from "../../StoreKey/PVCStoreKey";
import SnackBarStoreKey from "../../StoreKey/SnackBarStoreKey";
export default defineComponent({
  components: {
    DataTable,
  },
  setup() {
    const store = inject(PersistentVolumeClaimStoreKey);
    if (!store) {
      throw new Error(`${PersistentVolumeClaimStoreKey} is not provided`);
    }

    const snackbarstore = inject(SnackBarStoreKey);
    if (!snackbarstore) {
      throw new Error(`${SnackBarStoreKey} is not provided`);
    }

    const setServerErrorMessage = (error: string) => {
      snackbarstore.setServerErrorMessage(error);
    };

    const getPVCList = async () => {
      await store.fetchlist().catch((e) => setServerErrorMessage(e));
    };
    const onClick = (pvc: V1PersistentVolumeClaim) => {
      store.select(pvc, false);
    };
    onMounted(getPVCList);
    const pvcs = computed(() => store.pvcs);
    const search = "";
    const headers = [
      {
        text: "Name",
        value: "metadata.name",
        sortable: true,
      },
      { text: "Namespace", value: "metadata.namespace", sortable: true },
      { text: "VolumeName", value: "spec.volumeName", sortable: true },
      { text: "VolumeMode", value: "spec.volumeMode", sortable: true },
      {
        text: "Capacity",
        value: "spec.resources.requests.storage",
        sortable: true,
      },
      {
        text: "CreationTime",
        value: "metadata.creationTimestamp",
        sortable: true,
      },
      {
        text: "UpdateTime",
        value: "metadata.managedFields[0].time",
        sortable: true,
      },
    ];
    return {
      pvcs,
      search,
      headers,
      onClick,
    };
  },
});
</script>
