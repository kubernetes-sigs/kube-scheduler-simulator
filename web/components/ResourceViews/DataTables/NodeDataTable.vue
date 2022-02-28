<template>
  <v-row>
    <v-col>
      <v-card class="ma-2" outlined>
        <v-card-title class="mb-1">
          <v-row
            ><v-col>Nodes<v-spacer></v-spacer> </v-col
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
          :items="nodes"
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
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from "@nuxtjs/composition-api";
import NodeStoreKey from "../../StoreKey/NodeStoreKey";
import { V1Node } from "@kubernetes/client-node";
import PodStoreKey from "../../StoreKey/PodStoreKey";
import {} from "../../lib/util";

export default defineComponent({
  setup() {
    const pstore = inject(PodStoreKey);
    if (!pstore) {
      throw new Error(`${PodStoreKey} is not provided`);
    }

    const nstore = inject(NodeStoreKey);
    if (!nstore) {
      throw new Error(`${NodeStoreKey} is not provided`);
    }

    const getNodeList = async () => {
      await nstore.fetchlist();
    };

    onMounted(getNodeList);

    const nodes = computed(() => nstore.nodes);
    const onClick = (node: V1Node) => {
      nstore.select(node, false);
    };
    const search = "";
    const headers = [
      {
        text: "Name",
        value: "metadata.name",
        sortable: true,
      },
      { text: "CPU", value: "status.capacity.cpu", sortable: true },
      { text: "Memory", value: "status.capacity.memory", sortable: true },
      { text: "Pods", value: "status.capacity.pods", sortable: true },
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
      nodes,
      search,
      headers,
      onClick,
    };
  },
});
</script>
