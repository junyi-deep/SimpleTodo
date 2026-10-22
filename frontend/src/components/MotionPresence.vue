<script setup lang="ts">
import { reveal, leave, cancelMotion } from "../composables/motion";
const props = withDefaults(
  defineProps<{
    kind?: "pane" | "popover" | "drawer" | "row" | "toast" | "dialog";
  }>(),
  { kind: "popover" },
);
</script>
<template>
  <Transition
    :css="false"
    @enter="(el, done) => reveal(el, props.kind, done)"
    @leave="(el, done) => leave(el, props.kind, done)"
    @enter-cancelled="cancelMotion"
    @leave-cancelled="cancelMotion"
    ><slot
  /></Transition>
</template>
