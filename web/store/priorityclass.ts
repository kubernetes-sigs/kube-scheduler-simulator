import { reactive, inject } from "@nuxtjs/composition-api";
import { V1PriorityClass } from "@kubernetes/client-node";
import { PriorityClassAPIKey } from "~/api/APIProviderKeys";
import {
  createResourceState,
  addResourceToState,
  modifyResourceInState,
  deleteResourceInState,
} from "./helpers/storeHelper";
import { WatchEventType } from "@/types/resources";

type stateType = {
  selectedPriorityClass: selectedPriorityClass | null;
  priorityclasses: V1PriorityClass[];
  lastResourceVersion: string;
};

type selectedPriorityClass = {
  // isNew represents whether this is a new PriorityClass or not.
  isNew: boolean;
  item: V1PriorityClass;
  resourceKind: string;
  isDeletable: boolean;
};

export default function priorityclassStore() {
  const state: stateType = reactive({
    selectedPriorityClass: null,
    priorityclasses: [],
    lastResourceVersion: "",
  });

  const priorityClassAPI = inject(PriorityClassAPIKey);
  if (!priorityClassAPI) {
    throw new Error(`${PriorityClassAPIKey.description} is not provided`);
  }

  // `CheckIsDeletable` returns whether the given PriorityClass can be deleted or not.
  // The PriorityClasses that have the name prefixed with `system-` are reserved by the system so can't be deleted.
  const checkIsDeletable = (n: V1PriorityClass) => {
    return !!n.metadata?.name && !n.metadata?.name?.startsWith("system-");
  };

  return {
    get priorityclasses() {
      return state.priorityclasses;
    },

    get count(): number {
      return state.priorityclasses.length;
    },

    get selected() {
      return state.selectedPriorityClass;
    },

    select(n: V1PriorityClass | null, isNew: boolean) {
      if (n !== null) {
        state.selectedPriorityClass = {
          isNew: isNew,
          item: n,
          resourceKind: "PC",
          isDeletable: checkIsDeletable(n),
        };
      }
    },

    resetSelected() {
      state.selectedPriorityClass = null;
    },

    async apply(n: V1PriorityClass) {
      if (n.metadata?.name) {
        await priorityClassAPI.applyPriorityClass(n);
      } else if (n.metadata?.generateName) {
        // This PriorityClass can be expected to be a newly created PriorityClass. So, use `createPriorityClass` instead.
        await priorityClassAPI.createPriorityClass(n);
      } else {
        throw new Error(
          "failed to apply priorityclass: priorityclass should have metadata.name or metadata.generateName"
        );
      }
    },

    async fetchSelected() {
      if (
        state.selectedPriorityClass?.item.metadata?.name &&
        !this.selected?.isNew
      ) {
        const s = await priorityClassAPI.getPriorityClass(
          state.selectedPriorityClass.item.metadata.name
        );
        this.select(s, false);
      }
    },

    async delete(name: string) {
      await priorityClassAPI.deletePriorityClass(name);
    },

    // initList calls list API, and stores current resource data and lastResourceVersion.
    async initList() {
      const listpriorityclasses = await priorityClassAPI.listPriorityClass();
      state.priorityclasses = createResourceState<V1PriorityClass>(
        listpriorityclasses.items
      );
      state.lastResourceVersion =
        listpriorityclasses.metadata?.resourceVersion!;
    },

    // watchEventHandler handles each notified event.
    async watchEventHandler(eventType: WatchEventType, pc: V1PriorityClass) {
      switch (eventType) {
        case WatchEventType.ADDED: {
          state.priorityclasses = addResourceToState(state.priorityclasses, pc);
          break;
        }
        case WatchEventType.MODIFIED: {
          state.priorityclasses = modifyResourceInState(
            state.priorityclasses,
            pc
          );
          break;
        }
        case WatchEventType.DELETED: {
          state.priorityclasses = deleteResourceInState(
            state.priorityclasses,
            pc
          );
          break;
        }
        default:
          break;
      }
    },

    get lastResourceVersion() {
      return state.lastResourceVersion;
    },

    async setLastResourceVersion(pc: V1PriorityClass) {
      state.lastResourceVersion =
        pc.metadata!.resourceVersion || state.lastResourceVersion;
    },
  };
}

export type PriorityClassStore = ReturnType<typeof priorityclassStore>;
