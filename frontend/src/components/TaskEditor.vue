<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick } from "vue";
import { X, CalendarDays, Bell, AlignLeft } from "lucide-vue-next";
import CustomSelect from "./CustomSelect.vue";
import { api, confirmAction } from "../api";
import { useTodo } from "../stores/todo";
import type { Task } from "../api/types";
import { registerEditorGuard } from "../composables/editorGuard";
import { restoreDialogFocus } from "../composables/dialogFocus";
const props = defineProps<{ task: Task; embedded?: boolean }>(),
  emit = defineEmits<{ close: [] }>(),
  s = useTodo(),
  dialog = ref<HTMLDialogElement>(),
  title = ref(props.task.title),
  description = ref(props.task.description || ""),
  postpone = ref(props.task.postpone || "inherit"),
  error = ref(""),
  busy = ref(false),
  latest = ref<Task | null>(null),
  version = ref(
    s.taskMap.get(props.task.rootId)?.version || props.task.version,
  );
function local(ms: number) {
  const d = new Date(ms);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}T${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}
const start = ref(local(props.task.startAt || props.task.createdAt)),
  end = ref(props.task.endAt ? local(props.task.endAt) : ""),
  reminder = ref(props.task.remindAt ? local(props.task.remindAt) : "");
const state = () =>
  JSON.stringify([
    title.value,
    description.value,
    postpone.value,
    start.value,
    end.value,
    reminder.value,
  ]);
let initial = state();
const dirty = () => state() !== initial;
async function canLeave() {
  return (
    !busy.value && (!dirty() || (await confirmAction("放弃未保存的修改？")))
  );
}
const release = registerEditorGuard(dirty, canLeave);
onUnmounted(release);
restoreDialogFocus(props.task.id);
onMounted(() => {
  if (!props.embedded) dialog.value?.showModal();
  dialog.value?.querySelector("input")?.focus();
});
async function cancel() {
  if (await canLeave()) emit("close");
}
async function save(addChild = false) {
  const originalStart = props.task.startAt || props.task.createdAt;
  const a =
      start.value === local(originalStart)
        ? originalStart
        : new Date(start.value).getTime(),
    b = end.value
      ? props.task.endAt && end.value === local(props.task.endAt)
        ? props.task.endAt
        : new Date(end.value).getTime()
      : null;
  if (
    !title.value.trim() ||
    !Number.isFinite(a) ||
    (b !== null && (!Number.isFinite(b) || b < a))
  ) {
    error.value = "请填写标题和有效的起止时间";
    return;
  }
  busy.value = true;
  error.value = "";
  try {
    const spanChanged =
      start.value !== local(props.task.startAt || props.task.createdAt) ||
      end.value !== (props.task.endAt ? local(props.task.endAt) : "");
    await api("Update", {
      ...s.command(props.task),
      expectedVersion: version.value,
      title: title.value.trim(),
      description: description.value,
      date: props.task.date,
      color: "none",
      postpone: postpone.value,
      remindAt: reminder.value ? new Date(reminder.value).getTime() : null,
      reminderTime: reminder.value ? reminder.value.slice(11, 16) : null,
      span: spanChanged ? { start: a, end: b } : null,
    });
    if (reminder.value) await s.requestNotifications();
    initial = state();
    await s.load();
    emit("close");
    if (addChild) {
      await nextTick();
      document.dispatchEvent(
        new CustomEvent("todo:add-child", {
          detail: { id: props.task.id, rootId: props.task.rootId },
        }),
      );
    }
  } catch (e) {
    error.value = String(e);
  } finally {
    busy.value = false;
  }
}
async function latestVersion() {
  latest.value = await api<Task>("FindTask", props.task.rootId);
}
async function rebase() {
  if (latest.value && (await confirmAction("基于最新版本重新提交当前草稿？"))) {
    version.value = latest.value.version;
    latest.value = null;
    await save();
  }
}
</script>
<template>
  <component
    :is="embedded ? 'section' : 'dialog'"
    ref="dialog"
    v-motion
    :class="embedded ? 'embedded-task-editor' : 'task-settings-dialog'"
    aria-label="任务详情"
    @cancel.prevent="cancel"
    @keydown.esc.stop.prevent="cancel"
  >
    <header class="dialog-header">
      <strong>任务详情</strong
      ><button
        class="icon-button"
        aria-label="关闭任务详情"
        :disabled="busy"
        @click="cancel"
      >
        <X :size="16" />
      </button>
    </header>
    <div class="dialog-body">
      <input
        v-model="title"
        class="task-editor-title"
        aria-label="任务标题"
        maxlength="500"
        @keydown.meta.enter.prevent="save()"
      />
      <label v-if="!task.parentId" class="field"
        ><span><AlignLeft :size="14" />描述</span
        ><textarea
          v-model="description"
          aria-label="描述"
          rows="5"
          maxlength="20000"
        />
      </label>
      <div class="form-grid">
        <label class="field"
          ><span><CalendarDays :size="14" />开始时间</span
          ><input
            v-model="start"
            type="datetime-local"
            aria-label="开始时间"
            required /></label
        ><label class="field"
          ><span>结束时间（可选）</span
          ><input
            v-model="end"
            type="datetime-local"
            aria-label="结束时间"
            :min="start"
        /></label>
      </div>
      <p class="muted small">任务完成后会记录结束时间，并停止后续安排。</p>
      <template v-if="!task.parentId"
        ><label class="field"
          ><span>自动延期</span
          ><CustomSelect v-model="postpone" aria-label="自动延期"
            ><option value="inherit">跟随全局设置</option>
            <option value="on">开启</option>
            <option value="off">关闭</option></CustomSelect
          ></label
        ><label class="field"
          ><span><Bell :size="14" />提醒</span
          ><input
            v-model="reminder"
            type="datetime-local"
            aria-label="提醒时间" /></label
      ></template>
      <p class="muted small">
        创建：{{ new Date(task.createdAt).toLocaleString("zh-CN") }}<br />更新：{{
          new Date(task.updatedAt).toLocaleString("zh-CN")
        }}<template v-if="task.completedAt"
          ><br />完成：{{
            new Date(task.completedAt).toLocaleString("zh-CN")
          }}</template
        >
      </p>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <div v-if="error.includes('CONFLICT')" class="inline-actions">
        <button class="button" @click="latestVersion">查看最新版本</button
        ><button v-if="latest" class="button" @click="rebase">
          基于最新版本重试
        </button>
      </div>
    </div>
    <footer class="dialog-footer">
      <button class="button" :disabled="busy" @click="cancel">取消</button
      ><button
        class="button primary"
        :disabled="busy || !title.trim()"
        @click="save()"
      >
        保存修改
      </button>
    </footer></component
  >
</template>
