<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from "vue";
import { X } from "lucide-vue-next";
import { api, confirmAction } from "../api";
import { useTodo } from "../stores/todo";
import type { Task } from "../api/types";
import CustomSelect from "./CustomSelect.vue";
import { registerEditorGuard } from "../composables/editorGuard";
import { restoreDialogFocus } from "../composables/dialogFocus";
const props = defineProps<{ task: Task }>(),
  emit = defineEmits<{ close: [] }>(),
  s = useTodo(),
  dialog = ref<HTMLDialogElement>(),
  parent = ref(props.task.parentId || ""),
  busy = ref(false),
  error = ref("");
function local(ms: number) {
  const d = new Date(ms);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}T${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}
const start = ref(local(props.task.startAt || props.task.createdAt)),
  end = ref(props.task.endAt ? local(props.task.endAt) : "");
const initial = JSON.stringify([parent.value, start.value, end.value]);
const dirty = () =>
  JSON.stringify([parent.value, start.value, end.value]) !== initial;
const canLeave = async () =>
  !busy.value && (!dirty() || (await confirmAction("放弃未保存的时间设置？")));
const release = registerEditorGuard(dirty, canLeave);
onUnmounted(release);
restoreDialogFocus(props.task.id);
onMounted(() => dialog.value?.showModal());
const candidates = computed(() =>
  s.snapshot.tasks.filter(
    (t) =>
      t.rootId === props.task.rootId && t.id !== props.task.id && !isChild(t),
  ),
);
function isChild(t: Task) {
  let p = t.parentId;
  while (p) {
    if (p === props.task.id) return true;
    p = s.taskMap.get(p)?.parentId || null;
  }
  return false;
}
async function cancel() {
  if (await canLeave()) emit("close");
}
const expectedVersion = s.command(props.task).expectedVersion;
async function save() {
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
  if (!Number.isFinite(a) || (b !== null && b < a)) {
    error.value = "请选择有效的起止时间";
    return;
  }
  const spanChanged = a !== originalStart || b !== props.task.endAt;
  const parentChanged =
    props.task.parentId && parent.value !== props.task.parentId;
  if (!spanChanged && !parentChanged) {
    emit("close");
    return;
  }
  busy.value = true;
  try {
    const c = { ...s.command(props.task), expectedVersion };
    if (parentChanged)
      await api("Reparent", {
        ...c,
        parentId: parent.value,
        span: spanChanged ? { start: a, end: b } : null,
      });
    else await api("SetTaskSpan", c, a, b);
    await s.load();
    emit("close");
  } catch (e) {
    error.value = String(e);
  } finally {
    busy.value = false;
  }
}
</script>
<template>
  <Teleport to="body"
    ><dialog
      ref="dialog"
      v-motion
      class="task-settings-dialog task-placement-dialog"
      aria-label="调整任务时间"
      @cancel.prevent="cancel"
      @keydown.esc.stop.prevent="cancel"
    >
      <header class="dialog-header">
        <strong>调整任务时间</strong
        ><button
          class="icon-button"
          aria-label="关闭时间设置"
          :disabled="busy"
          @click="cancel"
        >
          <X :size="16" />
        </button>
      </header>
      <div class="dialog-body">
        <p class="placement-task-title">{{ task.title }}</p>
        <label v-if="task.parentId" class="field"
          ><span>父任务</span
          ><CustomSelect v-model="parent" aria-label="父任务"
            ><option v-for="t in candidates" :key="t.id" :value="t.id">
              {{ t.title }}
            </option></CustomSelect
          ></label
        ><label class="field"
          ><span>开始时间</span
          ><input
            v-model="start"
            aria-label="开始时间"
            type="datetime-local" /></label
        ><label class="field"
          ><span>结束时间（可选）</span
          ><input
            v-model="end"
            aria-label="结束时间"
            type="datetime-local"
            :min="start"
        /></label>
        <p class="muted small">每天独立完成。不设置结束时间则持续显示。</p>
        <p v-if="error" class="error" role="alert">{{ error }}</p>
      </div>
      <footer class="dialog-footer">
        <button class="button" :disabled="busy" @click="cancel">取消</button
        ><button class="button primary" :disabled="busy" @click="save">
          确认调整
        </button>
      </footer>
    </dialog></Teleport
  >
</template>
