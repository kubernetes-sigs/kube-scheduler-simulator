<template v-slot:prepend>
  <v-list-item two-line>
    <v-list-item-content>
      <v-list-item-title> {{ title }} </v-list-item-title>
      <v-row>
        <v-col>
          <v-switch
            class="ma-5 mb-0"
            @change="editmodeOnChange"
            v-if="enableEditmodeSwitch"
            inset
            label="edit"
          />
        </v-col>
        <v-spacer v-for="n in 3" :key="n" />
        <v-col>
          <v-btn class="ma-5 mb-0" @click="applyOnClick"> Apply </v-btn>
        </v-col>
        <v-col>
          <ResourceDeleteButton
            :deleteOnClick="deleteOnClick"
            v-if="enableDeleteBtn"
          />
        </v-col>
      </v-row>
    </v-list-item-content>
  </v-list-item>
</template>

<script lang="ts">
import { ref, defineComponent, PropType } from '@nuxtjs/composition-api'
import ResourceDeleteButton from './DeleteButton.vue'

export default defineComponent({
  props: {
    title: String,
    deleteOnClick: Function,
    applyOnClick: Function,
    editmodeOnChange: Function,
    enableDeleteBtn: Boolean,
    enableEditmodeSwitch: Boolean,
  },
  components: {
    ResourceDeleteButton,
  },
  setup(props, context) {
    const dialog = ref(false)
    return {
      dialog,
    }
  },
})
</script>
