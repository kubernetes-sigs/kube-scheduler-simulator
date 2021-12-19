<template>
  <v-app :style="{ background: $vuetify.theme.themes.light.background }">
    <v-app-bar color="primary" app>
      <img
        src="/kubernetes-logo_with_border.svg"
        height="40"
        alt="p.metadata.name"
        class="mx-2"
      />
      <v-toolbar-title
        class="white--text"
        :style="{ cursor: 'pointer' }"
        v-text="title"
      />
    </v-app-bar>
    <v-main>
      <v-container>
        <nuxt />
      </v-container>
    </v-main>
    <v-footer app>
      <span>&copy; {{ new Date().getFullYear() }}</span>
    </v-footer>
  </v-app>
</template>

<script>
import { defineComponent } from "@nuxtjs/composition-api";

export default defineComponent({
  setup() {},
  data() {
    return {
      title: "Kubernetes scheduler simulator",
    };
  },
});
</script>
