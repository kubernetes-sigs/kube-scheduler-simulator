<template>
  <v-card-actions>
    <v-chip
      v-for="(p, i) in pods[nodeName]"
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
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from "@nuxtjs/composition-api";
import {} from "../../lib/util";
import PodStoreKey from "../../StoreKey/PodStoreKey";
import SnackBarStoreKey from "../../StoreKey/SnackBarStoreKey";
export default defineComponent({
  props: {
    nodeName: {
      type: String,
      required: true,
    },
  },
  setup() {
    const store = inject(PodStoreKey);
    if (!store) {
      throw new Error(`${PodStoreKey} is not provided`);
    }

    const snackbarstore = inject(SnackBarStoreKey);
    if (!snackbarstore) {
      throw new Error(`${SnackBarStoreKey} is not provided`);
    }

    const setServerErrorMessage = (error: string) => {
      snackbarstore.setServerErrorMessage(error);
    };

    const getPodList = async () => {
      await store.fetchlist().catch((e) => setServerErrorMessage(e));
    };
    const onClick = (pod: V1Pod) => {
      store.select(pod, false);
    };
    onMounted(getPodList);
    const pods = computed(() => store.pods);
    return {
      pods,
      onClick,
    };
  },
});
</script>
