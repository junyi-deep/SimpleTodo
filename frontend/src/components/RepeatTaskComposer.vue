<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { Repeat2, CalendarDays, Bell } from "lucide-vue-next";
import CustomSelect from "./CustomSelect.vue";
import { useTodo } from "../stores/todo";
import { api, confirmAction } from "../api";
import { parseDay, shiftDay, type Task } from "../api/types";
import { registerEditorGuard } from "../composables/editorGuard";
const props = defineProps<{ days: string[] }>();
const emit = defineEmits<{ close: []; created: [string] }>();
const s = useTodo(),
  title = ref(""),
  start = ref(props.days[0]!),
  end = ref(""),
  frequency = ref("weekly"),
  interval = ref(1),
  time = ref("");
const weekdays = ref(
  [...new Set(props.days.map((d) => parseDay(d).getDay()))].sort(),
);
const names = ["日", "一", "二", "三", "四", "五", "六"];
const summary = computed(() =>
  frequency.value === "weekly"
    ? `每 ${interval.value} 周的${weekdays.value.map((n) => "周" + names[n]).join("、")}`
    : frequency.value === "daily"
      ? `每 ${interval.value} 天`
      : `每 ${interval.value} 个月，缺少对应日期时使用月末`,
);
const state = () =>
  JSON.stringify([
    title.value,
    start.value,
    end.value,
    frequency.value,
    interval.value,
    time.value,
    weekdays.value,
  ]);
const initial = state();
let saved = false;
const release = registerEditorGuard(
  () => !saved && state() !== initial,
  async () =>
    !s.saving &&
    (saved ||
      state() === initial ||
      (await confirmAction("放弃未保存的重复任务？"))),
);
onUnmounted(release);
const host = ref<HTMLElement>();
onMounted(() => host.value?.querySelector("input")?.focus());
function toggle(n: number) {
  weekdays.value = weekdays.value.includes(n)
    ? weekdays.value.filter((x) => x !== n)
    : [...weekdays.value, n].sort();
}
async function cancel() {
  if (
    !s.saving &&
    (state() === initial || (await confirmAction("放弃未保存的重复任务？")))
  )
    emit("close");
}
async function save() {
  if (!title.value.trim() || s.saving) return;
  if (
    !start.value ||
    (end.value && end.value < start.value) ||
    (frequency.value === "weekly" && !weekdays.value.length)
  ) {
    s.error = "请选择有效的起止日期和至少一个重复星期";
    return;
  }
  let first = start.value;
  if (frequency.value === "weekly")
    for (
      let n = 0;
      n < 7 && !weekdays.value.includes(parseDay(first).getDay());
      n++
    )
      first = shiftDay(first, 1);
  if (end.value && first > end.value) {
    s.error = "起止日期内没有符合规则的任务日期";
    return;
  }
  const remindAt = time.value
    ? new Date(`${first}T${time.value}:00`).getTime()
    : null;
  if (remindAt && remindAt <= Date.now()) {
    s.error = "首次提醒时间须晚于当前时间，请调整开始日期或提醒时间";
    return;
  }
  const result = await s.run(() =>
    api<Task[]>("Create", {
      commandId: crypto.randomUUID(),
      title: title.value.trim(),
      date: first,
      description: "",
      color: "none",
      postpone: "off",
      remindAt,
      reminderTime: time.value || null,
      repeatRule: {
        frequency: frequency.value,
        interval: interval.value,
        weekdays: weekdays.value,
        until: end.value || null,
      },
    }),
  );
  if (!result) return;
  saved = true;
  if (time.value) await s.requestNotifications();
  emit("created", first);
}
</script>
<template>
  <form
    ref="host"
    class="repeat-composer"
    aria-label="创建重复任务"
    @submit.prevent="save"
    @keydown.esc.stop.prevent="cancel"
  >
    <div class="dialog-body">
      <label class="field"
        ><span>任务名称</span
        ><input
          v-model="title"
          v-sparks
          aria-label="重复任务名称"
          placeholder="准备重复做什么？"
          maxlength="500"
          required
      /></label>
      <p class="muted small">
        已选 {{ days.length }} 天，默认按对应星期重复。{{ summary }}，{{
          end ? "直到 " + end : "一直重复"
        }}。
      </p>
      <label class="field"
        ><span><Repeat2 :size="14" />重复方式</span
        ><CustomSelect v-model="frequency" aria-label="重复方式"
          ><option value="daily">每日</option>
          <option value="weekly">每周</option>
          <option value="monthly">每月</option></CustomSelect
        ></label
      >
      <div
        v-if="frequency === 'weekly'"
        class="weekday-picker"
        role="group"
        aria-label="重复星期"
      >
        <button
          v-for="(name, n) in names"
          :key="n"
          type="button"
          :aria-label="'每周' + name"
          :aria-pressed="weekdays.includes(n)"
          :class="{ active: weekdays.includes(n) }"
          @click="toggle(n)"
        >
          {{ name }}
        </button>
      </div>
      <label class="field"
        ><span>重复间隔</span
        ><input
          v-model.number="interval"
          type="number"
          min="1"
          max="99"
          required
      /></label>
      <div class="field-grid">
        <label class="field"
          ><span><CalendarDays :size="14" />开始日期</span
          ><input
            v-model="start"
            aria-label="开始日期"
            type="date"
            min="2000-01-01"
            max="2100-12-31"
            required /></label
        ><label class="field"
          ><span>结束日期（可选）</span
          ><input
            v-model="end"
            aria-label="结束日期（可选）"
            type="date"
            :min="start"
            max="2100-12-31"
          /><small>留空一直重复</small></label
        >
      </div>
      <label class="field"
        ><span><Bell :size="14" />提醒时间（可选）</span
        ><input v-model="time" type="time" /><small
          >每次任务在同一时间提醒</small
        ></label
      >
      <p v-if="s.error" class="inline-error" role="alert">{{ s.error }}</p>
    </div>
    <footer class="dialog-footer">
      <button type="button" class="button" @click="cancel">取消</button
      ><button class="button primary" :disabled="s.saving || !title.trim()">
        {{ s.saving ? "创建中…" : "创建重复任务" }}
      </button>
    </footer>
  </form>
</template>
