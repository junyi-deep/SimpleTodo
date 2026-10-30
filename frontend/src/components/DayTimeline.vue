<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from "vue";
import { ChevronLeft, ChevronRight } from "lucide-vue-next";
import { runMotion, cancelMotion } from "../composables/motion";
import { hasDirtyEditor } from "../composables/editorGuard";
import { parseDay, shiftDay } from "../api/types";
const props = defineProps<{ date: string; today?: string }>();
const emit = defineEmits<{
  select: [string];
  step: [number];
  today: [];
  drop: [DragEvent, string];
}>();
const track = ref<HTMLElement>(),
  pill = ref<HTMLElement>(),
  anchor = ref<number | null>(null),
  offset = ref(0),
  dragging = ref(false);
const days = computed(() => {
  const n = anchor.value ?? (parseDay(props.date).getDay() + 6) % 7;
  return Array.from({ length: 7 }, (_, i) => shiftDay(props.date, i - n));
});
let previousX: number | undefined, resize: ResizeObserver;
let gesture:
  | {
      id: number;
      x: number;
      date: string;
      width: number;
      target: HTMLElement;
      moved: boolean;
      last: string;
    }
  | undefined;
let suppressClick = false;
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
  if (previousX !== undefined && !dragging.value)
    runMotion(
      pill.value,
      { transform: [`translateX(${current}px)`, `translateX(${x}px)`] },
      0.28,
    );
  previousX = x;
}
function begin(event: PointerEvent, day: string, i: number) {
  if (event.button !== 0 || day !== props.date || hasDirtyEditor()) return;
  const target = event.currentTarget as HTMLElement;
  anchor.value = i;
  gesture = {
    id: event.pointerId,
    x: event.clientX,
    date: day,
    width: track.value!.clientWidth / 7,
    target,
    moved: false,
    last: day,
  };
  target.setPointerCapture(event.pointerId);
}
function move(event: PointerEvent) {
  if (!gesture || gesture.id !== event.pointerId) return;
  const dx = event.clientX - gesture.x;
  if (!gesture.moved && Math.abs(dx) < 5) return;
  gesture.moved = true;
  dragging.value = true;
  event.preventDefault();
  const steps = Math.round(dx / gesture.width);
  const candidate = shiftDay(gesture.date, -steps);
  if (candidate < "2000-01-01" || candidate > "2100-12-31") {
    offset.value = 0;
    return;
  }
  offset.value = dx - steps * gesture.width;
  if (candidate !== gesture.last) {
    gesture.last = candidate;
    emit("select", candidate);
  }
}
function end(event: PointerEvent) {
  if (!gesture || gesture.id !== event.pointerId) return;
  const g = gesture;
  gesture = undefined;
  suppressClick = g.moved;
  setTimeout(() => {
    suppressClick = false;
  }, 0);
  if (g.target.hasPointerCapture(event.pointerId))
    g.target.releasePointerCapture(event.pointerId);
  dragging.value = false;
  const from = offset.value;
  offset.value = 0;
  for (const el of Array.from(
    track.value?.querySelectorAll<HTMLElement>(".timeline-date-label") || [],
  )) {
    runMotion(
      el,
      { transform: [`translateX(${from}px)`, "translateX(0px)"] },
      0.18,
    );
  }
  void position();
}
function select(day: string) {
  if (suppressClick) {
    suppressClick = false;
    return;
  }
  emit("select", day);
}
function step(n: number) {
  anchor.value = null;
  emit("step", n);
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
  gesture = undefined;
});
</script>
<template>
  <div
    class="compact-week smooth-timeline"
    :class="{ 'is-scrubbing': dragging }"
  >
    <button class="icon-button" aria-label="前一天" @click="step(-1)">
      <ChevronLeft :size="14" />
    </button>
    <div ref="track" class="timeline-days">
      <span ref="pill" class="timeline-pill" aria-hidden="true" />
      <button
        v-for="(day, i) in days"
        :key="i"
        class="week-date"
        :class="{ selected: date === day, today: today === day }"
        :data-timeline-day="day"
        :aria-label="
          '周' +
          ['日', '一', '二', '三', '四', '五', '六'][parseDay(day).getDay()] +
          ' ' +
          Number(day.slice(-2))
        "
        :aria-current="date === day ? 'date' : undefined"
        @pointerdown="begin($event, day, i)"
        @pointermove="move"
        @pointerup="end"
        @pointercancel="end"
        @lostpointercapture="end"
        @click="select(day)"
        @dragover.prevent
        @drop="emit('drop', $event, day)"
      >
        <span
          class="timeline-date-label"
          :style="{ transform: `translateX(${offset}px)` }"
          ><small>{{
            ["日", "一", "二", "三", "四", "五", "六"][parseDay(day).getDay()]
          }}</small
          >{{ Number(day.slice(-2)) }}</span
        >
      </button>
    </div>
    <button class="icon-button" aria-label="后一天" @click="step(1)">
      <ChevronRight :size="14" />
    </button>
    <button
      class="text-button"
      @click="
        anchor = null;
        emit('today');
      "
    >
      今天
    </button>
  </div>
</template>
