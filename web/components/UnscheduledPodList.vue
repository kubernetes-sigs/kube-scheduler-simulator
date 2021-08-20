<template>
  <v-card class="ma-2" outlined v-if="pods['unscheduled'].length !== 0">
    <v-card-title class="mb-1"> Unscheduled Pods </v-card-title>
    <PodList nodeName="unscheduled" />
  </v-card>
</template>

<script lang="ts">
import {
  computed,
  inject,
  onMounted,
  defineComponent,
} from '@nuxtjs/composition-api'
import {} from './lib/util'
import PodStoreKey from './StoreKey/PodStoreKey'
export default defineComponent({
  setup(_, context) {
    const store = inject(PodStoreKey)
    if (!store) {
      throw new Error(`${PodStoreKey} is not provided`)
    }

    const getPodList = async () => {
      await store.fetchlist()
    }
    onMounted(getPodList)

    const pods = computed(() => store.pods)
    return {
      pods,
    }
  },
})
</script>
