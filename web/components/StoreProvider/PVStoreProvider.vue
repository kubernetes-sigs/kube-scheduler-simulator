<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide } from '@nuxtjs/composition-api'
import PersistentVolumeStore from '../../store/pv'
import PersistentVolumeStoreKey from '../StoreKey/PVStoreKey'

export default defineComponent({
  setup() {
    provide(PersistentVolumeStoreKey, PersistentVolumeStore())
    return {}
  },
})
</script>
