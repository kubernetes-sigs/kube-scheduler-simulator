<template>
  <v-navigation-drawer
    v-model="drawer"
    fixed
    right
    temporary
    bottom
    width="70%"
  >
    <BarHeader
      title="Resource"
      :delete-on-click="deleteOnClick"
      :apply-on-click="applyOnClick"
      :editmode-on-change="
        () => {
          editmode = !editmode;
        }
      "
      :enable-delete-btn="selected && !selected.isNew && selected.isDeletable"
      :enable-editmode-switch="selected && !selected.isNew"
    />

    <v-divider></v-divider>

    <template v-if="editmode">
      <v-spacer v-for="n in 3" :key="n" />
      <v-divider></v-divider>

      <YamlEditor v-model="formData" />
    </template>

    <template v-if="!editmode">
      <SchedulingResults />
      <ResourceDefinitionTree :items="treeData" />
      <!-- This is required to work around the vuetify's bug, refer more details in #10 -->
      <div style="height: 80%"></div>
    </template>
  </v-navigation-drawer>
</template>
<script lang="ts">
import {
  ref,
  computed,
  inject,
  watch,
  defineComponent,
} from "@nuxtjs/composition-api";
import yaml from "js-yaml";
import PodStoreKey from "../StoreKey/PodStoreKey";
import { objectToTreeViewData } from "../lib/util";
import NodeStoreKey from "../StoreKey/NodeStoreKey";
import PersistentVolumeStoreKey from "../StoreKey/PVStoreKey";
import PersistentVolumeClaimStoreKey from "../StoreKey/PVCStoreKey";
import StorageClassStoreKey from "../StoreKey/StorageClassStoreKey";
import PriorityClassStoreKey from "../StoreKey/PriorityClassStoreKey";
import SchedulerConfigurationStoreKey from "../StoreKey/SchedulerConfigurationStoreKey";
import YamlEditor from "./YamlEditor.vue";
import SchedulingResults from "./SchedulingResults.vue";
import ResourceDefinitionTree from "./DefinitionTree.vue";
import BarHeader from "./BarHeader.vue";
import {
  V1Node,
  V1PersistentVolumeClaim,
  V1PersistentVolume,
  V1Pod,
  V1StorageClass,
  V1PriorityClassList,
} from "@kubernetes/client-node";
import SnackBarStoreKey from "../StoreKey/SnackBarStoreKey";
import { SchedulerConfiguration } from "~/api/v1/types";

type Resource =
  | V1Pod
  | V1Node
  | V1PersistentVolumeClaim
  | V1PersistentVolume
  | V1StorageClass
  | V1PriorityClassList
  | SchedulerConfiguration;

interface Store {
  readonly selected: object | null;
  resetSelected(): void;
  apply(_: Resource): Promise<void>;
  delete(_: string): Promise<void>;
  fetchSelected(): Promise<void>;
}

interface SelectedItem {
  isNew: boolean;
  item: Resource;
  resourceKind: string;
  isDeletable: boolean;
}

