<template>
  <v-row v-if="priorityclasses.length !== 0" no-gutters>
    <v-col>
      <v-card class="ma-2" outlined>
        <v-card-title class="mb-1"> PriorityClasses </v-card-title>
        <v-card-actions>
          <v-chip
            v-for="(p, i) in priorityclasses"
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
import { V1PriorityClass } from "@kubernetes/client-node";
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from "@nuxtjs/composition-api";
import {} from "../../lib/util";
import PriorityClassStoreKey from "../../StoreKey/PriorityClassStoreKey";
import SnackBarStoreKey from "../../StoreKey/SnackBarStoreKey";
export default defineComponent({
  setup() {
    const store = inject(PriorityClassStoreKey);
    if (!store) {
      throw new Error(`${PriorityClassStoreKey} is not provided`);
    }

    const snackbarstore = inject(SnackBarStoreKey);
    if (!snackbarstore) {
      throw new Error(`${SnackBarStoreKey} is not provided`);
    }

    const setServerErrorMessage = (error: string) => {
      snackbarstore.setServerErrorMessage(error);
    };

    const getPriorityClassList = async () => {
      await store.fetchlist().catch((e) => setServerErrorMessage(e));
    };
    const onClick = (priorityclass: V1PriorityClass) => {
      store.select(priorityclass, false);
    };
    onMounted(getPriorityClassList);
    const priorityclasses = computed(() => store.priorityclasses);
    return {
      priorityclasses,
      onClick,
    };
  },
});
</script>
