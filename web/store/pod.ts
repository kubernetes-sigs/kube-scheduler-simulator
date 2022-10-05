import { reactive, inject } from "@nuxtjs/composition-api";
import { V1Pod } from "@kubernetes/client-node";
import { PodAPIKey } from "~/api/APIProviderKeys";
import { WatchEventType } from "@/types/resources";

type stateType = {
  selectedPod: SelectedPod | null;
  pods: StatePods;
  lastResourceVersion: string;
};

type StatePods = {
  // key is node name or "unscheduled"
  [key: string]: Array<V1Pod>;
};
type StatePodsKeyIndexTuple = { key: string; index: number };

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
    pods: {},
    lastResourceVersion: "",
  });

  const podAPI = inject(PodAPIKey);
  if (!podAPI) {
    throw new Error(`${podAPI} is not provided`);
  }

  function createPodState(pods: V1Pod[]): void {
    pods.forEach((p) => {
      addPodToState(p);
    });
  }
  function addPodToState(p: V1Pod): void {
    if (!p.spec?.nodeName) {
      // unscheduled pod
      if (!state.pods["unscheduled"]) {
        state.pods = Object.assign({}, state.pods, { unscheduled: [p] });
      } else {
        state.pods["unscheduled"].push(p);
      }
    } else if (!state.pods[p.spec?.nodeName as string]) {
      // first pod on the node
      state.pods = Object.assign({}, state.pods, { [p.spec?.nodeName]: [p] });
    } else {
      state.pods[p.spec?.nodeName as string].push(p);
    }
  }
  function modifyPodInState(p: V1Pod): void {
    const targetInfo = findStatePodsKeyByUID(state.pods, p.metadata?.uid!);
    // the pod doesn't exist in the state
    if (targetInfo.index === -1) {
      console.warn("pod doesn't exist in the state");
      addPodToState(p);
    }
    state.pods[targetInfo.key].splice(targetInfo.index, 1);
    addPodToState(p);
  }
  function deletePodInState(p: V1Pod): void {
    const targetInfo = findStatePodsKeyByUID(state.pods, p.metadata?.uid!);
    // the pod doesn't exist in the state
    if (targetInfo.index === -1) {
      console.warn("pod doesn't exist in the state");
    }
    state.pods[targetInfo.key].splice(targetInfo.index, 1);
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

    async fetchSelected() {
      if (this.selected?.item.metadata?.name && this.selected?.item.metadata?.namespace && !this.selected?.isNew) {
        const p = await podAPI.getPod(this.selected.item.metadata.namespace, this.selected.item.metadata.name);
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
    },

    async delete(p: V1Pod) {
      if (p.metadata?.name && p.metadata.namespace) {
        await podAPI.deletePod(p.metadata.namespace, p.metadata.name);
      } else {
        throw new Error(
          "failed to delete pod: pod should have metadata.name"
        );
      }
    },

    // initList calls list API, and stores current resource data and lastResourceVersion.
    async initList() {
      const listpods = await podAPI.listPod();
      createPodState(listpods.items);
      state.lastResourceVersion = listpods.metadata?.resourceVersion!;
    },

    // watchEventHandler handles each notified event.
    async watchEventHandler(eventType: WatchEventType, pod: V1Pod) {
      switch (eventType) {
        case WatchEventType.ADDED: {
          addPodToState(pod);
          break;
        }
        case WatchEventType.MODIFIED: {
          modifyPodInState(pod);
          break;
        }
        case WatchEventType.DELETED: {
          deletePodInState(pod);
          break;
        }
        default:
          break;
      }
    },

    get lastResourceVersion() {
      return state.lastResourceVersion;
    },

    async setLastResourceVersion(pod: V1Pod) {
      state.lastResourceVersion =
        pod.metadata!.resourceVersion || state.lastResourceVersion;
    },
  };

  // findStatePodsKeyByUID searches the pods in the state by uid and returns the key and index of the pods.
  function findStatePodsKeyByUID(
    statePods: StatePods,
    uid: string
  ): StatePodsKeyIndexTuple {
    for (const k of Object.keys(statePods)) {
      const i = statePods[k].findIndex((pod) => pod.metadata?.uid === uid);
      // found the pod
      if (i !== -1) {
        return { key: k, index: i } as StatePodsKeyIndexTuple;
      }
    }
    // not found.
    return { key: "", index: -1 } as StatePodsKeyIndexTuple;
  }
}

export type PodStore = ReturnType<typeof podStore>;
