<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import { Search, Plus } from "lucide-vue-next";
import { api } from "../api";
import { useTodo } from "../stores/todo";
import type { Task, Tag } from "../api/types";
import { tagStyle } from "../composables/tags";
import { cancelMotion, reveal } from "../composables/motion";
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
    !trigger()?.contains(e.target as Node)
  )
    emit("close");
}
function trigger() {
  return document.querySelector<HTMLButtonElement>(
    `[data-tag-trigger="${CSS.escape(props.task.id)}"]`,
  );
}
function position() {
  const panel = box.value,
    anchor = trigger();
  if (!panel || !anchor) return;
  const r = anchor.getBoundingClientRect(),
    width = Math.min(240, innerWidth - 16),
    below = innerHeight - r.bottom - 14,
    above = r.top - 14,
    down = below >= 220 || below >= above;
  Object.assign(panel.style, {
    width: `${width}px`,
    maxHeight: `${Math.max(0, Math.min(360, down ? below : above))}px`,
    left: `${Math.max(8, Math.min(innerWidth - width - 8, r.right - width))}px`,
  });
  const height = panel.getBoundingClientRect().height;
  panel.style.top = `${Math.max(8, down ? r.bottom + 6 : r.top - height - 6)}px`;
}
function scroll(event: Event) {
  // Scrolling the options must keep the picker open; moving its anchor closes it.
  if (event.target instanceof Node && box.value?.contains(event.target)) return;
  emit("close");
}
watch([search, () => tags.value.length, error], position, { flush: "post" });
function escape(e: KeyboardEvent) {
  if (e.key !== "Escape" || e.isComposing) return;
  e.preventDefault();
  e.stopPropagation();
  emit("close");
  trigger()?.focus();
}
onMounted(() => {
  box.value?.showPopover();
  position();
  if (box.value) reveal(box.value, "popover");
  document.addEventListener("pointerdown", outside);
  document.addEventListener("keydown", escape, true);
  window.addEventListener("resize", position);
  window.addEventListener("scroll", scroll, true);
  box.value?.querySelector("input")?.focus();
});
onUnmounted(() => {
  if (box.value) cancelMotion(box.value);
  document.removeEventListener("pointerdown", outside);
  document.removeEventListener("keydown", escape, true);
  window.removeEventListener("resize", position);
  window.removeEventListener("scroll", scroll, true);
});
</script>
<template>
  <Teleport to="body">
    <div
      ref="box"
      popover="manual"
      class="tag-popover filter-popover"
      role="dialog"
      aria-label="选择任务标签"
      @pointerdown.stop
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
          /><span class="tag-chip" :style="tagStyle(tag)">{{
            tag.name
          }}</span></label
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
  </Teleport>
</template>
