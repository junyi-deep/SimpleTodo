<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { ArrowUp, ArrowDown, Trash2, Plus } from "lucide-vue-next";
import { registerEditorGuard } from "../composables/editorGuard";
import { api, confirmAction } from "../api";
import { useTodo } from "../stores/todo";
import type { Tag } from "../api/types";
import CustomSelect from "./CustomSelect.vue";
import { tagColors, tagStyle } from "../composables/tags";
const s = useTodo(),
  tags = ref<Tag[]>([]),
  name = ref(""),
  busy = ref(false),
  error = ref("");
let initial = "";
let loaded = false;
const state = () => JSON.stringify([tags.value, name.value]);
const release = registerEditorGuard(
  () => loaded && state() !== initial,
  async () =>
    !busy.value &&
    (!loaded ||
      state() === initial ||
      (await confirmAction("放弃未保存的标签设置？"))),
);
onUnmounted(release);
onMounted(load);
async function load() {
  tags.value = await api("Tags");
  initial = state();
  loaded = true;
}
async function save() {
  if (busy.value) return;
  busy.value = true;
  error.value = "";
  try {
    await api("SaveTags", tags.value);
    await s.metadata();
    initial = state();
    s.inform("标签已保存");
  } catch (e) {
    error.value = String(e);
  } finally {
    busy.value = false;
  }
}
function shift(i: number, d: number) {
  const a = [...tags.value];
  [a[i], a[i + d]] = [a[i + d], a[i]];
  tags.value = a;
}
async function remove(i: number) {
  if (await confirmAction("删除此标签？任务内容会保留。")) {
    tags.value.splice(i, 1);
    await save();
  }
}
async function add() {
  if (!name.value.trim()) return;
  tags.value.push({
    id: crypto.randomUUID(),
    name: name.value.trim(),
    order: tags.value.length,
    color: tagColors[0].value,
  });
  name.value = "";
  await save();
}
</script>
<template>
  <section class="tag-settings">
    <h2>任务标签</h2>
    <p class="muted small">调整标签名称、颜色和显示顺序。</p>
    <div v-for="(tag, i) in tags" :key="tag.id" class="tag-setting-row">
      <span
        class="tag-color-preview"
        :style="tagStyle(tag)"
        aria-hidden="true"
      />
      <input
        v-model="tag.name"
        :disabled="busy"
        :aria-label="'标签名称 ' + (i + 1)"
        maxlength="40"
      /><CustomSelect
        v-model="tag.color"
        :aria-label="'标签颜色 ' + (i + 1)"
        :disabled="busy"
        ><option
          v-for="color in tagColors"
          :key="color.value"
          :value="color.value"
        >
          {{ color.name }}
        </option></CustomSelect
      ><button
        class="icon-button"
        aria-label="上移标签"
        :disabled="!i || busy"
        @click="shift(i, -1)"
      >
        <ArrowUp :size="14" /></button
      ><button
        class="icon-button"
        aria-label="下移标签"
        :disabled="i === tags.length - 1 || busy"
        @click="shift(i, 1)"
      >
        <ArrowDown :size="14" /></button
      ><button
        class="icon-button"
        aria-label="删除标签"
        :disabled="busy"
        @click="remove(i)"
      >
        <Trash2 :size="14" />
      </button>
    </div>
    <div class="inline-actions">
      <input
        v-model="name"
        :disabled="busy"
        aria-label="新标签名称"
        placeholder="新标签"
        maxlength="40"
        @keydown.enter.prevent="add"
      /><button
        class="icon-button"
        :disabled="busy || !name.trim()"
        aria-label="新增标签"
        @click="add"
      >
        <Plus :size="16" />
      </button>
    </div>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <button class="button primary" :disabled="busy" @click="save">
      保存标签与顺序
    </button>
  </section>
</template>
