<template>
  <v-row v-if="storageclasses.length !== 0" no-gutters>
    <v-col>
      <v-card class="ma-2" outlined>
        <v-card-title class="mb-1"> StorageClasses </v-card-title>
        <v-card-actions>
          <v-chip
            v-for="(p, i) in storageclasses"
            :key="i"
            class="ma-2"
            color="primary"
            outlined
            large
            label
            @click.stop="onClick(p)"
          >
            <img src="/sc.svg" height="40" alt="p.metadata.name" class="mr-2" />
            {{ p.metadata.name }}
          </v-chip>
        </v-card-actions>
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
    return {
      storageclasses,
      onClick,
    };
  },
});
</script>
