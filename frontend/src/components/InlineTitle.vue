<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from "vue";
import { useTodo } from "../stores/todo";
import { api } from "../api";
import type { Task } from "../api/types";
import { registerEditorGuard } from "../composables/editorGuard";
const props = defineProps<{ task?: Task; date?: string | null }>();
const emit = defineEmits<{ done: []; paste: [string] }>();
const s = useTodo(),
  input = ref<HTMLTextAreaElement>(),
  draft = ref(props.task?.title || "");
const base = props.task ? { ...props.task } : undefined,
  command = base ? s.command(base) : undefined;
const composing = ref(false);
let focusAfter = false,
  savedID = props.task?.id,
  closed = false,
  pending: Promise<boolean> | undefined;
function finish() {
  if (closed) return;
  closed = true;
  emit("done");
  if (focusAfter)
    void nextTick(() => {
      const target = savedID
        ? document.querySelector<HTMLElement>(`[data-node="${savedID}"]`)
        : document.querySelector<HTMLElement>('[aria-label="新增待办"]');
      target?.focus();
    });
}
function cancel() {
  if (!pending) {
    focusAfter = true;
    finish();
  }
}
async function commit(): Promise<boolean> {
  if (closed) return true;
  if (pending) return pending;
  if (composing.value) return false;
  const title = draft.value.trim().replace(/[\r\n]+/g, " ");
  if (!title || title === base?.title) {
    finish();
    return true;
  }
  pending = (async () => {
    const changed = base
      ? await s.run(() =>
          api<Task[]>("Update", {
            ...command,
            title,
            description: base.description || "",
            date: base.date,
            color: "none",
            postpone: base.postpone || "inherit",
            remindAt: base.remindAt,
            reminderTime: base.reminderTime,
          }),
        )
      : await s.create(title, undefined, props.date);
    if (changed) {
      if (!base) {
        savedID = changed.find((t: any) => !t.parentId)?.id;
        if (savedID) s.activeTask = savedID;
      }
      finish();
      return true;
    }
    await nextTick();
    input.value?.focus();
    return false;
  })();
  try {
    return await pending;
  } finally {
    pending = undefined;
  }
}
function resize() {
  if (input.value) {
    input.value.style.height = "auto";
    input.value.style.height = input.value.scrollHeight + "px";
  }
}
function key(e: KeyboardEvent) {
  if (e.isComposing || composing.value) return;
  if (e.key === "Escape") {
    e.preventDefault();
    e.stopPropagation();
    cancel();
  } else if (e.key === "Enter") {
    focusAfter = true;
    e.preventDefault();
    e.stopPropagation();
    void commit();
  }
}
function paste(e: ClipboardEvent) {
  const text = e.clipboardData?.getData("text/plain") || "";
  if (!base && text.trim().split(/\r?\n/).length > 1) {
    e.preventDefault();
    finish();
    emit("paste", text);
  }
}
const release = registerEditorGuard(
  () => !closed && draft.value !== (base?.title || ""),
  commit,
);
onUnmounted(release);
onMounted(async () => {
  await nextTick();
  input.value?.focus();
  input.value?.setSelectionRange(draft.value.length, draft.value.length);
  resize();
});
</script>
<template>
  <textarea
    ref="input"
    v-model="draft"
    class="inline-title-input"
    :aria-label="task ? '编辑任务标题' : '新待办标题'"
    placeholder="输入待办事项…"
    maxlength="500"
    rows="1"
    :disabled="s.saving"
    @input="resize"
    @blur="commit"
    @keydown="key"
    @paste="paste"
    @compositionstart="composing = true"
    @compositionend="composing = false"
  />
</template>
