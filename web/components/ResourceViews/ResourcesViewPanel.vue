<template>
  <div>
    <div v-if="IsAlphaTableViewEnabled()">
      <ResourcesDataTable />
    </div>
    <div v-else>
      <ResourcesList />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "@nuxtjs/composition-api";
import ResourcesList from "./Lists/ResourcesList.vue";
import ResourcesDataTable from "./DataTables/ResourcesDataTable.vue";

export default defineComponent({
  components: {
    ResourcesList,
    ResourcesDataTable,
  },
  setup() {
    const IsAlphaTableViewEnabled = () => {
      return process.env.ALPHA_TABLE_VIEWS == "1";
    };

    return {
      IsAlphaTableViewEnabled,
    };
  },
});
</script>
