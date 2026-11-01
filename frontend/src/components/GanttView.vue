<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import {
  ChevronLeft,
  ChevronRight,
  Infinity as InfinityIcon,
} from "lucide-vue-next";
import { api, listen } from "../api";
import { useTodo } from "../stores/todo";
import { type Task, dayString, parseDay, shiftDay } from "../api/types";
const props = defineProps<{ start: string; showChildren: boolean }>(),
  emit = defineEmits<{ edit: [Task]; changed: [] }>(),
  s = useTodo(),
  all = ref<Task[]>([]),
  from = ref(props.start),
  error = ref(""),
  busy = ref(false),
  host = ref<HTMLElement>();
const width = 30,
  days = computed(() =>
    Array.from({ length: 42 }, (_, i) => shiftDay(from.value, i)),
  );
let stop: (() => void) | undefined;
async function load() {
  if (busy.value) return;
  try {
    all.value = await api("TaskReport", { view: "gantt", limit: 10000 });
  } catch (e) {
    error.value = String(e);
  }
}
onMounted(async () => {
  await load();
  stop = await listen("todo:changed", load);
});
onUnmounted(() => {
  stop?.();
  gesture = undefined;
});
watch(
  () => props.start,
  (v) => (from.value = v),
);
const rows = computed(() => {
  const roots = new Map<string, Task>(),
    today = s.boot?.today || dayString(new Date());
  const score = (t: Task) =>
    Math.abs(
      (parseDay(t.date || today).getTime() - parseDay(today).getTime()) /
        86400000,
    ) + (t.status === "completed" ? 10000 : 0);
  for (const t of all.value.filter(
    (t) => !t.parentId && t.status !== "skipped",
  )) {
    const key = t.seriesId || t.id,
      prev = roots.get(key);
    if (!prev || score(t) < score(prev)) roots.set(key, t);
  }
  const out: { task: Task; depth: number }[] = [];
  function add(t: Task, depth: number) {
    out.push({ task: t, depth });
    if (props.showChildren)
      for (const c of all.value
        .filter((n) => n.parentId === t.id)
        .sort((a, b) => a.order - b.order))
        add(c, depth + 1);
  }
  for (const t of [...roots.values()].sort((a, b) => a.order - b.order))
    add(t, 0);
  return out;
});
const dayOffset = (ms: number) =>
  Math.round(
    (parseDay(dayString(new Date(ms))).getTime() -
      parseDay(from.value).getTime()) /
      86400000,
  );
const drag = ref<{ id: string; start: number; end: number | null } | null>(
  null,
);
function values(t: Task) {
  return drag.value?.id === t.id
    ? drag.value
    : { start: t.startAt || t.createdAt, end: t.endAt };
}
function bar(t: Task) {
  const v = values(t),
    begin = dayOffset(v.start),
    end = v.end === null ? 42 : dayOffset(v.end) + 1;
  return {
    left: Math.max(0, begin) * width + "px",
    width: Math.max(8, (Math.min(42, end) - Math.max(0, begin)) * width) + "px",
    display: begin >= 42 || end <= 0 ? "none" : undefined,
  };
}
let gesture:
  | {
      x: number;
      task: Task;
      mode: string;
      start: number;
      end: number | null;
      pointer: number;
      target: HTMLElement;
    }
  | undefined;
