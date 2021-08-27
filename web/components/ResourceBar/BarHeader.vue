<template v-slot:prepend>
  <v-list-item two-line>
    <v-list-item-content>
      <v-list-item-title> {{ title }} </v-list-item-title>
      <v-row>
        <v-col>
          <v-switch
            v-if="enableEditmodeSwitch"
            class="ma-5 mb-0"
            inset
            label="edit"
            @change="editmodeOnChange"
          />
        </v-col>
        <v-spacer v-for="n in 3" :key="n" />
        <v-col>
          <v-btn class="ma-5 mb-0" @click="applyOnClick"> Apply </v-btn>
        </v-col>
        <v-col>
          <ResourceDeleteButton
            v-if="enableDeleteBtn"
            :delete-on-click="deleteOnClick"
          />
        </v-col>
      </v-row>
    </v-list-item-content>
  </v-list-item>
</template>

<script lang="ts">
import { ref, defineComponent } from "@nuxtjs/composition-api";
import ResourceDeleteButton from "./DeleteButton.vue";

export default defineComponent({
  components: {
    ResourceDeleteButton,
  },
  props: {
    title: {
      type: String,
      default: "",
    },
    deleteOnClick: {
      type: Function,
      default: null,
    },
    applyOnClick: {
      type: Function,
      default: null,
    },
    editmodeOnChange: {
      type: Function,
      default: null,
    },
    enableDeleteBtn: {
      type: Boolean,
      default: false,
    },
    enableEditmodeSwitch: {
      type: Boolean,
      default: false,
    },
  },
  setup() {
    const dialog = ref(false);
    return {
      dialog,
    };
  },
});
</script>
