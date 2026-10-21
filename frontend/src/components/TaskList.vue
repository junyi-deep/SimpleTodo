<script setup lang="ts">
import { ref, nextTick, onMounted, onUnmounted } from "vue";
import { Plus, CornerDownLeft } from "lucide-vue-next";
import InlineTitle from "./InlineTitle.vue";
import { canLeaveEditor } from "../composables/editorGuard";
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
  adding = ref(false);
async function start() {
  if (await canLeaveEditor()) {
    adding.value = true;
    await nextTick();
  }
}
function requestAdd() {
  void start();
}
onMounted(() => document.addEventListener("todo:new-task", requestAdd));
onUnmounted(() => document.removeEventListener("todo:new-task", requestAdd));
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
    <div v-if="adding" class="draft-task-row">
      <span class="draft-check" /><InlineTitle
        :date="date"
        @done="adding = false"
        @paste="emit('paste', $event)"
      />
    </div>
    <button
      v-else
      class="add-task-button"
      aria-label="新增待办"
      title="新增待办"
      @click="start"
    >
      <Plus :size="17" />
    </button>
    <p v-if="!roots.length && !s.loading" class="compact-empty">
      点击加号，记下第一件要做的事。
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
