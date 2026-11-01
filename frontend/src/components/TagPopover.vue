<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { Search, Plus } from "lucide-vue-next";
import { api } from "../api";
import { useTodo } from "../stores/todo";
import type { Task, Tag } from "../api/types";
const props = defineProps<{ task: Task }>(),
  emit = defineEmits<{ close: [] }>(),
  s = useTodo(),
  search = ref(""),
  box = ref<HTMLElement>(),
  busy = ref(false),
  error = ref("");
const tags = computed(() =>
  s.tags.filter((t) =>
    t.name.toLocaleLowerCase().includes(search.value.toLocaleLowerCase()),
  ),
);
const current = computed(() => s.taskMap.get(props.task.id) || props.task);
async function toggle(id: string) {
  busy.value = true;
  const ids = current.value.tagIds || [];
  await s.run(() =>
    api(
      "SetTaskTags",
      s.command(current.value),
      ids.includes(id) ? ids.filter((x) => x !== id) : [...ids, id],
    ),
  );
  busy.value = false;
}
async function add() {
  const name = search.value.trim();
  if (!name) return;
  busy.value = true;
  try {
    const latest = await api<Tag[]>("Tags");
    const found = latest.find((t) => t.name === name),
      id = found?.id || crypto.randomUUID();
    if (!found)
      await api("SaveTags", [...latest, { id, name, order: latest.length }]);
    await s.metadata();
    await s.run(() =>
      api("SetTaskTags", s.command(current.value), [
        ...new Set([...(current.value.tagIds || []), id]),
      ]),
    );
    search.value = "";
  } catch (e) {
    error.value = String(e);
  } finally {
    busy.value = false;
  }
}
function outside(e: PointerEvent) {
  if (
    !box.value?.contains(e.target as Node) &&
    !(e.target as HTMLElement).closest("[data-tag-trigger]")
  )
    emit("close");
}
function escape(e: KeyboardEvent) {
  if (e.key !== "Escape" || e.isComposing) return;
  e.preventDefault();
  e.stopPropagation();
  const owner = box.value?.closest("[data-node]");
  emit("close");
  owner?.querySelector<HTMLButtonElement>("[data-tag-trigger]")?.focus();
}
onMounted(() => {
  document.addEventListener("pointerdown", outside);
  document.addEventListener("keydown", escape, true);
  box.value?.querySelector("input")?.focus();
});
onUnmounted(() => {
  document.removeEventListener("pointerdown", outside);
  document.removeEventListener("keydown", escape, true);
});
</script>
<template>
  <div
    ref="box"
    v-motion:popover
    class="tag-popover filter-popover"
    role="dialog"
    aria-label="选择任务标签"
    @click.stop
    @keydown.esc.stop.prevent="emit('close')"
  >
    <label class="search-control"
      ><Search :size="13" /><input
        v-model="search"
        aria-label="搜索或新增标签"
        placeholder="搜索标签"
        @keydown.enter.prevent="add"
    /></label>
    <div class="tag-options">
      <label v-for="tag in tags" :key="tag.id"
        ><input
          type="checkbox"
          :checked="current.tagIds?.includes(tag.id)"
          :disabled="busy || s.saving"
          @change="toggle(tag.id)"
        />{{ tag.name }}</label
      ><small v-if="!tags.length" class="muted">没有匹配的标签</small>
    </div>
    <button
      v-if="search.trim() && !s.tags.some((t) => t.name === search.trim())"
      class="text-button"
      :disabled="busy"
      @click="add"
    >
      <Plus :size="13" />新增“{{ search.trim() }}”
    </button>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
  </div>
</template>
