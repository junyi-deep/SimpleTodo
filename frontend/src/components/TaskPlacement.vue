<script setup lang="ts">
import CustomSelect from "./CustomSelect.vue";
import { restoreDialogFocus } from "../composables/dialogFocus";
import { computed, onMounted, onUnmounted, ref } from "vue";
import { X, Repeat2 } from "lucide-vue-next";
import { api, confirmAction } from "../api";
import { type Task, parseDay, shiftDay } from "../api/types";
import { useTodo } from "../stores/todo";
import { registerEditorGuard } from "../composables/editorGuard";
const props = defineProps<{ task: Task }>(),
  emit = defineEmits<{ close: [] }>(),
  s = useTodo(),
  dialog = ref<HTMLDialogElement>();
const parent = ref(props.task.parentId || props.task.rootId),
  date = ref(props.task.date || s.boot?.today || s.date),
  frequency = ref("none"),
  interval = ref(1),
  until = ref(""),
  weekdays = ref<number[]>([parseDay(props.task.date || s.date).getDay()]);
const state = () =>
  JSON.stringify([
    parent.value,
    date.value,
    frequency.value,
    interval.value,
    until.value,
    weekdays.value,
  ]);
const initial = state();
const canLeave = async () =>
  !s.saving &&
  (state() === initial || (await confirmAction("放弃未保存的计划设置？")));
const release = registerEditorGuard(() => state() !== initial, canLeave);
onUnmounted(release);
async function cancel() {
  if (await canLeave()) emit("close");
}
function toggleDay(n: number) {
  weekdays.value = weekdays.value.includes(n)
    ? weekdays.value.filter((x) => x !== n)
    : [...weekdays.value, n];
}
const expectedVersion = ref(
  s.taskMap.get(props.task.rootId)?.version || props.task.version,
);
const latest = ref<Task | null>(null),
  nodes = ref(s.snapshot.tasks.filter((t) => t.rootId === props.task.rootId));
async function refresh() {
  const rows = await s.run(() => api<Task[]>("ReadTree", props.task.rootId));
  if (rows) {
    nodes.value = rows;
    latest.value = rows.find((t) => !t.parentId) || null;
  }
}
async function retry() {
  if (
    latest.value &&
    (await confirmAction("基于最新版本重新提交当前移动选择？"))
  ) {
    expectedVersion.value = latest.value.version;
    await save();
  }
}

const candidates = computed(() =>
  nodes.value.filter((t) => t.rootId === props.task.rootId && !isDescendant(t)),
);
function isDescendant(t: Task) {
  let cur: Task | undefined = t;
  while (cur) {
    if (cur.id === props.task.id) return true;
    cur = cur.parentId
      ? nodes.value.find((t) => t.id === cur!.parentId)
      : undefined;
  }
  return false;
}
function path(t: Task) {
  const parts = [t.title];
  let cur = t;
  while (cur.parentId) {
    const p = nodes.value.find((t) => t.id === cur!.parentId);
    if (!p) break;
    parts.unshift(p.title);
    cur = p;
  }
  return parts.join(" / ");
}
restoreDialogFocus(props.task.id);
onMounted(() => {
  dialog.value?.showModal();
  dialog.value?.querySelector<HTMLInputElement>("input")?.focus();
});

