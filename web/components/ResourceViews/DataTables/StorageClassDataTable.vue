<template>
  <v-row>
    <v-col>
      <v-card class="ma-2" outlined>
        <v-card-title class="mb-1">
          StorageClasses <v-spacer></v-spacer>
          <v-text-field
            v-model="search"
            append-icon="mdi-magnify"
            label="Search"
            single-line
            hide-details
          ></v-text-field
        ></v-card-title>
        <v-data-table
          :headers="headers"
          :items="storageclasses"
          :items-per-page="5"
          :search="search"
          multi-sort
          @click:row="onClick"
        ></v-data-table>
      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { V1StorageClass } from "@kubernetes/client-node";
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from "@nuxtjs/composition-api";
import {} from "../../lib/util";
import StorageClassStoreKey from "../../StoreKey/StorageClassStoreKey";
export default defineComponent({
  setup() {
    const store = inject(StorageClassStoreKey);
    if (!store) {
      throw new Error(`${StorageClassStoreKey} is not provided`);
    }

    const getStorageClassList = async () => {
      await store.fetchlist();
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
