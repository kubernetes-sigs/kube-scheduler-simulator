<template>
  <v-row v-if="pvcs.length !== 0" no-gutters>
    <v-col>
      <v-card class="ma-2" outlined>
        <v-card-title class="mb-1"> PersistentVolumeClaims </v-card-title>
        <v-card-actions>
          <v-chip
            v-for="(p, i) in pvcs"
            :key="i"
            class="ma-2"
            color="primary"
            outlined
            large
            label
            @click.stop="onClick(p)"
          >
            <img
              src="/pvc.svg"
              height="40"
              alt="p.metadata.name"
              class="mr-2"
            />
            {{ p.metadata.name }}
          </v-chip>
        </v-card-actions>
      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { V1PersistentVolumeClaim } from "@kubernetes/client-node";
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from "@nuxtjs/composition-api";
import {} from "../../lib/util";
import PersistentVolumeClaimStoreKey from "../../StoreKey/PVCStoreKey";
import SnackBarStoreKey from "../../StoreKey/SnackBarStoreKey";
export default defineComponent({
  setup() {
    const store = inject(PersistentVolumeClaimStoreKey);
    if (!store) {
      throw new Error(`${PersistentVolumeClaimStoreKey} is not provided`);
    }

    const snackbarstore = inject(SnackBarStoreKey);
    if (!snackbarstore) {
      throw new Error(`${SnackBarStoreKey} is not provided`);
    }

    const setServerErrorMessage = (error: string) => {
      snackbarstore.setServerErrorMessage(error);
    };

    const getPVCList = async () => {
      await store.fetchlist().catch((e) => setServerErrorMessage(e));
    };
    const onClick = (pvc: V1PersistentVolumeClaim) => {
      store.select(pvc, false);
    };
    onMounted(getPVCList);
    const pvcs = computed(() => store.pvcs);
    return {
      pvcs,
      onClick,
    };
  },
});
</script>