async function save() {
  if (s.saving) return;
  let plannedDate = date.value;
  if (frequency.value === "weekly") {
    if (!weekdays.value.length) {
      s.error = "请至少选择一个重复星期";
      return;
    }
    while (!weekdays.value.includes(parseDay(plannedDate).getDay()))
      plannedDate = shiftDay(plannedDate, 1);
  }
  if (frequency.value !== "none" && until.value && until.value < plannedDate) {
    s.error = "起止日期内没有符合规则的任务日期";
    return;
  }
  let ok = false;
  const c = {
    ...s.command(props.task),
    expectedVersion: expectedVersion.value,
  };
  if (props.task.parentId) {
    await s.run(async () => {
      await api("Reparent", { ...c, parentId: parent.value });
      ok = true;
    });
  } else {
    await s.run(async () => {
      await api("Move", {
        ...c,
        date: plannedDate,
        repeatRule:
          frequency.value === "none"
            ? null
            : {
                frequency: frequency.value,
                interval: interval.value,
                weekdays: weekdays.value,
                until: until.value || null,
              },
      });
      ok = true;
    });
  }
  if (ok) emit("close");
}
</script>
<template>
  <Teleport to="body"
    ><dialog
      :inert="s.maintenance || undefined"
      ref="dialog"
      v-motion
      class="task-settings-dialog placement-dialog"
      :aria-label="task.parentId ? '移动整个分支' : '调整计划日期'"
      @cancel.prevent="cancel()"
      @keydown.esc.stop.prevent="cancel()"
      @keydown.meta.enter.prevent="save"
      @keydown.ctrl.enter.prevent="save"
    >
      <form @submit.prevent="save">
        <header class="dialog-header">
          <span>{{ task.parentId ? "移动整个分支" : "调整计划日期" }}</span
          ><button
            type="button"
            class="icon-button"
            aria-label="关闭移动面板"
            @click="cancel()"
          >
            <X :size="18" />
          </button>
        </header>
        <div class="dialog-body">
          <p class="muted small placement-task-title">{{ task.title }}</p>
          <label v-if="task.parentId" class="field"
            ><span>新的父任务</span
            ><CustomSelect v-model="parent">
              <option v-for="t in candidates" :key="t.id" :value="t.id">
                {{ path(t) }}
              </option></CustomSelect
            ><small>携带完整子树；不能跨根任务或移入自身后代。</small></label
          ><template v-else
            ><label class="field"
              ><span>计划日期</span
              ><input
                v-model="date"
                type="date"
                min="2000-01-01"
                max="2100-12-31"
                required
            /></label>
            <section class="repeat-fields">
              <label class="field"
                ><span><Repeat2 :size="14" />重复</span
                ><CustomSelect v-model="frequency" aria-label="计划重复方式"
                  ><option value="none">不重复</option>
                  <option value="daily">每日</option>
                  <option value="weekly">每周</option>
                  <option value="monthly">
                    每月（缺少日期时使用月末）
                  </option></CustomSelect
                ></label
              >
              <template v-if="frequency !== 'none'">
                <p class="muted small">
                  从计划日期开始，结束日期留空则一直重复。
                </p>
                <div class="field-grid">
                  <label class="field"
                    ><span>重复间隔</span
                    ><input
                      v-model.number="interval"
                      type="number"
                      min="1"
                      max="99"
                      required /></label
                  ><label class="field"
                    ><span>结束日期（可选）</span
                    ><input
                      v-model="until"
                      type="date"
                      :min="date"
                      max="2100-12-31"
                  /></label>
                </div>
                <div
                  v-if="frequency === 'weekly'"
                  class="weekday-picker"
                  role="group"
                  aria-label="重复星期"
                >
                  <button
                    v-for="(label, n) in [
                      '日',
                      '一',
                      '二',
                      '三',
                      '四',
                      '五',
                      '六',
                    ]"
                    :key="n"
                    type="button"
                    :aria-label="'每周' + label"
                    :aria-pressed="weekdays.includes(n)"
                    :class="{ active: weekdays.includes(n) }"
                    @click="toggleDay(n)"
                  >
                    {{ label }}
                  </button>
                </div>
              </template>
            </section>
          </template>
          <p v-if="s.error" class="inline-error" role="alert">{{ s.error }}</p>
          <template v-if="s.error.includes('CONFLICT')"
            ><button type="button" class="button" @click="refresh">
              加载最新任务树
            </button></template
          >
          <div v-if="latest" class="path-output">
            最新根任务：{{ latest.title }}<br />{{ latest.date || "收集箱" }} ·
            当前移动选择仍保留<button
              type="button"
              class="button"
              @click="retry"
            >
              基于最新版本确认移动
            </button>
          </div>
        </div>
        <footer class="dialog-footer">
          <button type="button" class="button" @click="cancel()">取消</button
          ><button class="button primary" :disabled="s.saving">确认移动</button>
        </footer>
      </form>
    </dialog></Teleport
  >
</template>
