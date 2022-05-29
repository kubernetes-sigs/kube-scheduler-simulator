import { reactive, inject } from "@nuxtjs/composition-api";
import { V1Pod } from "@kubernetes/client-node";
import { PodAPIKey } from "~/api/APIProviderKeys";

type stateType = {
  selectedPod: SelectedPod | null;
  pods: {
    // key is node name or "unscheduled"
    [key: string]: Array<V1Pod>;
  };
};

export type SelectedPod = {
  // isNew represents whether this is a new one or not.
  isNew: boolean;
  item: V1Pod;
  resourceKind: string;
  isDeletable: boolean;
};

export default function podStore() {
  const state: stateType = reactive({
    selectedPod: null,
    pods: { unscheduled: [] },
  });

  const podAPI = inject(PodAPIKey);
  if (!podAPI) {
    throw new Error(`${podAPI} is not provided`);
  }

  return {
    get pods() {
      return state.pods;
    },

    get count(): number {
      let num = 0;
      Object.keys(state.pods).forEach((key) => {
        num += state.pods[key].length;
      });
      return num;
    },

    get selected() {
      return state.selectedPod;
    },

    select(p: V1Pod | null, isNew: boolean) {
      if (p !== null) {
        state.selectedPod = {
          isNew: isNew,
          item: p,
          resourceKind: "Pod",
          isDeletable: true,
        };
      }
    },

    resetSelected() {
      state.selectedPod = null;
    },

    async fetchlist() {
      const listpods = await podAPI.listPod();
      const pods = listpods.items;
      const result: { [key: string]: Array<V1Pod> } = {};
      result["unscheduled"] = [];
      pods.forEach((p) => {
        if (!p.spec?.nodeName) {
          // unscheduled pod
          result["unscheduled"].push(p);
        } else if (!result[p.spec?.nodeName as string]) {
          // first pod on the node
          result[p.spec?.nodeName as string] = [p];
        } else {
          result[p.spec?.nodeName as string].push(p);
        }
      });
      state.pods = result;
    },

    async fetchSelected() {
      if (this.selected?.item.metadata?.name && !this.selected?.isNew) {
        const p = await podAPI.getPod(this.selected.item.metadata.name);
        this.select(p, false);
      }
    },

    async apply(p: V1Pod) {
      if (p.metadata?.name) {
        await podAPI.applyPod(p);
      } else if (p.metadata?.generateName) {
        // This Pod can be expected to be a newly created Pod. So, use `createPod` instead.
        await podAPI.createPod(p);
      } else {
        throw new Error(
          "failed to apply pod: pod should have metadata.name or metadata.generateName"
        );
      }
      await this.fetchlist();
    },

    async delete(name: string) {
      await podAPI.deletePod(name);
      await this.fetchlist();
    },
  };
}

export type PodStore = ReturnType<typeof podStore>;
