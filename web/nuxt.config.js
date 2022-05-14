import fs from "fs";

const getTemplateEnv = (hostEnv) => {
  return hostEnv !== "production"
    ? {
        // Build Configuration: https://go.nuxtjs.dev/config-build
        POD_TEMPLATE: fs.readFileSync(
          "./components/lib/templates/pod.yaml",
          "utf8"
        ),
        NODE_TEMPLATE: fs.readFileSync(
          "./components/lib/templates/node.yaml",
          "utf8"
        ),
        PV_TEMPLATE: fs.readFileSync(
          "./components/lib/templates/pv.yaml",
          "utf8"
        ),
        PVC_TEMPLATE: fs.readFileSync(
          "./components/lib/templates/pvc.yaml",
          "utf8"
        ),
        SC_TEMPLATE: fs.readFileSync(
          "./components/lib/templates/storageclass.yaml",
          "utf8"
        ),
        PC_TEMPLATE: fs.readFileSync(
          "./components/lib/templates/priorityclass.yaml",
          "utf8"
        ),
      }
    : true;
};

export default {
  // Global page headers: https://go.nuxtjs.dev/config-head
  head: {
    titleTemplate: "%s - scheduler-simulator",
    title: "scheduler-simulator",
    htmlAttrs: {
      lang: "en",
    },
    meta: [
      { charset: "utf-8" },
      { name: "viewport", content: "width=device-width, initial-scale=1" },
      { hid: "description", name: "description", content: "" },
    ],
    link: [{ rel: "icon", type: "image/x-icon", href: "/favicon.ico" }],
  },

  // Global CSS: https://go.nuxtjs.dev/config-css
  css: [],

  // Plugins to run before rendering page: https://go.nuxtjs.dev/config-plugins
  plugins: [
    // ApiRuntimeConfigPlugin is a custom plugin that is used to create and inject axios instance for API using the value of publicRuntimeConfig.
    { src: "~/plugins/ApiRuntimeConfigPlugin.ts" },
  ],

  // Auto import components: https://go.nuxtjs.dev/config-components
  components: true,

  // Modules for dev and build (recommended): https://go.nuxtjs.dev/config-modules
  buildModules: [
    // https://go.nuxtjs.dev/typescript
    "@nuxt/typescript-build",
    // https://go.nuxtjs.dev/vuetify
    "@nuxtjs/vuetify",
    // for nuxtjs/composition-api
    "@nuxtjs/composition-api/module",
  ],

  // Modules: https://go.nuxtjs.dev/config-modules
  modules: [
    // https://go.nuxtjs.dev/axios
    "@nuxtjs/axios",
  ],

  // Axios module configuration: https://go.nuxtjs.dev/config-axios
  axios: {},

  // Vuetify module configuration: https://go.nuxtjs.dev/config-vuetify
  vuetify: {
    customVariables: ["~/assets/variables.scss"],
    theme: {
      themes: {
        light: {
          primary: "#326ce5",
          background: "#f5f5f5",
        },
      },
    },
  },
  // publicRuntimeConfig should hold all env variables that are public as these will be exposed on the frontend.
  publicRuntimeConfig: {
    baseURL: process.env.BASE_URL || "http://localhost:1212",
    kubeApiServerURL:
      process.env.KUBE_API_SERVER_URL || "http://localhost:3131",
    // alphaTableViews is a optional parameter for the datatable view. This is an alpha feature.
    // If this value is set to "1", the datatable view will be enabled.
    alphaTableViews: process.env.ALPHA_TABLE_VIEWS || "0",
  },

  env: getTemplateEnv(process.env.HOST_ENV),
};
