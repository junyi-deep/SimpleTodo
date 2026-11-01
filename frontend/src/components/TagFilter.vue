<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, useId } from "vue";
import { ChevronDown, Search, Tag as TagIcon } from "lucide-vue-next";
import type { Tag } from "../api/types";
import { cancelMotion, leave, reveal } from "../composables/motion";

const props = defineProps<{ modelValue: string[]; options: Tag[] }>();
const emit = defineEmits<{ "update:modelValue": [string[]] }>();
const host = ref<HTMLElement>(),
  trigger = ref<HTMLButtonElement>(),
  popup = ref<HTMLElement>();
const open = ref(false),
  search = ref(""),
  id = useId();
const matches = computed(() =>
  props.options.filter((tag) =>
    tag.name
      .toLocaleLowerCase()
      .includes(search.value.trim().toLocaleLowerCase()),
  ),
);
const names = computed(() =>
  props.options
    .filter((tag) => props.modelValue.includes(tag.id))
    .map((tag) => tag.name)
    .join("、"),
);
function close(focus = false) {
  if (!open.value) return;
  open.value = false;
  const p = popup.value;
  if (p?.matches(":popover-open")) {
    p.inert = true;
    leave(p, "popover", () => {
      if (!open.value && p.matches(":popover-open")) p.hidePopover();
      p.inert = false;
    });
  }
  if (focus) trigger.value?.focus();
}
async function toggle() {
  if (open.value) return close();
  search.value = "";
  open.value = true;
  await nextTick();
  const p = popup.value!,
    rect = trigger.value!.getBoundingClientRect();
  const width = Math.min(250, innerWidth - 20);
  Object.assign(p.style, {
    width: `${width}px`,
    left: `${Math.max(10, Math.min(rect.right - width, innerWidth - width - 10))}px`,
    top: `${rect.bottom + 290 < innerHeight ? rect.bottom + 5 : Math.max(10, rect.top - Math.min(280, rect.top - 10))}px`,
  });
  cancelMotion(p);
  p.inert = false;
  p.showPopover();
  reveal(p, "popover");
  p.querySelector("input")?.focus();
}
function select(tagId: string) {
  emit(
    "update:modelValue",
    props.modelValue.includes(tagId)
      ? props.modelValue.filter((value) => value !== tagId)
      : [...props.modelValue, tagId],
  );
}
function outside(event: Event) {
  if (!(event.target instanceof Node) || !host.value?.contains(event.target))
    close();
}
onMounted(() => {
  document.addEventListener("pointerdown", outside);
  document.addEventListener("focusin", outside);
  window.addEventListener("resize", outside);
  window.addEventListener("scroll", outside, true);
});
onUnmounted(() => {
  if (popup.value) cancelMotion(popup.value);
  document.removeEventListener("pointerdown", outside);
  document.removeEventListener("focusin", outside);
  window.removeEventListener("resize", outside);
  window.removeEventListener("scroll", outside, true);
});
</script>
<template>
  <div ref="host" class="tag-filter">
    <button
      ref="trigger"
      type="button"
      class="select-trigger"
      aria-label="筛选报告标签"
      aria-haspopup="dialog"
      :aria-controls="id"
      :aria-expanded="open"
      :title="names || '全部标签'"
      @click="toggle"
      @keydown.down.prevent="!open && toggle()"
    >
      <TagIcon :size="13" /><span>{{
        modelValue.length ? `标签 ${modelValue.length}` : "全部标签"
      }}</span
      ><ChevronDown :size="13" />
    </button>
    <div
      :id="id"
      ref="popup"
      popover="manual"
      class="tag-filter-popup"
      role="dialog"
      aria-label="报告标签筛选"
      @keydown.esc.stop.prevent="close(true)"
    >
      <label class="search-control"
        ><Search :size="13" /><input
          v-model="search"
          aria-label="搜索筛选标签"
          placeholder="搜索标签"
      /></label>
      <div class="tag-filter-options">
        <label v-for="tag in matches" :key="tag.id"
          ><input
            type="checkbox"
            :checked="modelValue.includes(tag.id)"
            @change="select(tag.id)"
          /><span>{{ tag.name }}</span></label
        >
        <p v-if="!matches.length" class="muted small">没有匹配的标签</p>
      </div>
      <button
        v-if="modelValue.length"
        type="button"
        class="text-button"
        @click="emit('update:modelValue', [])"
      >
        清除标签筛选
      </button>
    </div>
  </div>
</template>
