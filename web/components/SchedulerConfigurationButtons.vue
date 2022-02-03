<template>
  <v-sheet class="transparent">
    <v-btn class="ma-2" @click="onClick(buttonName[0])">
      <v-icon> mdi-cog </v-icon>
    </v-btn>
    <ExportButton />
    <ImportButton />
  </v-sheet>
</template>

<script lang="ts">
import { inject, defineComponent } from "@nuxtjs/composition-api";
import SchedulerConfigurationStoreKey from "./StoreKey/SchedulerConfigurationStoreKey";
import ExportButton from "./ExportButton.vue";
import ImportButton from "./ImportButton.vue";

export default defineComponent({
  components: {
    ImportButton,
    ExportButton,
  },
  setup() {
    const schedulerconfigurationstore = inject(SchedulerConfigurationStoreKey);
    if (!schedulerconfigurationstore) {
      throw new Error(`${SchedulerConfigurationStoreKey} is not provided`);
    }

    const buttonName = ["SchedulerConfiguration"];
    const onClick = (bn: string) => {
      switch (bn) {
        case "SchedulerConfiguration":
          schedulerconfigurationstore.fetchSelected();
          break;
      }
    };
    return {
      onClick,
      buttonName,
    };
  },
});
</script>
