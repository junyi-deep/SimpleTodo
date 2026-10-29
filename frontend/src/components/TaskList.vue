<script setup lang="ts">
import { reveal, leave, cancelMotion } from "../composables/motion";
import { ref, nextTick, onMounted, onUnmounted } from "vue";
import { Plus, CornerDownLeft } from "lucide-vue-next";
import InlineTitle from "./InlineTitle.vue";
import { canLeaveEditor } from "../composables/editorGuard";
import TaskTree from "./TaskTree.vue";
import { taskDropProjection } from "../composables/taskDrag";
import { api } from "../api";
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
  drop: [Task, DragEvent, "before" | "after" | "inside", string];
  paste: [string];
}>();
const s = useTodo(),
  adding = ref(false),
  sibling = ref<Task>();
async function requestSibling(event: Event) {
  const t = s.taskMap.get((event as CustomEvent).detail.id);
  if (!t || t.parentId || !props.roots.some((r) => r.id === t.id)) return;
  adding.value = false;
  sibling.value = undefined;
  await nextTick();
  sibling.value = t;
}
async function start() {
  if (await canLeaveEditor()) {
    sibling.value = undefined;
    adding.value = false;
    await nextTick();
    adding.value = true;
    await nextTick();
  }
}
async function dropEnd(event: DragEvent) {
  const projected = taskDropProjection(event);
  if (projected) {
    event.preventDefault();
    const id = event.dataTransfer?.getData("text/plain") || "";
    if (await canLeaveEditor()) await s.dropProjected(id, projected);
    return;
  }
  if (
    event.target !== event.currentTarget &&
    !(event.target as HTMLElement).closest(
      ".list-end-drop,.add-task-button,.compact-empty",
    )
  )
    return;
  event.preventDefault();
  const id = event.dataTransfer?.getData("text/plain") || "";
  if (!(await canLeaveEditor())) return;
  const source = s.taskMap.get(id);
  if (!source || source.status !== "pending") return;
  if (source.parentId)
    await s.run(() =>
      api("Reparent", {
        ...s.command(source),
        parentId: source.parentId,
        beforeId: "",
      }),
    );
  else await s.reorderRoot(source);
}
function requestAdd() {
  void start();
}
onMounted(() => {
  document.addEventListener("todo:new-task", requestAdd);
  document.addEventListener("todo:new-sibling", requestSibling);
});
onUnmounted(() => {
  document.removeEventListener("todo:new-task", requestAdd);
  document.removeEventListener("todo:new-sibling", requestSibling);
});
</script>
<template>
  <section
    v-motion:timeline="date"
    class="task-list"
    :aria-busy="s.loading"
    :data-completed-bottom="
      s.boot?.settings.completedBottom === 'true' ? '' : undefined
    "
    @dragover.prevent
    @drop="dropEnd"
  >
    <TransitionGroup
      :css="false"
      @enter="(el, done) => reveal(el, 'row', done)"
      @leave="(el, done) => leave(el, 'row', done)"
      @enter-cancelled="cancelMotion"
      @leave-cancelled="cancelMotion"
      move-class="task-move"
    >
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
            :peers="roots"
            :editing="editing"
            @edit="emit('edit', $event)"
            @drag="emit('drag', $event)"
            @drop="(t, e, mode, id) => emit('drop', t, e, mode, id)"
          />
        </div>
        <div v-if="sibling?.id === t.id" class="draft-task-row">
          <span class="draft-check" /><InlineTitle
            :key="t.id"
            :after-task="t"
            @done="sibling = undefined"
            @paste="emit('paste', $event)"
          />
        </div>
      </div>
    </TransitionGroup>
    <div class="list-end-drop" aria-hidden="true" />
    <div v-if="adding" v-motion:row class="draft-task-row">
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
      @mousedown.prevent
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
