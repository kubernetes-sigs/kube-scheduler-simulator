<template>
  <v-row no-gutters v-if="storageclasses.length !== 0">
    <v-col>
      <v-card class="ma-2" outlined>
        <v-card-title class="mb-1"> StorageClasses </v-card-title>
        <v-card-actions>
          <v-chip
            class="ma-2"
            v-for="(p, i) in storageclasses"
            :key="i"
            @click.stop="onClick(p)"
            color="primary"
            outlined
            large
            label
          >
            <img src="/sc.svg" height="40" alt="p.metadata.name" class="mr-2" />
            {{ p.metadata.name }}
          </v-chip>
        </v-card-actions>
      </v-card>
    </v-col>
  </v-row>
</template>

<script lang="ts">
import { V1StorageClass } from '@kubernetes/client-node'
import {
  ref,
  computed,
  inject,
  onMounted,
  PropType,
  defineComponent,
} from '@nuxtjs/composition-api'
import {} from './lib/util'
import StorageClassStoreKey from './StoreKey/StorageClassStoreKey'
export default defineComponent({
  setup(_, context) {
    const store = inject(StorageClassStoreKey)
    if (!store) {
      throw new Error(`${StorageClassStoreKey} is not provided`)
    }

    const getStorageClassList = async () => {
      const route = context.root.$route
      await store.fetchlist()
    }
    const onClick = (storageclass: V1StorageClass) => {
      store.select(storageclass, false)
    }
    onMounted(getStorageClassList)
    const storageclasses = computed(() => store.storageclasses)
    return {
      storageclasses,
      onClick,
    }
  },
})
</script>
