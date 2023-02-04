<template>
  <v-dialog v-model="data.dialog" width="500">
    <template #activator="{ on }">
      <v-btn color="ma-2" :retain-focus-on-click="false" v-on="on">
        Export
      </v-btn>
    </template>

    <v-card>
      <v-card-title class="text-h5 grey lighten-2"> Export </v-card-title>

      <v-card-text>
        Export the current created resources and scheduler configuration.
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="green darken-1" text @click="data.dialog = false">
          Cancel
        </v-btn>
        <v-btn color="green darken-1" text @click="ExportScheduler()">
          Export
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { saveAs } from "file-saver";
import { defineComponent, inject, reactive } from "@nuxtjs/composition-api";
import SnackBarStoreKey from "../StoreKey/SnackBarStoreKey";
import yaml from "js-yaml";
import { ExportAPIKey } from "~/api/APIProviderKeys";

export default defineComponent({
  setup() {
    const data = reactive({
      dialog: false,
    });

    const exportAPI = inject(ExportAPIKey);
    if (!exportAPI) {
      throw new Error(`${ExportAPIKey.description} is not provided`);
    }

    const snackbarstore = inject(SnackBarStoreKey);
    if (!snackbarstore) {
      throw new Error(`${SnackBarStoreKey.description} is not provided`);
    }
    const setServerErrorMessage = (error: string) => {
      snackbarstore.setServerErrorMessage(error);
    };

    const ExportScheduler = async () => {
      try {
        const c = await exportAPI.exportScheduler();
        if (c) {
          const blob = new Blob([yaml.dump(c)], {
            type: "application/yaml",
          });
          saveAs(blob, "export.yml");
          data.dialog = false;
        }
      } catch (e: any) {
        setServerErrorMessage(e);
      }
    };
    return {
      data,
      ExportScheduler,
    };
  },
});
</script>
