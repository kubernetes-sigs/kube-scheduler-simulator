<template>
  <v-sheet class="transparent">
    <v-btn color="primary ma-2" dark @click="reset()">
      Reset Alll Resources</v-btn
    >
  </v-sheet>
</template>

<script lang="ts">
import { inject, defineComponent } from "@nuxtjs/composition-api";
import ResetStoreKey from "./StoreKey/ResetStoreKey";
import PodStoreKey from "./StoreKey/PodStoreKey";
import NodeStoreKey from "./StoreKey/NodeStoreKey";
import PersistentVolumeStoreKey from "./StoreKey/PVStoreKey";
import PersistentVolumeClaimStoreKey from "./StoreKey/PVCStoreKey";
import StorageClassStoreKey from "./StoreKey/StorageClassStoreKey";
import PriorityClassStoreKey from "./StoreKey/PriorityClassStoreKey";
import SchedulerConfigurationStoreKey from "./StoreKey/SchedulerConfigurationStoreKey";
import SnackBarStoreKey from "./StoreKey/SnackBarStoreKey";

export default defineComponent({
  setup() {
    const resetstore = inject(ResetStoreKey);
    if (!resetstore) {
      throw new Error(`${ResetStoreKey} is not provided`);
    }

    const podstore = inject(PodStoreKey);
    if (!podstore) {
      throw new Error(`${PodStoreKey} is not provided`);
    }

    const nodestore = inject(NodeStoreKey);
    if (!nodestore) {
      throw new Error(`${NodeStoreKey} is not provided`);
    }

    const pvstore = inject(PersistentVolumeStoreKey);
    if (!pvstore) {
      throw new Error(`${pvstore} is not provided`);
    }

    const pvcstore = inject(PersistentVolumeClaimStoreKey);
    if (!pvcstore) {
      throw new Error(`${pvcstore} is not provided`);
    }

    const storageclassstore = inject(StorageClassStoreKey);
    if (!storageclassstore) {
      throw new Error(`${StorageClassStoreKey} is not provided`);
    }

    const priorityclassstore = inject(PriorityClassStoreKey);
    if (!priorityclassstore) {
      throw new Error(`${PriorityClassStoreKey} is not provided`);
    }

    const schedulerconfigurationstore = inject(SchedulerConfigurationStoreKey);
    if (!schedulerconfigurationstore) {
      throw new Error(`${SchedulerConfigurationStoreKey} is not provided`);
    }

    const snackbarstore = inject(SnackBarStoreKey);
    if (!snackbarstore) {
      throw new Error(`${SnackBarStoreKey} is not provided`);
    }

    const setServerErrorMessage = (error: string) => {
      snackbarstore.setServerErrorMessage(error);
    };

    const setInfoMessage = (message: string) => {
      snackbarstore.setServerInfoMessage(message);
    };

    const reset = async () => {
      await resetstore.reset().catch((error) => {
        setServerErrorMessage(error);
      });
      await nodestore.fetchlist();
      await podstore.fetchlist();
      await pvstore.fetchlist();
      await pvcstore.fetchlist();
      await storageclassstore.fetchlist();
      await priorityclassstore.fetchlist();
      setInfoMessage("Successfully reset all resources");
    };

    return {
      reset,
    };
  },
});
</script>
