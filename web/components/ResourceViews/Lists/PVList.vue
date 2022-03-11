<template>
  <v-row v-if="pvs.length !== 0" no-gutters>
    <v-col>
      <v-card class="ma-2" outlined>
        <v-card-title class="mb-1"> PersistentVolumes </v-card-title>
        <v-card-actions>
          <v-chip
            v-for="(p, i) in pvs"
            :key="i"
            class="ma-2"
            color="primary"
            outlined
            large
            label
            @click.stop="onClick(p)"
          >
            <img src="/pv.svg" height="40" alt="p.metadata.name" class="mr-2" />
            {{ p.metadata.name }}
          </v-chip>
        </v-card-actions>
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
    return {
      pvs,
      onClick,
    };
  },
});
</script>
