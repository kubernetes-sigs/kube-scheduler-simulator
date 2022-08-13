<template>
  <v-card v-if="nodes.length !== 0" class="ma-2" outlined>
    <v-card-title class="mb-1"> Nodes </v-card-title>
    <v-container>
      <v-row no-gutters>
        <v-col v-for="(n, i) in nodes" :key="i" tile cols="auto">
          <v-card class="ma-2" outlined @click="onClick(n)">
            <v-card-title>
              <img
                src="/node.svg"
                height="40"
                alt="p.metadata.name"
                class="mr-2"
              />
              {{ n.metadata.name }}
            </v-card-title>
            <PodList :node-name="n.metadata.name" />
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </v-card>
</template>

<script lang="ts">
import { computed, inject, defineComponent } from "@nuxtjs/composition-api";
import NodeStoreKey from "../../StoreKey/NodeStoreKey";
import PodList from "./PodList.vue";
import { V1Node } from "@kubernetes/client-node";
import {} from "../../lib/util";

export default defineComponent({
  components: { PodList },
  setup() {
    const nstore = inject(NodeStoreKey);
    if (!nstore) {
      throw new Error(`${NodeStoreKey} is not provided`);
    }

    const nodes = computed(() => nstore.nodes);

    const onClick = (node: V1Node) => {
      nstore.select(node, false);
    };

    return {
      nodes,
      onClick,
    };
  },
});
</script>
