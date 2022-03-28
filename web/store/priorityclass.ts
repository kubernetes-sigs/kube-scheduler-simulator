import { reactive } from "@nuxtjs/composition-api";
import {
  applyPriorityClass,
  deletePriorityClass,
  getPriorityClass,
  listPriorityClass,
} from "~/api/v1/priorityclass";
import { V1PriorityClass } from "@kubernetes/client-node";

type stateType = {
  selectedPriorityClass: selectedPriorityClass | null;
  priorityclasses: V1PriorityClass[];
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
  });

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

    async fetchlist() {
      const priorityclasses = await listPriorityClass();
      state.priorityclasses = priorityclasses.items;
    },

    async apply(n: V1PriorityClass) {
      await applyPriorityClass(n);
      await this.fetchlist();
    },

    async fetchSelected() {
      if (
        state.selectedPriorityClass?.item.metadata?.name &&
        !this.selected?.isNew
      ) {
        const s = await getPriorityClass(
          state.selectedPriorityClass.item.metadata.name
        );
        this.select(s, false);
      }
    },

    async delete(name: string) {
      await deletePriorityClass(name);
      await this.fetchlist();
    },
  };
}

export type PriorityClassStore = ReturnType<typeof priorityclassStore>;
