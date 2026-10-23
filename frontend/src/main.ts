import { createApp } from "vue";
import { createPinia } from "pinia";
import App from "./App.vue";
import "./style.css";
import { vMotion, vComplete } from "./composables/motion";
import { vSparks } from "./composables/typingStars";
createApp(App)
  .use(createPinia())
  .directive("sparks", vSparks)
  .directive("motion", vMotion)
  .directive("complete", vComplete)
  .mount("#app");
