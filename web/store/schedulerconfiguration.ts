import { reactive } from "@nuxtjs/composition-api";
import {
  applySchedulerConfiguration,
  getSchedulerConfiguration,
} from "~/api/v1/schedulerconfiguration";
import { SchedulerConfiguration } from "~/api/v1/types";
import { NuxtAxiosInstance } from "@nuxtjs/axios";

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

export default function schedulerconfigurationStore(
  instance: NuxtAxiosInstance
) {
  const state: stateType = reactive({
    selectedConfig: null,
    schedulerconfigurations: [],
  });

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
      const c = await getSchedulerConfiguration(instance);
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
      await applySchedulerConfiguration(instance, cfg);
    },

    async delete(_: string) {
      // This function do nothing, but exist to satisfy interface on ResourceBar.vue.
    },
  };
}

export type SchedulerConfigurationStore = ReturnType<
  typeof schedulerconfigurationStore
>;
