<script setup lang="ts">
import CustomSelect from "./CustomSelect.vue";
import { restoreDialogFocus } from "../composables/dialogFocus";
import { computed, onMounted, ref } from "vue";
import { X } from "lucide-vue-next";
import { api, confirmAction } from "../api";
import type { Task } from "../api/types";
import { useTodo } from "../stores/todo";
const props = defineProps<{ task: Task }>(),
  emit = defineEmits<{ close: [] }>(),
  s = useTodo(),
  dialog = ref<HTMLElement>();
const parent = ref(props.task.parentId || props.task.rootId),
  date = ref(props.task.date || s.boot?.today || s.date),
  inbox = ref(!props.task.date);
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

async function save() {
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
    if (
      inbox.value &&
      props.task.remindAt &&
      !(await confirmAction("移入收集箱将取消未发送的提醒。继续？"))
    )
      return;
    await s.run(async () => {
      await api("Move", { ...c, date: inbox.value ? null : date.value });
      ok = true;
    });
  }
  if (ok) emit("close");
}
</script>
<template>
  <section
    :inert="s.maintenance || undefined"
    ref="dialog"
    v-motion
    class="dialog embedded-pane"
    @keydown.esc.stop.prevent="emit('close')"
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
          @click="emit('close')"
        >
          <X :size="18" />
        </button>
      </header>
      <div class="dialog-body">
        <p class="muted small">{{ task.title }}</p>
        <label v-if="task.parentId" class="field"
          ><span>新的父任务</span
          ><CustomSelect v-model="parent">
            <option v-for="t in candidates" :key="t.id" :value="t.id">
              {{ path(t) }}
            </option></CustomSelect
          ><small>携带完整子树；不能跨根任务或移入自身后代。</small></label
        ><template v-else
          ><label class="check-label"
            ><input v-model="inbox" type="checkbox" />移到收集箱</label
          ><label v-if="!inbox" class="field"
            ><span>计划日期</span
            ><input
              v-model="date"
              type="date"
              min="2000-01-01"
              max="2100-12-31"
              required /></label
        ></template>
        <p v-if="s.error" class="inline-error" role="alert">{{ s.error }}</p>
        <template v-if="s.error.includes('CONFLICT')"
          ><button type="button" class="button" @click="refresh">
            加载最新任务树
          </button></template
        >
        <div v-if="latest" class="path-output">
          最新根任务：{{ latest.title }}<br />{{ latest.date || "收集箱" }} ·
          当前移动选择仍保留<button type="button" class="button" @click="retry">
            基于最新版本确认移动
          </button>
        </div>
      </div>
      <footer class="dialog-footer">
        <button type="button" class="button" @click="emit('close')">取消</button
        ><button class="button primary" :disabled="s.saving">确认移动</button>
      </footer>
    </form>
  </section>
</template>
