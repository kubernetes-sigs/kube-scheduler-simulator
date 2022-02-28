<template>
  <v-row>
    <v-col>
      <v-card class="ma-2" outlined>
        <v-card-title class="mb-1">
          <v-row
            ><v-col>PersistentVolumes<v-spacer></v-spacer> </v-col
            ><v-col>
              <v-text-field
                v-model="search"
                append-icon="mdi-magnify"
                label="Search"
                single-line
                hide-details
              ></v-text-field></v-col></v-row></v-card-title
        ><v-data-table
          :headers="headers"
          :items="pvs"
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
import { V1PersistentVolume } from "@kubernetes/client-node";
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from "@nuxtjs/composition-api";
import {} from "../../lib/util";
import PersistentVolumeStoreKey from "../../StoreKey/PVStoreKey";
export default defineComponent({
  setup() {
    const store = inject(PersistentVolumeStoreKey);
    if (!store) {
      throw new Error(`${PersistentVolumeStoreKey} is not provided`);
    }

    const getPVList = async () => {
      await store.fetchlist();
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
