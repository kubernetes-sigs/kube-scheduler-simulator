<template>
  <v-dialog v-model="dialog" width="500">
    <template v-slot:activator="{ on, attrs }">
      <v-btn class="ma-5 mb-0" color="error" v-bind="attrs" v-on="on">
        Delete
      </v-btn>
    </template>

    <v-card>
      <v-card-title class="2">
        Are you sure to delete the resource?
      </v-card-title>
      <v-divider></v-divider>
      <v-divider></v-divider>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="green darken-1" text @click="deleteOnClick"> Yes </v-btn>
        <v-btn color="green darken-1" text @click="dialog = false"> No </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script lang="ts">
import { ref, defineComponent, PropType } from '@nuxtjs/composition-api'
export default defineComponent({
  props: {
    deleteOnClick: Function,
  },
  setup(props, context) {
    const dialog = ref(false)

    return {
      dialog,
    }
  },
})
</script>
