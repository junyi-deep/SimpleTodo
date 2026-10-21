<script setup lang="ts">
import CustomSelect from "./CustomSelect.vue";
import { registerEditorGuard } from "../composables/editorGuard";
import { restoreDialogFocus } from "../composables/dialogFocus";
import { ref, onMounted, onUnmounted, nextTick, watch } from "vue";
import { X, CalendarDays, Bell, Repeat2, AlignLeft } from "lucide-vue-next";
import { type Task, type Rule } from "../api/types";
import { api, confirmAction } from "../api";
import { useTodo } from "../stores/todo";
const props = defineProps<{ task: Task }>(),
  emit = defineEmits<{ close: [] }>(),
  s = useTodo(),
  dialog = ref<HTMLDialogElement>();
const expectedVersion = ref(
  s.taskMap.get(props.task.rootId)?.version || props.task.version,
);
const latest = ref<Task | null>(null);
const composing = ref(false);
const editCommand = () => ({
  ...s.command(props.task),
  expectedVersion: expectedVersion.value,
});
async function latestVersion() {
  latest.value = await api<Task>("FindTask", props.task.rootId);
}
async function rebase() {
  if (
    latest.value &&
    (await confirmAction("使用最新版本作为基础，重新提交当前草稿？"))
  ) {
    expectedVersion.value = latest.value.version;
    latest.value = null;
    await save();
  }
}

const title = ref(props.task.title),
  description = ref(props.task.description || ""),
  date = ref(props.task.date || ""),
  color = ref(props.task.color || "none"),
  postpone = ref(props.task.postpone || "inherit");
const reminder = ref(props.task.remindAt ? toLocal(props.task.remindAt) : ""),
  frequency = ref("none"),
  interval = ref(1),
  until = ref(""),
  weekdays = ref<number[]>([new Date().getDay()]),
  following = ref(false);
watch(
  () => date.value,
  (value, old) => {
    if (reminder.value && value && old && value !== old)
      reminder.value = value + reminder.value.slice(10);
    if (!value) reminder.value = "";
  },
);
const formState = () =>
  JSON.stringify([
    title.value,
    description.value,
    date.value,
    color.value,
    postpone.value,
    reminder.value,
    frequency.value,
    interval.value,
    until.value,
    [...weekdays.value].sort(),
    following.value,
  ]);
let initialState = formState();
const dirty = () => formState() !== initialState;
async function canLeave() {
  return !s.saving && (!dirty() || (await confirmAction("放弃未保存的修改？")));
}
const releaseGuard = registerEditorGuard(dirty, canLeave);
onUnmounted(releaseGuard);
function toLocal(ms: number) {
  const d = new Date(ms);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}T${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}
