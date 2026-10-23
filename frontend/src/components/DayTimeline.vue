<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from "vue";
import { ChevronLeft, ChevronRight } from "lucide-vue-next";
import { runMotion, cancelMotion } from "../composables/motion";
import { parseDay, shiftDay } from "../api/types";
const props = defineProps<{ date: string; today?: string }>();
const emit = defineEmits<{
  select: [string];
  step: [number];
  today: [];
  drop: [DragEvent, string];
}>();
const track = ref<HTMLElement>(),
  pill = ref<HTMLElement>();
const days = computed(() => {
  const n = (parseDay(props.date).getDay() + 6) % 7;
  return Array.from({ length: 7 }, (_, i) => shiftDay(props.date, i - n));
});
let previousX: number | undefined, resize: ResizeObserver;
async function position() {
  await nextTick();
  const el = track.value?.querySelector<HTMLElement>(".selected");
  if (!el || !pill.value || !track.value) return;
  const r = el.getBoundingClientRect(),
    host = track.value.getBoundingClientRect(),
    x = r.left - host.left;
  const current = pill.value.getBoundingClientRect().left - host.left;
  cancelMotion(pill.value);
  Object.assign(pill.value.style, {
    width: r.width + "px",
    height: r.height + "px",
    top: r.top - host.top + "px",
    transform: `translateX(${x}px)`,
    opacity: "1",
  });
  if (previousX !== undefined)
    runMotion(
      pill.value,
      { transform: [`translateX(${current}px)`, `translateX(${x}px)`] },
      0.28,
    );
  previousX = x;
}
watch(() => props.date, position);
onMounted(() => {
  position();
  resize = new ResizeObserver(position);
  if (track.value) resize.observe(track.value);
});
onUnmounted(() => {
  resize?.disconnect();
  if (pill.value) cancelMotion(pill.value);
});
</script>
<template>
  <div ref="track" class="compact-week smooth-timeline">
    <span ref="pill" class="timeline-pill" aria-hidden="true" /><button
      class="icon-button"
      aria-label="前一天"
      @click="emit('step', -1)"
    >
      <ChevronLeft :size="14" /></button
    ><button
      v-for="(day, i) in days"
      :key="day"
      class="week-date"
      :class="{ selected: date === day, today: today === day }"
      :aria-label="
        '周' +
        ['一', '二', '三', '四', '五', '六', '日'][i] +
        ' ' +
        Number(day.slice(-2))
      "
      :aria-current="date === day ? 'date' : undefined"
      @click="emit('select', day)"
      @dragover.prevent
      @drop="emit('drop', $event, day)"
    >
      <small>{{ ["一", "二", "三", "四", "五", "六", "日"][i] }}</small
      >{{ Number(day.slice(-2)) }}</button
    ><button class="icon-button" aria-label="后一天" @click="emit('step', 1)">
      <ChevronRight :size="14" /></button
    ><button class="text-button" @click="emit('today')">今天</button>
  </div>
</template>
