<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide } from '@nuxtjs/composition-api'
import podStore from '../../store/pod'
import PodStoreKey from '../StoreKey/PodStoreKey'

export default defineComponent({
  setup() {
    provide(PodStoreKey, podStore())
    return {}
  },
})
</script>