restoreDialogFocus(props.task.id);
onMounted(async () => {
  dialog.value?.showModal();
  dialog.value?.querySelector("input")?.focus();
  if (props.task.seriesId) {
    try {
      const rules = await api<Rule[]>("Rules");
      const rule = rules.find((r) => r.id === props.task.ruleId);
      if (rule) {
        frequency.value = rule.frequency;
        interval.value = rule.interval;
        until.value = rule.until || "";
        weekdays.value = rule.weekdays || [];
        initialState = formState();
      }
    } catch (e) {
      s.error = String(e);
    }
  }
});
async function save(addChild = false) {
  if (!title.value.trim() || s.saving || composing.value) return;
  const t = props.task;
  const repeatRule =
    !t.parentId &&
    frequency.value !== "none" &&
    (!t.seriesId || following.value)
      ? {
          frequency: frequency.value,
          interval: interval.value,
          until: until.value || null,
          weekdays: weekdays.value,
        }
      : null;
  if (t.seriesId && following.value && frequency.value === "none") {
    s.error = "请使用停止本次及以后的重复，单次任务可复制为普通任务";
    return;
  }
  if (repeatRule && t.seriesId) {
    const impact = await s.run(() =>
      api<{ updated: number; preserved: number }>(
        "SeriesImpact",
        editCommand(),
      ),
    );
    if (
      !impact ||
      !(await confirmAction(
        `修改本次及以后：更新 ${impact.updated} 个未完成实例，保留 ${impact.preserved} 个已完成、跳过或单次例外。继续保存？`,
      ))
    )
      return;
  }
  const changed = await s.run(() =>
    api<Task[]>("Update", {
      ...editCommand(),
      title: title.value.trim(),
      description: description.value,
      date: date.value || null,
      color: "none",
      postpone: postpone.value,
      remindAt: reminder.value ? new Date(reminder.value).getTime() : null,
      reminderTime: reminder.value ? reminder.value.slice(11, 16) : null,
      repeatRule,
      following: following.value,
    }),
  );
  if (!changed) return;
  if (reminder.value) await s.requestNotifications();
  emit("close");
  if (addChild) {
    await nextTick();
    document.dispatchEvent(
      new CustomEvent("todo:add-child", {
        detail: { id: t.id, rootId: t.rootId },
      }),
    );
  }
}
async function stopFollowing() {
  const impact = await s.run(() =>
    api<{ removed: number; nodes: number; preserved: number }>(
      "SeriesImpact",
      editCommand(),
    ),
  );
  if (
    impact &&
    (await confirmAction(
      `停止本次及以后的重复？删除本次及以后 ${impact.removed} 个未完成根任务及其后代（共 ${impact.nodes} 个节点），保留已完成和跳过记录。`,
    ))
  ) {
    let ok = false;
    await s.run(async () => {
      await api("StopFollowing", editCommand());
      ok = true;
    });
    if (ok) emit("close");
  }
}
function toggleWeekday(n: number) {
  weekdays.value = weekdays.value.includes(n)
    ? weekdays.value.filter((w) => w !== n)
    : [...weekdays.value, n];
}
async function cancel() {
  if (await canLeave()) emit("close");
}
</script>
<template>
  <Teleport to="body"
    ><dialog
      :inert="s.maintenance || undefined"
      ref="dialog"
      class="task-settings-dialog"
      aria-label="任务详细设置"
      @cancel.prevent="cancel"
      @keydown.esc.stop.prevent="cancel"
      @keydown.meta.enter.prevent="save()"
      @keydown.ctrl.enter.prevent="save()"
    >
      <form @submit.prevent="save()">
        <header class="dialog-header">
          <span>{{ task.parentId ? "编辑子任务" : "任务详情" }}</span
          ><button
            type="button"
            class="icon-button"
            aria-label="关闭"
            @click="cancel"
          >
            <X :size="18" />
          </button>
        </header>
        <div class="dialog-body">
          <input
            v-model="title"
            class="title-input"
            @compositionstart="composing = true"
            @compositionend="composing = false"
            aria-label="任务标题"
            @keydown.shift.enter.prevent="save(true)"
            placeholder="准备做什么？"
            maxlength="500"
            required
          />
          <p v-if="task.parentId" class="muted small">
            子任务继承根任务的日期和提醒。你可以继续把它拆分为更细的步骤。
          </p>
          <template v-else>
            <label class="field"
              ><span><AlignLeft :size="14" />描述</span
              ><textarea
                v-model="description"
                placeholder="补充说明、链接或执行步骤…"
                rows="4"
                maxlength="20000"
              />
            </label>
            <div class="field-grid">
              <label class="field"
                ><span><CalendarDays :size="14" />计划日期</span
                ><input
                  v-model="date"
                  type="date"
                  :disabled="!!task.seriesId"
                  min="2000-01-01"
                  max="2100-12-31"
                /><small>留空即存入收集箱</small></label
              ><label class="field"
                ><span>自动延期</span
                ><CustomSelect v-model="postpone" :disabled="!!task.seriesId">
                  <option value="inherit">跟随全局设置</option>
                  <option value="on">开启</option>
                  <option value="off">关闭</option>
                </CustomSelect></label
              >
            </div>
            <label class="field"
              ><span><Bell :size="14" />提醒</span
              ><input
                v-model="reminder"
                type="datetime-local"
                :disabled="!date"
              /><small v-if="!s.boot?.notificationAllowed"
                >系统通知未授权。<button
                  type="button"
                  class="text-button"
                  @click="
                    s.run(() => api('Permission')).then(() => s.metadata())
                  "
                >
                  申请通知权限
                </button></small
              ><small v-else>完全退出应用期间不保证准时提醒</small></label
            >
            <p class="muted small">
              系统时区：{{
                s.boot?.timezone ||
                Intl.DateTimeFormat().resolvedOptions().timeZone
              }}
            </p>
            <p
              v-if="
                task.originalDate && task.remindAt && task.remindAt < Date.now()
              "
              class="muted small"
            >
              自动延期已取消旧提醒，请重新选择未来时间。
            </p>
            <section class="repeat-fields">
              <label class="field"
                ><span><Repeat2 :size="14" />重复</span
                ><CustomSelect
                  v-model="frequency"
                  :disabled="!!task.seriesId && !following"
                >
                  <option value="none">不重复</option>
                  <option value="daily">每日</option>
                  <option value="weekly">每周</option>
                  <option value="monthly">每月（当月无此日时使用月末）</option>
                </CustomSelect></label
              >
              <label v-if="task.seriesId" class="check-label"
                ><input
                  v-model="following"
                  type="checkbox"
                  :disabled="!!task.date && task.date < (s.boot?.today || '')"
                />
                修改本次及以后（保留已完成和单次例外）</label
              >
              <div v-if="frequency !== 'none'" class="field-grid">
                <label class="field"
                  ><span>重复间隔</span
                  ><input
                    v-model.number="interval"
                    type="number"
                    min="1"
                    max="99"
                    :disabled="!!task.seriesId && !following" /></label
                ><label class="field"
                  ><span>结束日期（可选）</span
                  ><input
                    v-model="until"
                    type="date"
                    :min="date"
                    :disabled="!!task.seriesId && !following"
                /></label>
              </div>
              <div v-if="frequency === 'weekly'" class="weekday-picker">
                <button
                  v-for="(label, i) in [
                    '日',
                    '一',
                    '二',
                    '三',
                    '四',
                    '五',
                    '六',
                  ]"
                  :key="i"
                  type="button"
                  :class="{ active: weekdays.includes(i) }"
                  :disabled="!!task.seriesId && !following"
                  @click="toggleWeekday(i)"
                >
                  {{ label }}
                </button>
              </div>
              <button
                v-if="task.seriesId"
                type="button"
                class="text-button danger"
                @click="stopFollowing"
              >
                停止本次及以后的重复
              </button>
            </section>
            <p v-if="task.originalDate" class="muted small">
              原计划日期：{{ task.originalDate }}，已自动延期至
              {{ task.date }}。
            </p>
          </template>
          <p v-if="s.error" class="inline-error" role="alert">{{ s.error }}</p>
          <template v-if="s.error.includes('CONFLICT')"
            ><button type="button" class="button" @click="latestVersion">
              查看最新版本
            </button>
            <div v-if="latest" class="path-output">
              最新标题：{{ latest.title }}<br />{{ latest.description
              }}<br /><button type="button" class="button" @click="rebase">
                基于最新版本重试
              </button>
            </div></template
          >
        </div>
        <footer class="dialog-footer">
          <button type="button" class="button" @click="cancel">取消</button
          ><button class="button primary" :disabled="s.saving || !title.trim()">
            {{ s.saving ? "保存中…" : "保存修改" }}
          </button>
        </footer>
      </form>
    </dialog></Teleport
  >
</template>
