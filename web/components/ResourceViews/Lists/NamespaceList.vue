<template>
  <v-row v-if="namespaces.length !== 0" no-gutters>
    <v-col>
      <v-card class="ma-2" outlined>
        <v-card-title class="mb-1"> Namespaces </v-card-title>
        <v-card-actions>
          <v-chip
            v-for="(p, i) in namespaces"
            :key="i"
            class="ma-2"
            color="primary"
            outlined
            large
            label
            @click.stop="onClick(p)"
          >
            {{ p.metadata.name }}
          </v-chip>
        </v-card-actions>
      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { V1Namespace } from "@kubernetes/client-node";
import { computed, inject, defineComponent } from "@nuxtjs/composition-api";
import NamespaceStoreKey from "../../StoreKey/NamespaceStoreKey";
export default defineComponent({
  setup() {
    const store = inject(NamespaceStoreKey);
    if (!store) {
      throw new Error(`${NamespaceStoreKey.description} is not provided`);
    }

    const onClick = (ns: V1Namespace) => {
      store.select(ns, false)
    };
    const namespaces = computed(() => store.namespaces);
    return {
      namespaces,
      onClick,
    }
  }
})
</script>

