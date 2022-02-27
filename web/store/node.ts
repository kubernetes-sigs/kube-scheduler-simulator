import { reactive } from "@nuxtjs/composition-api";
import { applyNode, deleteNode, listNode, getNode } from "~/api/v1/node";
import { V1Node } from "@kubernetes/client-node";

type stateType = {
  selectedNode: selectedNode | null;
  nodes: V1Node[];
};

type selectedNode = {
  // isNew represents whether this Node is a new one or not.
  isNew: boolean;
  item: V1Node;
  resourceKind: string;
  isDeletable: boolean;
};

export default function nodeStore() {
  const state: stateType = reactive({
    selectedNode: null,
    nodes: [],
  });

  return {
    get nodes() {
      return state.nodes;
    },

    get count(): number {
      return state.nodes.length;
    },

    get selected() {
      return state.selectedNode;
    },

    select(n: V1Node | null, isNew: boolean) {
      if (n !== null) {
        state.selectedNode = {
          isNew: isNew,
          item: n,
          resourceKind: "Node",
          isDeletable: true,
        };
      }
    },

    resetSelected() {
      state.selectedNode = null;
    },

    async fetchlist(onError: (_: string) => void) {
      const nodes = await listNode(onError);
      if (!nodes) return;
      state.nodes = nodes.items;
    },

    async fetchSelected(onError: (_: string) => void) {
      if (state.selectedNode?.item.metadata?.name && !this.selected?.isNew) {
        const n = await getNode(state.selectedNode.item.metadata.name, onError);
        if (!n) return;
        this.select(n, false);
      }
    },

    async apply(
      n: V1Node,

      onError: (_: string) => void
    ) {
      await applyNode(n, onError);
      await this.fetchlist(onError);
    },

    async delete(name: string, onError: (_: string) => void) {
      await deleteNode(name, onError);
      await this.fetchlist(onError);
    },
  };
}

export type NodeStore = ReturnType<typeof nodeStore>;
