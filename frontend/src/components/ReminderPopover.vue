<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from "vue";
import { Bell, X } from "lucide-vue-next";
import { useTodo } from "../stores/todo";
import CustomSelect from "./CustomSelect.vue";
import { api, confirmAction } from "../api";
import type { Task, Rule } from "../api/types";
const props = defineProps<{ task: Task }>(),
  emit = defineEmits<{ close: [] }>(),
  s = useTodo();
const panel = ref<HTMLElement>(),
  value = ref(""),
  frequency = ref("none"),
  interval = ref(1),
  weekdays = ref<number[]>([new Date().getDay()]),
  until = ref(""),
  loading = ref(!!props.task.seriesId),
  expectedVersion = props.task.version;
function local(ms: number) {
  const d = new Date(ms);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}T${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}
value.value = local(props.task.remindAt || Date.now() + 3600000);
const weekday = (v: string) => new Date(v.slice(0, 10) + "T12:00:00").getDay();
weekdays.value = [weekday(value.value)];
watch(value, (v, old) => {
  if (
    !props.task.seriesId &&
    weekdays.value.length === 1 &&
    weekdays.value[0] === weekday(old)
  )
    weekdays.value = [weekday(v)];
});
async function save(clear = false) {
  const at = clear ? null : new Date(value.value).getTime();
  if (!clear && (!Number.isFinite(at) || at! <= Date.now())) {
    s.error = "请选择未来的提醒时间";
    return;
  }
  const t = props.task;
  const repeatRule =
    !clear && frequency.value !== "none"
      ? {
          frequency: frequency.value,
          interval: interval.value,
          weekdays: weekdays.value,
          until: until.value || null,
          monthDay: Number(value.value.slice(8, 10)),
        }
      : null;
  if (repeatRule?.frequency === "weekly" && !weekdays.value.length) {
    s.error = "请选择至少一个星期";
    return;
  }
  if (repeatRule && t.seriesId && value.value.slice(0, 10) !== t.date) {
    s.error =
      "重复提醒请选择本次计划日期内的时间；其他日期可从对应日历任务设置。";
    return;
  }
  if (repeatRule && t.seriesId) {
    const impact = await s.run(() =>
      api<{ updated: number; preserved: number }>("SeriesImpact", s.command(t)),
    );
    if (
      !impact ||
      !(await confirmAction(
        `调整本次及以后 ${impact.updated} 个未完成重复任务的提醒，保留 ${impact.preserved} 个历史与例外记录。继续？`,
      ))
    )
      return;
  }

  const result = await s.run(() =>
    api<Task[]>("Update", {
      ...s.command(t),
      expectedVersion,
      repeatRule,
      following: !!t.seriesId && !clear,
      title: t.title,
      description: t.description || "",
      date:
        repeatRule && !t.seriesId
          ? value.value.slice(0, 10)
          : t.date || value.value.slice(0, 10),
      color: "none",
      postpone: t.postpone || "inherit",
      remindAt: at,
      reminderTime: clear ? null : value.value.slice(11, 16),
    }),
  );
  if (result) {
    if (!clear) await s.requestNotifications();
    emit("close");
  }
}
let observer: ResizeObserver | undefined;
function outside(e: PointerEvent) {
  if (
    !panel.value?.contains(e.target as Node) &&
    !(e.target as HTMLElement).closest("[data-reminder-trigger]")
  )
    emit("close");
}
onMounted(async () => {
  if (props.task.seriesId) {
    try {
      const rule = (await api<Rule[]>("Rules")).find(
        (r) => r.id === props.task.ruleId,
      );
      if (rule) {
        frequency.value = rule.frequency;
        interval.value = rule.interval;
        weekdays.value = rule.weekdays || [];
        until.value = rule.until || "";
      }
    } catch (e) {
      s.error = String(e);
    } finally {
      loading.value = false;
    }
  }
  await nextTick();
  const el = panel.value!;
  el.showPopover();
  function place() {
    const anchor = document
      .querySelector(`[data-reminder-trigger="${props.task.id}"]`)
      ?.getBoundingClientRect();
    if (!anchor) return;
    const height = el.getBoundingClientRect().height;
    Object.assign(el.style, {
      left:
        Math.max(
          8,
          Math.min(
            innerWidth - el.offsetWidth - 8,
            anchor.right - el.offsetWidth,
          ),
        ) + "px",
      top:
        Math.max(8, Math.min(innerHeight - height - 8, anchor.bottom + 6)) +
        "px",
    });
  }
  place();
  observer = new ResizeObserver(place);
  observer.observe(el);
  el.querySelector("input")?.focus();
  document.addEventListener("pointerdown", outside);
});
onUnmounted(() => {
  observer?.disconnect();
  document.removeEventListener("pointerdown", outside);
});
</script>
<template>
  <div
    ref="panel"
    popover="manual"
    class="reminder-popover"
    @keydown.esc.stop.prevent="emit('close')"
  >
    <header>
      <strong><Bell :size="14" />设置提醒</strong
      ><button
        class="icon-button"
        aria-label="关闭提醒设置"
        @click="emit('close')"
      >
        <X :size="14" />
      </button>
    </header>
    <form @submit.prevent="save()">
      <input
        v-model="value"
        type="datetime-local"
        aria-label="提醒时间"
        required
      />
      <label class="field"
        ><span>重复提醒</span
        ><CustomSelect
          v-model="frequency"
          aria-label="重复提醒"
          :disabled="loading"
          ><option value="none" :disabled="!!task.seriesId">不重复</option>
          <option value="daily">每天</option>
          <option value="weekly">每周</option>
          <option value="monthly">每月</option></CustomSelect
        ></label
      >
      <template v-if="frequency !== 'none'"
        ><label class="field"
          ><span>重复间隔</span
          ><input
            v-model.number="interval"
            type="number"
            min="1"
            max="99"
            aria-label="重复间隔"
            required
        /></label>
        <div v-if="frequency === 'weekly'" class="repeat-weekdays">
          <button
            v-for="(label, i) in ['日', '一', '二', '三', '四', '五', '六']"
            :key="i"
            type="button"
            :aria-label="'星期' + label"
            :aria-pressed="weekdays.includes(i)"
            :class="{ active: weekdays.includes(i) }"
            @click="
              weekdays = weekdays.includes(i)
                ? weekdays.filter((d) => d !== i)
                : [...weekdays, i]
            "
          >
            {{ label }}
          </button>
        </div>
        <label class="field"
          ><span>结束日期（可留空）</span
          ><input v-model="until" type="date" aria-label="提醒重复结束日期"
        /></label>
        <p v-if="task.seriesId" class="muted small">
          停止重复可在任务设置中操作。
        </p></template
      >
      <p v-if="s.error" class="inline-error" role="alert">{{ s.error }}</p>
      <footer>
        <button
          type="button"
          class="text-button"
          :disabled="!task.remindAt || s.saving"
          @click="save(true)"
        >
          移除提醒</button
        ><button class="button primary" :disabled="s.saving || loading">
          保存提醒
        </button>
      </footer>
    </form>
  </div>
</template>
