<template>
  <div>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide } from '@nuxtjs/composition-api'
import storageClassStore from '../../store/storageclass'
import StorageClassStoreKey from '../StoreKey/StorageClassStoreKey'

export default defineComponent({
  setup() {
    provide(StorageClassStoreKey, storageClassStore())
    return {}
  },
})
</script>
