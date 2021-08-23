<template>
  <v-row no-gutters v-if="pvcs.length !== 0">
    <v-col>
      <v-card class="ma-2" outlined>
        <v-card-title class="mb-1"> PersistentVolumeClaims </v-card-title>
        <v-card-actions>
          <v-chip
            class="ma-2"
            v-for="(p, i) in pvcs"
            :key="i"
            @click.stop="onClick(p)"
            color="primary"
            outlined
            large
            label
          >
            <img
              src="/pvc.svg"
              height="40"
              alt="p.metadata.name"
              class="mr-2"
            />
            {{ p.metadata.name }}
          </v-chip>
        </v-card-actions>
      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { V1PersistentVolumeClaim } from '@kubernetes/client-node'
import {
  ref,
  computed,
  inject,
  onMounted,
  defineComponent,
} from '@nuxtjs/composition-api'
import {} from './lib/util'
import PersistentVolumeClaimStoreKey from './StoreKey/PVCStoreKey'
export default defineComponent({
  setup(_, context) {
    const store = inject(PersistentVolumeClaimStoreKey)
    if (!store) {
      throw new Error(`${PersistentVolumeClaimStoreKey} is not provided`)
    }

    const getPVCList = async () => {
      await store.fetchlist()
    }
    const onClick = (pvc: V1PersistentVolumeClaim) => {
      store.select(pvc, false)
    }
    onMounted(getPVCList)
    const pvcs = computed(() => store.pvcs)
    return {
      pvcs,
      onClick,
    }
  },
})
</script>
