<template>
  <v-row>
    <v-col>
      <v-card class="ma-2" outlined>
        <v-card-title class="mb-1">
          Pods <v-spacer></v-spacer>
          <v-text-field
            v-model="search"
            append-icon="mdi-magnify"
            label="Search"
            single-line
            hide-details
          ></v-text-field></v-card-title
        ><v-data-table
          :headers="headers"
          :items="pods"
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
import { V1Pod } from "@kubernetes/client-node";
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from "@nuxtjs/composition-api";
import {} from "../../lib/util";
import PodStoreKey from "../../StoreKey/PodStoreKey";

export default defineComponent({
  setup() {
    const store = inject(PodStoreKey);
    if (!store) {
      throw new Error(`${PodStoreKey} is not provided`);
    }

    const getPodList = async () => {
      await store.fetchlist();
    };
    const onClick = (pod: V1Pod) => {
      store.select(pod, false);
    };
    onMounted(getPodList);
    const pods = computed(() => {
      return Array<V1Pod>().concat(
        ...Object.values(store.pods).map((p) => {
          return p;
        })
      );
    });
    const search = "";
    const headers = [
      {
        text: "Name",
        value: "metadata.name",
        sortable: true,
      },
      { text: "Namespace", value: "metadata.namespace", sortable: true },
      { text: "Node", value: "spec.nodeName", sortable: true },
      {
        text: "Conditions",
        value: "status.conditions[0].type",
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
      pods,
      search,
      onClick,
      headers,
    };
  },
});
</script>
