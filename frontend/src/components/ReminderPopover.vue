<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from "vue";
import { Bell, X } from "lucide-vue-next";
import { useTodo } from "../stores/todo";
import { api } from "../api";
import type { Task } from "../api/types";
const props = defineProps<{ task: Task }>(),
  emit = defineEmits<{ close: [] }>(),
  s = useTodo();
const panel = ref<HTMLElement>(),
  value = ref(""),
  expectedVersion = props.task.version;
function local(ms: number) {
  const d = new Date(ms);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}T${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
}
value.value = local(props.task.remindAt || Date.now() + 3600000);
async function save(clear = false) {
  const at = clear ? null : new Date(value.value).getTime();
  if (!clear && (!Number.isFinite(at) || at! <= Date.now())) {
    s.error = "请选择未来的提醒时间";
    return;
  }
  const t = props.task;
  const result = await s.run(() =>
    api<Task[]>("Update", {
      ...s.command(t),
      expectedVersion,
      title: t.title,
      description: t.description || "",
      date: t.date || value.value.slice(0, 10),
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
function outside(e: PointerEvent) {
  if (
    !panel.value?.contains(e.target as Node) &&
    !(e.target as HTMLElement).closest("[data-reminder-trigger]")
  )
    emit("close");
}
onMounted(async () => {
  await nextTick();
  const el = panel.value!,
    anchor = document
      .querySelector(`[data-reminder-trigger="${props.task.id}"]`)
      ?.getBoundingClientRect();
  if (anchor)
    Object.assign(el.style, {
      left: Math.max(8, Math.min(innerWidth - 296, anchor.right - 280)) + "px",
      top: Math.min(innerHeight - 246, anchor.bottom + 6) + "px",
    });
  el.showPopover();
  el.querySelector("input")?.focus();
  document.addEventListener("pointerdown", outside);
});
onUnmounted(() => document.removeEventListener("pointerdown", outside));
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
      <p v-if="s.error" class="inline-error" role="alert">{{ s.error }}</p>
      <footer>
        <button
          type="button"
          class="text-button"
          :disabled="!task.remindAt || s.saving"
          @click="save(true)"
        >
          移除提醒</button
        ><button class="button primary" :disabled="s.saving">保存提醒</button>
      </footer>
    </form>
  </div>
</template>
