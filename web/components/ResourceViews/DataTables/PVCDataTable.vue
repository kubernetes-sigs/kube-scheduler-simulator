<template>
  <v-row>
    <v-col>
      <v-card class="ma-2" outlined>
        <v-card-title class="mb-1">
          PersistentVolumeClaims <v-spacer></v-spacer>
          <v-text-field
            v-model="search"
            append-icon="mdi-magnify"
            label="Search"
            single-line
            hide-details
          ></v-text-field></v-card-title
        ><v-data-table
          :headers="headers"
          :items="pvcs"
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
import { V1PersistentVolumeClaim } from "@kubernetes/client-node";
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from "@nuxtjs/composition-api";
import {} from "../../lib/util";
import PersistentVolumeClaimStoreKey from "../../StoreKey/PVCStoreKey";
export default defineComponent({
  setup() {
    const store = inject(PersistentVolumeClaimStoreKey);
    if (!store) {
      throw new Error(`${PersistentVolumeClaimStoreKey} is not provided`);
    }

    const getPVCList = async () => {
      await store.fetchlist();
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
        text: "Creation-Time",
        value: "metadata.creationTimestamp",
        sortable: true,
      },
      {
        text: "Update-Time",
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
