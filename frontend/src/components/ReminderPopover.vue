<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { Bell, X } from "lucide-vue-next";
import { api } from "../api";
import { useTodo } from "../stores/todo";
import type { Task } from "../api/types";
const props = defineProps<{ task: Task }>(),
  emit = defineEmits<{ close: [] }>(),
  s = useTodo(),
  panel = ref<HTMLElement>(),
  error = ref(""),
  busy = ref(false);
function local(ms: number) {
  const d = new Date(ms);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}T${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}
const value = ref(local(props.task.remindAt || Date.now() + 3600000));
async function save(clear = false) {
  const at = clear ? null : new Date(value.value).getTime();
  if (at !== null && (!Number.isFinite(at) || at <= Date.now())) {
    error.value = "请选择未来的提醒时间";
    return;
  }
  busy.value = true;
  const t = s.taskMap.get(props.task.id) || props.task;
  try {
    await api("Update", {
      ...s.command(t),
      title: t.title,
      description: t.description || "",
      date: t.date || value.value.slice(0, 10),
      color: "none",
      postpone: t.postpone || "inherit",
      remindAt: at,
      reminderTime: clear ? null : value.value.slice(11, 16),
    });
    await s.load();
    if (!clear) await s.requestNotifications();
    emit("close");
  } catch (e) {
    error.value = String(e);
  } finally {
    busy.value = false;
  }
}
function outside(e: PointerEvent) {
  if (
    !panel.value?.contains(e.target as Node) &&
    !(e.target as HTMLElement).closest("[data-reminder-trigger]") &&
    !busy.value
  )
    emit("close");
}
onMounted(() => {
  document.addEventListener("pointerdown", outside);
  const trigger = document.querySelector(
    `[data-reminder-trigger="${props.task.id}"]`,
  );
  if (panel.value && trigger) {
    const r = trigger.getBoundingClientRect();
    panel.value.style.position = "fixed";
    panel.value.style.left =
      Math.max(8, Math.min(innerWidth - 284, r.right - 276)) + "px";
    panel.value.style.top = Math.min(innerHeight - 236, r.bottom + 6) + "px";
  }
  panel.value?.querySelector("input")?.focus();
});
onUnmounted(() => document.removeEventListener("pointerdown", outside));
</script>
<template>
  <Teleport to="body"
    ><div
      ref="panel"
      v-motion:popover
      class="reminder-popover"
      role="dialog"
      aria-label="设置任务提醒"
      @pointerdown.stop
      @keydown.esc.stop.prevent="emit('close')"
    >
      <header>
        <strong><Bell :size="14" />提醒时间</strong
        ><button
          class="icon-button"
          aria-label="关闭提醒设置"
          @click="emit('close')"
        >
          <X :size="14" />
        </button>
      </header>
      <input v-model="value" type="datetime-local" aria-label="提醒时间" />
      <p class="muted small">时间范围内的每天任务使用相同时刻提醒。</p>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <footer>
        <button
          v-if="task.remindAt"
          class="text-button"
          :disabled="busy"
          @click="save(true)"
        >
          移除提醒</button
        ><button class="button primary" :disabled="busy" @click="save()">
          保存提醒
        </button>
      </footer>
    </div></Teleport
  >
</template>
