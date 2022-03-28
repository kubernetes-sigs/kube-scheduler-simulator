<template>
  <v-card v-if="pods['unscheduled'].length !== 0" class="ma-2" outlined>
    <v-card-title class="mb-1"> Unscheduled Pods </v-card-title>
    <PodList node-name="unscheduled" />
  </v-card>
</template>

<script lang="ts">
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from "@nuxtjs/composition-api";
import {} from "../../lib/util";
import PodStoreKey from "../../StoreKey/PodStoreKey";
import PodList from "./PodList.vue";
import SnackBarStoreKey from "../../StoreKey/SnackBarStoreKey";
export default defineComponent({
  components: { PodList },
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
    onMounted(getPodList);

    const pods = computed(() => store.pods);
    return {
      pods,
    };
  },
});
</script>
