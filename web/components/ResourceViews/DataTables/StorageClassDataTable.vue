<template>
  <DataTable
    :title="`StorageClasses`"
    :headers="headers"
    :items="storageclasses"
    :on-click="onClick"
  />
</template>

<script lang="ts">
import { V1StorageClass } from "@kubernetes/client-node";
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from "@nuxtjs/composition-api";
import DataTable from "./DataTable.vue";
import StorageClassStoreKey from "../../StoreKey/StorageClassStoreKey";
import SnackBarStoreKey from "../../StoreKey/SnackBarStoreKey";
import {} from "../../lib/util";

export default defineComponent({
  components: {
    DataTable,
  },
  setup() {
    const store = inject(StorageClassStoreKey);
    if (!store) {
      throw new Error(`${StorageClassStoreKey} is not provided`);
    }

    const snackbarstore = inject(SnackBarStoreKey);
    if (!snackbarstore) {
      throw new Error(`${SnackBarStoreKey} is not provided`);
    }

    const setServerErrorMessage = (error: string) => {
      snackbarstore.setServerErrorMessage(error);
    };

    const getStorageClassList = async () => {
      await store.fetchlist().catch((e) => setServerErrorMessage(e));
    };
    const onClick = (storageclass: V1StorageClass) => {
      store.select(storageclass, false);
    };
    onMounted(getStorageClassList);
    const storageclasses = computed(() => store.storageclasses);
    const search = "";
    const headers = [
      {
        text: "Name",
        value: "metadata.name",
        sortable: true,
      },
      { text: "Provisioner", value: "provisioner", sortable: true },
      { text: "Parameters", value: "parameters", sortable: true },
      { text: "Reclaim-Policy", value: "reclaimPolicy", sortable: true },
      {
        text: "VolumeBindingMode",
        value: "volumeBindingMode",
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
      storageclasses,
      headers,
      search,
      onClick,
    };
  },
});
</script>
