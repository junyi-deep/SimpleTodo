<script setup lang="ts">
import { ref } from "vue";
import { Plus, CornerDownLeft } from "lucide-vue-next";
import TaskTree from "./TaskTree.vue";
import { useTodo } from "../stores/todo";
import type { Task } from "../api/types";
const props = defineProps<{
  roots: Task[];
  date: string | null;
  editing?: string;
  batch?: boolean;
}>();
const emit = defineEmits<{
  edit: [Task | null];
  drag: [Task];
  drop: [Task, DragEvent];
  paste: [string];
}>();
const s = useTodo(),
  draft = ref(""),
  input = ref<HTMLInputElement>(),
  composing = ref(false);
async function add(e?: KeyboardEvent) {
  if (composing.value || e?.isComposing || s.saving || !draft.value.trim())
    return;
  if (await s.create(draft.value.trim(), undefined, props.date)) {
    draft.value = "";
    input.value?.focus();
  }
}
function paste(e: ClipboardEvent) {
  const text = e.clipboardData?.getData("text/plain") || "";
  if (text.trim().split(/\r?\n/).length > 1) {
    e.preventDefault();
    emit("paste", text);
  }
}
</script>
<template>
  <section class="task-list" :aria-busy="s.loading">
    <div v-for="t in roots" :key="t.id" class="root-entry">
      <div class="selectable-root">
        <input
          v-if="batch"
          class="root-select"
          type="checkbox"
          :aria-label="'选择 ' + t.title"
          :checked="s.selected.includes(t.id)"
          @change="
            s.selected = s.selected.includes(t.id)
              ? s.selected.filter((x) => x !== t.id)
              : [...s.selected, t.id]
          "
        />
        <TaskTree
          :root="t"
          :editing="editing"
          @edit="emit('edit', $event)"
          @drag="emit('drag', $event)"
          @drop="(t, e) => emit('drop', t, e)"
        />
      </div>
    </div>
    <form class="inline-add" @submit.prevent="add()">
      <Plus :size="15" /><input
        ref="input"
        v-model="draft"
        aria-label="新待办标题"
        placeholder="添加任务，按 Enter 保存"
        maxlength="500"
        @paste="paste"
        @compositionstart="composing = true"
        @compositionend="composing = false"
        @keydown.enter.prevent="add($event)"
        @keydown.tab="
          (e) => {
            if (s.boot?.settings.tabEntry === 'true') {
              e.preventDefault();
              add(e);
            }
          }
        "
        @keydown.esc.stop="
          draft = '';
          input?.blur();
        "
      /><button
        aria-label="添加待办"
        class="icon-button"
        :disabled="!draft.trim() || s.saving"
      >
        <CornerDownLeft :size="14" />
      </button>
    </form>
    <p v-if="!roots.length && !s.loading" class="compact-empty">
      还没有任务。先记下一件要做的事。
    </p>
    <p v-if="s.loading" class="compact-empty">正在读取…</p>
    <button
      v-if="s.snapshot.truncated"
      class="text-button"
      @click="
        s.limit = Math.min(10000, s.limit + 500);
        s.load();
      "
    >
      加载更多任务
    </button>
  </section>
</template>
