import { createApp } from "vue";
import { createPinia } from "pinia";
import App from "./App.vue";
import "./style.css";
import { vMotion, vComplete } from "./composables/motion";
createApp(App)
  .use(createPinia())
  .directive("motion", vMotion)
  .directive("complete", vComplete)
  .mount("#app");
