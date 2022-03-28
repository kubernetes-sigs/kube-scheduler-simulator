<template>
  <DataTable
    :title="`PersistentVolumes`"
    :headers="headers"
    :items="pvs"
    :on-click="onClick"
  />
</template>

<script lang="ts">
import { V1PersistentVolume } from "@kubernetes/client-node";
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from "@nuxtjs/composition-api";
import DataTable from "./DataTable.vue";
import PersistentVolumeStoreKey from "../../StoreKey/PVStoreKey";
import SnackBarStoreKey from "../../StoreKey/SnackBarStoreKey";
import {} from "../../lib/util";

export default defineComponent({
  components: {
    DataTable,
  },
  setup() {
    const store = inject(PersistentVolumeStoreKey);
    if (!store) {
      throw new Error(`${PersistentVolumeStoreKey} is not provided`);
    }

    const snackbarstore = inject(SnackBarStoreKey);
    if (!snackbarstore) {
      throw new Error(`${SnackBarStoreKey} is not provided`);
    }

    const setServerErrorMessage = (error: string) => {
      snackbarstore.setServerErrorMessage(error);
    };

    const getPVList = async () => {
      await store.fetchlist().catch((e) => setServerErrorMessage(e));
    };
    const onClick = (pv: V1PersistentVolume) => {
      store.select(pv, false);
    };
    onMounted(getPVList);
    const pvs = computed(() => store.pvs);
    const search = "";
    const headers = [
      {
        text: "Name",
        value: "metadata.name",
        sortable: true,
      },
      { text: "Status", value: "status.phase", sortable: true },
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
      pvs,
      search,
      headers,
      onClick,
    };
  },
});
</script>
