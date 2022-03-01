<template>
  <v-form>
    <v-container>
      <v-row no-gutters>
        <v-col cols="12" sm="6" md="4" lg="3">
          <v-text-field
            class="custom-input-color"
            readonly
            append-icon="mdi-cog"
            outlined
            :label="version"
            value="SCHEDULER CONFIGURATION"
            @click="onClick"
            @click:append="onClick"
          >
          </v-text-field>
        </v-col>
      </v-row>
    </v-container>
  </v-form>
</template>

<script lang="ts">
import { ref, inject, defineComponent } from "@nuxtjs/composition-api";
import SchedulerConfigurationStoreKey from "../StoreKey/SchedulerConfigurationStoreKey";

export default defineComponent({
  setup() {
    const schedulerconfigurationstore = inject(SchedulerConfigurationStoreKey);
    if (!schedulerconfigurationstore) {
      throw new Error(`${SchedulerConfigurationStoreKey} is not provided`);
    }
    let version = ref("");
    schedulerconfigurationstore
      .fetchVersion()
      .then((v) => (version.value = v.data));

    const onClick = () => {
      schedulerconfigurationstore.fetchSelected();
    };
    return {
      version,
      onClick,
    };
  },
});
</script>

<style>
.custom-input-color input,
.v-label {
  color: #326ce5 !important;
}
.v-input__icon--append .v-icon {
  color: #326ce5;
}
</style>