function begin(e: PointerEvent, t: Task, mode: string) {
  if (e.button !== 0 || busy.value) return;
  e.preventDefault();
  e.stopPropagation();
  const target = e.currentTarget as HTMLElement;
  target.setPointerCapture(e.pointerId);
  gesture = {
    x: e.clientX,
    task: t,
    mode,
    start: t.startAt || t.createdAt,
    end: t.endAt,
    pointer: e.pointerId,
    target,
  };
  drag.value = { id: t.id, start: gesture.start, end: gesture.end };
}
function move(e: PointerEvent) {
  if (!gesture || !drag.value) return;
  const n = Math.round((e.clientX - gesture.x) / width);
  const dateShift = (ms: number, n: number) => {
    const d = new Date(ms);
    d.setDate(d.getDate() + n);
    return d.getTime();
  };
  let start = gesture.start,
    end = gesture.end;
  if (gesture.mode === "move") {
    start = dateShift(start, n);
    if (end !== null) end = dateShift(end, n);
  } else if (gesture.mode === "start") {
    start = dateShift(start, n);
    if (end !== null) start = Math.min(start, end);
  } else {
    end =
      gesture.end !== null
        ? dateShift(gesture.end, n)
        : new Date(shiftDay(from.value, 41 + n) + "T23:59:59").getTime();
    end = Math.max(start, end);
  }
  drag.value = { id: gesture.task.id, start, end };
}
async function finish(e: PointerEvent) {
  const g = gesture,
    v = drag.value;
  if (!g || !v) return;
  gesture = undefined;
  if (g.target.hasPointerCapture(e.pointerId))
    g.target.releasePointerCapture(e.pointerId);
  if (v.start === g.start && v.end === g.end) {
    drag.value = null;
    return;
  }
  busy.value = true;
  error.value = "";
  try {
    const current = await api<Task>("FindTask", g.task.rootId);
    await api(
      "SetTaskSpan",
      { ...s.command(g.task), expectedVersion: current.version },
      v.start,
      v.end,
    );
    emit("changed");
  } catch (e) {
    error.value = String(e);
  } finally {
    busy.value = false;
    drag.value = null;
    await load();
  }
}
function cancel() {
  gesture = undefined;
  drag.value = null;
}
async function keyboard(e: KeyboardEvent, t: Task, mode: string) {
  if (!["ArrowLeft", "ArrowRight"].includes(e.key) || busy.value) return;
  e.preventDefault();
  const n = e.key === "ArrowRight" ? 1 : -1,
    d = new Date(t.startAt || t.createdAt);
  d.setDate(d.getDate() + n);
  let start = t.startAt || t.createdAt,
    end = t.endAt;
  if (mode === "start") start = d.getTime();
  else {
    const b = new Date(
      end || new Date(shiftDay(from.value, 41) + "T23:59:59").getTime(),
    );
    b.setDate(b.getDate() + n);
    end = b.getTime();
  }
  if (end !== null && start > end) return;
  busy.value = true;
  try {
    const current = await api<Task>("FindTask", t.rootId);
    await api(
      "SetTaskSpan",
      { ...s.command(t), expectedVersion: current.version },
      start,
      end,
    );
    emit("changed");
  } catch (e) {
    error.value = String(e);
  } finally {
    busy.value = false;
    await load();
  }
}
</script>
<template>
  <section
    class="gantt-view"
    aria-label="甘特图"
    @keydown.esc.stop.prevent="cancel"
  >
    <header class="gantt-toolbar">
      <span class="muted small">拖动任务条移动时间，拖动两端调整起止时间</span
      ><span class="grow" /><button
        class="icon-button"
        aria-label="甘特图前两周"
        @click="from = shiftDay(from, -14)"
      >
        <ChevronLeft :size="15" /></button
      ><span>{{ from }} — {{ days.at(-1) }}</span
      ><button
        class="icon-button"
        aria-label="甘特图后两周"
        @click="from = shiftDay(from, 14)"
      >
        <ChevronRight :size="15" />
      </button>
    </header>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <div ref="host" class="gantt-scroll">
      <div class="gantt-sheet" :style="{ width: 220 + 42 * width + 'px' }">
        <div class="gantt-head">
          <span class="gantt-label">任务</span
          ><span
            v-for="day in days"
            :key="day"
            class="gantt-day"
            :class="{ today: day === s.boot?.today }"
            :title="day"
            >{{ day.slice(-2)
            }}<small>{{
              "日一二三四五六"[parseDay(day).getDay()]
            }}</small></span
          >
        </div>
        <div v-for="row in rows" :key="row.task.id" class="gantt-row">
          <button
            class="gantt-label"
            :style="{ paddingLeft: 12 + row.depth * 14 + 'px' }"
            :title="row.task.title"
            @click="emit('edit', row.task)"
          >
            {{ row.task.title }}
          </button>
          <div class="gantt-track" :style="{ width: 42 * width + 'px' }">
            <span
              v-for="day in days"
              :key="day"
              class="gantt-column"
              :class="{ today: day === s.boot?.today }"
            />
            <div
              class="gantt-bar"
              :class="{
                done: row.task.status === 'completed',
                dragging: drag?.id === row.task.id,
              }"
              :style="bar(row.task)"
              :title="
                new Date(values(row.task).start).toLocaleString() +
                ' — ' +
                (values(row.task).end
                  ? new Date(values(row.task).end!).toLocaleString()
                  : '持续')
              "
              @pointerdown="begin($event, row.task, 'move')"
              @pointermove="move"
              @pointerup="finish"
              @pointercancel="cancel"
            >
              <button
                class="gantt-handle start"
                :aria-label="'调整开始时间 ' + row.task.title"
                @pointerdown.stop="begin($event, row.task, 'start')"
                @pointermove.stop="move"
                @pointerup.stop="finish"
                @keydown="keyboard($event, row.task, 'start')"
              /><span>{{ row.task.title }}</span
              ><InfinityIcon
                v-if="values(row.task).end === null"
                :size="13"
              /><button
                class="gantt-handle end"
                :aria-label="'调整结束时间 ' + row.task.title"
                @pointerdown.stop="begin($event, row.task, 'end')"
                @pointermove.stop="move"
                @pointerup.stop="finish"
                @keydown="keyboard($event, row.task, 'end')"
              />
            </div>
          </div>
        </div>
        <p v-if="!rows.length" class="muted">
          暂无任务，从任务列表或日历新增任务后即可安排时间。
        </p>
      </div>
    </div>
  </section>
</template>