export default defineComponent({
  components: {
    YamlEditor,
    BarHeader,
    ResourceDefinitionTree,
    SchedulingResults,
  },
  setup() {
    var store: Store | null = null;

    // inject stores
    const podstore = inject(PodStoreKey);
    if (!podstore) {
      throw new Error(`${PodStoreKey} is not provided`);
    }
    const nodestore = inject(NodeStoreKey);
    if (!nodestore) {
      throw new Error(`${NodeStoreKey} is not provided`);
    }
    const pvstore = inject(PersistentVolumeStoreKey);
    if (!pvstore) {
      throw new Error(`${pvstore} is not provided`);
    }
    const pvcstore = inject(PersistentVolumeClaimStoreKey);
    if (!pvcstore) {
      throw new Error(`${pvcstore} is not provided`);
    }
    const storageclassstore = inject(StorageClassStoreKey);
    if (!storageclassstore) {
      throw new Error(`${StorageClassStoreKey} is not provided`);
    }
    const priorityclassstore = inject(PriorityClassStoreKey);
    if (!priorityclassstore) {
      throw new Error(`${PriorityClassStoreKey} is not provided`);
    }
    const schedulerconfigurationstore = inject(SchedulerConfigurationStoreKey);
    if (!schedulerconfigurationstore) {
      throw new Error(`${SchedulerConfigurationStoreKey} is not provided`);
    }

    const snackbarstore = inject(SnackBarStoreKey);
    if (!snackbarstore) {
      throw new Error(`${SnackBarStoreKey} is not provided`);
    }

    const treeData = ref(objectToTreeViewData(null));

    // for edit mode
    const formData = ref("");

    // boolean to switch some view
    const drawer = ref(false);
    const editmode = ref(false);

    // watch each selected resource
    const selected = ref(null as SelectedItem | null);
    const pod = computed(() => podstore.selected);
    watch(pod, () => {
      store = podstore;
      selected.value = pod.value;
    });

    const node = computed(() => nodestore.selected);
    watch(node, () => {
      store = nodestore;
      selected.value = node.value;
    });

    const pv = computed(() => pvstore.selected);
    watch(pv, () => {
      store = pvstore;
      selected.value = pv.value;
    });

    const pvc = computed(() => pvcstore.selected);
    watch(pvc, () => {
      store = pvcstore;
      selected.value = pvc.value;
    });

    const sc = computed(() => storageclassstore.selected);
    watch(sc, () => {
      store = storageclassstore;
      selected.value = sc.value;
    });

    const pc = computed(() => priorityclassstore.selected);
    watch(pc, () => {
      store = priorityclassstore;
      selected.value = pc.value;
    });

    const config = computed(() => schedulerconfigurationstore.selected);
    watch(config, () => {
      store = schedulerconfigurationstore;
      selected.value = config.value;
    });

    watch(selected, (newVal, oldVal) => {
      if (selected.value) {
        if (!oldVal) {
          fetchSelected().then((_) => {
            if (selected.value) {
              editmode.value = selected.value.isNew;

              formData.value = yaml.dump(selected.value.item);
              treeData.value = objectToTreeViewData(selected.value.item);
              drawer.value = true;
            }
          });
        }
      }
    });

    watch(drawer, (newValue, _) => {
      if (!newValue) {
        // reset editmode.
        editmode.value = false;
        if (store) {
          store.resetSelected();
        }
        store = null;
        selected.value = null;
      }
    });

    const fetchSelected = async () => {
      if (store) {
        await store.fetchSelected().catch((e) => setServerErrorMessage(e));
      }
    };

    const setServerErrorMessage = (error: string) => {
      snackbarstore.setServerErrorMessage(error);
    };

    const applyOnClick = () => {
      if (store) {
        const y = yaml.load(formData.value);
        store.apply(y).catch((e) => setServerErrorMessage(e));
      }
      drawer.value = false;
    };

    const deleteOnClick = () => {
      if (selected.value?.resourceKind === "Node") {
        // when the Node is deleted, all Pods on the Node should be deleted as well.
        //@ts-ignore
        podstore.pods[selected.value?.item.metadata?.name].forEach((p) => {
          //@ts-ignore
          if (p.spec?.nodeName === selected.value?.item.metadata?.name) {
            podstore
              //@ts-ignore
              .delete(p.metadata?.name)
              .catch((e) => setServerErrorMessage(e));
          }
        });
      }
      if (selected.value?.resourceKind != "SchedulerConfiguration") {
        //@ts-ignore // Only SchedulerConfiguration don't have the metadata field.
        if (selected.value?.item.metadata?.name && store) {
          store
            .delete(
              //@ts-ignore
              selected.value.item.metadata.name
            )
            .catch((e) => setServerErrorMessage(e));
        }
      }
      drawer.value = false;
    };

    return {
      drawer,
      editmode,
      selected,
      formData,
      treeData,
      applyOnClick,
      deleteOnClick,
    };
  },
});
</script>
