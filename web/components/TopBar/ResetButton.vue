<template>
  <v-dialog v-model="data.dialog" width="500">
    <template #activator="{ on }">
      <v-btn class="ma-2" color="error" v-on="on"> Reset </v-btn>
    </template>

    <v-card>
      <v-card-title class="2">
        Are you sure to reset all resources and scheduler configuration?
      </v-card-title>
      <v-divider></v-divider>
      <v-divider></v-divider>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="green darken-1" text @click="reset"> Reset </v-btn>
        <v-btn color="green darken-1" text @click="data.dialog = false">
          Cancel
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { inject, defineComponent, reactive } from "@nuxtjs/composition-api";
import ResetStoreKey from "../StoreKey/ResetStoreKey";
import PodStoreKey from "../StoreKey/PodStoreKey";
import NodeStoreKey from "../StoreKey/NodeStoreKey";
import PersistentVolumeStoreKey from "../StoreKey/PVStoreKey";
import PersistentVolumeClaimStoreKey from "../StoreKey/PVCStoreKey";
import StorageClassStoreKey from "../StoreKey/StorageClassStoreKey";
import PriorityClassStoreKey from "../StoreKey/PriorityClassStoreKey";
import SchedulerConfigurationStoreKey from "../StoreKey/SchedulerConfigurationStoreKey";
import SnackBarStoreKey from "../StoreKey/SnackBarStoreKey";

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
      await Promise.all([
        nodestore.fetchlist(),
        podstore.fetchlist(),
        pvstore.fetchlist(),
        pvcstore.fetchlist(),
        storageclassstore.fetchlist(),
        priorityclassstore.fetchlist(),
      ]).catch((error) => {
        setServerErrorMessage(error);
      });
      setInfoMessage("Successfully reset all resources");
      data.dialog = false;
    };

    const data = reactive({
      dialog: false,
    });

    return {
      reset,
      data,
    };
  },
});
</script>
