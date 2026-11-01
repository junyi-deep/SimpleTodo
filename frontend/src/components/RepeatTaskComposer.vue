<script setup lang="ts">
import { ref, onUnmounted, onMounted } from "vue";
import { useTodo } from "../stores/todo";
import { api, confirmAction } from "../api";
import { registerEditorGuard } from "../composables/editorGuard";
const props = defineProps<{ days: string[] }>(),
  emit = defineEmits<{ close: []; created: [string] }>(),
  s = useTodo(),
  title = ref(""),
  start = ref(props.days[0] || s.date),
  end = ref(props.days.length > 1 ? props.days.at(-1)! : ""),
  error = ref(""),
  host = ref<HTMLElement>();
let saved = false;
const dirty = () => !saved && !!title.value.trim();
const leave = async () =>
  !s.saving && (!dirty() || (await confirmAction("放弃未保存的任务？")));
const release = registerEditorGuard(dirty, leave);
onUnmounted(release);
onMounted(() => host.value?.querySelector("input")?.focus());
async function cancel() {
  if (await leave()) emit("close");
}
async function save() {
  if (
    !title.value.trim() ||
    !start.value ||
    (end.value && end.value < start.value)
  ) {
    error.value = "请填写标题和有效的日期范围";
    return;
  }
  const a = new Date(start.value + "T00:00:00").getTime(),
    b = end.value ? new Date(end.value + "T23:59:59").getTime() : null;
  const out = await s.run(() =>
    api("Create", {
      commandId: crypto.randomUUID(),
      title: title.value.trim(),
      date: start.value,
      color: "none",
      postpone: "inherit",
      span: { start: a, end: b },
      repeatRule: {
        frequency: "daily",
        interval: 1,
        until: end.value || null,
        weekdays: [],
      },
    }),
  );
  if (out) {
    saved = true;
    emit("created", start.value);
  }
}
</script>
<template>
  <section ref="host" class="embedded-task-editor" aria-label="创建范围任务">
    <header class="dialog-header"><strong>新建任务</strong></header>
    <div class="dialog-body">
      <label class="field"
        ><span>任务标题</span
        ><input v-model="title" aria-label="任务标题" maxlength="500" /></label
      ><label class="field"
        ><span>开始日期</span
        ><input v-model="start" type="date" aria-label="开始日期" /></label
      ><label class="field"
        ><span>结束日期（可选）</span
        ><input v-model="end" type="date" aria-label="结束日期" :min="start"
      /></label>
      <p class="muted small">范围内每天独立完成；留空结束日期则持续显示。</p>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
    </div>
    <footer class="dialog-footer">
      <button class="button" @click="cancel">取消</button
      ><button
        class="button primary"
        :disabled="s.saving || !title.trim()"
        @click="save"
      >
        创建任务
      </button>
    </footer>
  </section>
</template>
