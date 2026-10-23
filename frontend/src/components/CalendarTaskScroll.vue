<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
const props = defineProps<{ label: string }>();
const viewport = ref<HTMLElement>(),
  content = ref<HTMLElement>(),
  rail = ref<HTMLElement>();
const height = ref(0),
  total = ref(0),
  top = ref(0);
const maximum = computed(() => Math.max(0, total.value - height.value));
const thumbHeight = computed(() =>
  Math.min(
    height.value,
    Math.max(12, (height.value * height.value) / (total.value || 1)),
  ),
);
const thumbTop = computed(() =>
  maximum.value
    ? (top.value / maximum.value) * (height.value - thumbHeight.value)
    : 0,
);
const id = "calendar-scroll-" + Math.random().toString(36).slice(2);
let observer: ResizeObserver, drag: { y: number; scroll: number } | undefined;
function measure() {
  if (!viewport.value) return;
  height.value = viewport.value.clientHeight;
  total.value = viewport.value.scrollHeight;
  top.value = viewport.value.scrollTop;
}
function scrollTo(value: number) {
  if (viewport.value) {
    viewport.value.scrollTop = Math.max(0, Math.min(maximum.value, value));
    measure();
  }
}
function pointerDown(event: PointerEvent) {
  if (event.button !== 0 || !rail.value) return;
  const y = event.clientY - rail.value.getBoundingClientRect().top;
  if (y < thumbTop.value || y > thumbTop.value + thumbHeight.value)
    scrollTo(top.value + (y < thumbTop.value ? -1 : 1) * height.value * 0.9);
  drag = { y: event.clientY, scroll: top.value };
  rail.value.setPointerCapture(event.pointerId);
}
function pointerMove(event: PointerEvent) {
  if (drag)
    scrollTo(
      drag.scroll +
        ((event.clientY - drag.y) * maximum.value) /
          Math.max(1, height.value - thumbHeight.value),
    );
}
function pointerUp(event: PointerEvent) {
  drag = undefined;
  if (rail.value?.hasPointerCapture(event.pointerId))
    rail.value.releasePointerCapture(event.pointerId);
}
function key(event: KeyboardEvent) {
  const steps: Record<string, number> = {
    ArrowDown: 25,
    ArrowUp: -25,
    PageDown: height.value * 0.9,
    PageUp: -height.value * 0.9,
    Home: -maximum.value,
    End: maximum.value,
  };
  if (event.key in steps) {
    event.preventDefault();
    event.stopPropagation();
    scrollTo(top.value + steps[event.key]);
  }
}
onMounted(() => {
  observer = new ResizeObserver(measure);
  if (viewport.value) observer.observe(viewport.value);
  if (content.value) observer.observe(content.value);
  measure();
});
onUnmounted(() => observer?.disconnect());
</script>
<template>
  <div class="calendar-task-viewport" :class="{ 'has-overflow': maximum > 1 }">
    <div
      ref="viewport"
      :id="id"
      class="calendar-task-list"
      :aria-label="label"
      tabindex="0"
      @scroll="measure"
    >
      <div ref="content" class="calendar-task-content"><slot /></div>
    </div>
    <div
      v-if="maximum > 1"
      ref="rail"
      class="calendar-scrollbar"
      role="scrollbar"
      tabindex="0"
      :aria-label="label + ' 滚动条'"
      aria-orientation="vertical"
      :aria-controls="id"
      :aria-valuemin="0"
      :aria-valuemax="maximum"
      :aria-valuenow="Math.round(top)"
      @pointerdown.stop.prevent="pointerDown"
      @pointermove="pointerMove"
      @pointerup="pointerUp"
      @pointercancel="pointerUp"
      @lostpointercapture="drag = undefined"
      @click.stop
      @keydown="key"
      @wheel.stop.prevent="scrollTo(top + $event.deltaY)"
    >
      <div
        class="calendar-scroll-thumb"
        :style="{
          height: thumbHeight + 'px',
          transform: `translateY(${thumbTop}px)`,
        }"
      />
    </div>
  </div>
</template>
