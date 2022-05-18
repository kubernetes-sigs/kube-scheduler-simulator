import { reactive, inject } from "@nuxtjs/composition-api";
import { V1Node } from "@kubernetes/client-node";
import { NodeAPIKey } from "~/api/APIProviderKeys";

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
  const nodeAPI = inject(NodeAPIKey);
  if (!nodeAPI) {
    throw new Error(`${nodeAPI} is not provided`);
  }
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

    async fetchlist() {
      const nodes = await nodeAPI.listNode();
      state.nodes = nodes.items;
    },

    async fetchSelected() {
      if (state.selectedNode?.item.metadata?.name && !this.selected?.isNew) {
        const n = await nodeAPI.getNode(state.selectedNode.item.metadata.name);
        this.select(n, false);
      }
    },

    async apply(n: V1Node) {
      if (n.metadata?.name) {
        await nodeAPI.applyNode(n);
      } else if (!n.metadata?.name && n.metadata?.generateName) {
        await nodeAPI.createNode(n);
      } else {
        throw new Error(`
        failed to apply node: node has no metadata.name or metadata.generateName
        `);
      }
      await this.fetchlist();
    },

    async delete(name: string) {
      await nodeAPI.deleteNode(name);
      await this.fetchlist();
    },
  };
}

export type NodeStore = ReturnType<typeof nodeStore>;
