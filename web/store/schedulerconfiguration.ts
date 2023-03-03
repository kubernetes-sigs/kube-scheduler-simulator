import { reactive, inject } from "@nuxtjs/composition-api";
import { SchedulerConfiguration } from "~/api/v1/types";
import { SchedulerconfigurationAPIKey } from "~/api/APIProviderKeys";

type stateType = {
  // when users use an external scheduler, we disable it.
  disabled: boolean;
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
    disabled: false,
    schedulerconfigurations: [],
  });

  const schedconfAPI = inject(SchedulerconfigurationAPIKey);
  if (!schedconfAPI) {
    throw new Error(`${SchedulerconfigurationAPIKey.description} is not provided`);
  }

  return {
    get disabled() {
      return state.disabled;
    },

    get selected() {
      return state.selectedConfig;
    },

    resetSelected() {
      state.selectedConfig = null;
    },

    select() {
      this.fetchSelected();
    },

    async initialize() {
      await schedconfAPI.getSchedulerConfiguration().catch((e) => {
        if (e.response.status == 400) {
          // users use an external scheduler on backend.
          state.disabled = true;
          return;
        }
        throw new Error(`failed to apply scheduler configration: ${e}`);
      });
    },

    async fetchSelected() {
      const c = await schedconfAPI.getSchedulerConfiguration().catch((e) => {
        if (e.response.status == 400) {
          // users use an external scheduler on backend.
          state.disabled = true;
          return;
        }
        throw new Error(`failed to apply scheduler configration: ${e}`);
      });
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
