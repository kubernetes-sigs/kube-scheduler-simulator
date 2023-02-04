<template>
  <v-card-actions>
    <v-chip
      v-for="(p, i) in pods"
      :key="i"
      class="ma-2"
      color="primary"
      outlined
      large
      label
      @click.stop="onClick(p)"
    >
      <img src="/pod.svg" height="40" alt="p.metadata.name" class="mr-2" />
      {{ p.metadata.name }}
    </v-chip>
  </v-card-actions>
</template>

<script lang="ts">
import { V1Pod } from "@kubernetes/client-node";
import { computed, inject, defineComponent } from "@nuxtjs/composition-api";
import {} from "../../lib/util";
import PodStoreKey from "../../StoreKey/PodStoreKey";
export default defineComponent({
  props: {
    nodeName: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const store = inject(PodStoreKey);
    if (!store) {
      throw new Error(`${PodStoreKey.description} is not provided`);
    }

    const onClick = (pod: V1Pod) => {
      store.select(pod, false);
    };

    const pods: any = computed(function () {
      return store.pods[props.nodeName];
    });
    return {
      pods,
      onClick,
    };
  },
});
</script>
