import { reactive, inject } from "@nuxtjs/composition-api";
import { SchedulerConfiguration } from "~/api/v1/types";
import { SchedulerconfigurationAPIKey } from "~/api/APIProviderKeys";

type stateType = {
  selectedConfig: selectedConfig | null;
};

type selectedConfig = {
  // isNew represents whether this Config is a new one or not.
  isNew: boolean;
  item: SchedulerConfiguration;
  resourceKind: string;
  isDeletable: boolean;
};

export default function schedulerconfigurationStore() {
  const state: stateType = reactive({
    selectedConfig: null,
    schedulerconfigurations: [],
  });

  const schedconfAPI = inject(SchedulerconfigurationAPIKey);
  if (!schedconfAPI) {
    throw new Error(`${schedconfAPI} is not provided`);
  }

  return {
    get selected() {
      return state.selectedConfig;
    },

    resetSelected() {
      state.selectedConfig = null;
    },

    select() {
      this.fetchSelected();
    },

    async fetchSelected() {
      const c = await schedconfAPI.getSchedulerConfiguration();
      if (c) {
        state.selectedConfig = {
          isNew: true,
          item: c,
          resourceKind: "SchedulerConfiguration",
          isDeletable: true,
        };
      }
    },

    async apply(cfg: SchedulerConfiguration) {
      await schedconfAPI.applySchedulerConfiguration(cfg);
    },

    async delete(_: string) {
      // This function do nothing, but exist to satisfy interface on ResourceBar.vue.
    },
  };
}

export type SchedulerConfigurationStore = ReturnType<
  typeof schedulerconfigurationStore
>;
