<script setup lang="ts">
import {
  ref,
  useSlots,
  useAttrs,
  useId,
  onMounted,
  onUnmounted,
  nextTick,
  watch,
  type VNode,
} from "vue";
import { reveal, leave, cancelMotion } from "../composables/motion";
import { ChevronDown, Check, Search } from "lucide-vue-next";
defineOptions({ inheritAttrs: false });
const props = defineProps<{
  modelValue?: string | number;
  disabled?: boolean;
  modelModifiers?: Record<string, boolean>;
  searchable?: boolean;
  searchPlaceholder?: string;
}>();
const emit = defineEmits(["update:modelValue", "change"]);
const slots = useSlots(),
  attrs = useAttrs(),
  id = useId();
const host = ref<HTMLElement>(),
  trigger = ref<HTMLButtonElement>(),
  popup = ref<HTMLElement>();
const open = ref(false),
  active = ref(0),
  search = ref(""),
  composing = ref(false);
type Item = { value: string | number; label: string; disabled: boolean };
function content(nodes: any): string {
  return typeof nodes === "string"
    ? nodes
    : Array.isArray(nodes)
      ? nodes.map((n) => content(n.children)).join("")
      : "";
}
function items(): Item[] {
  const out: Item[] = [];
  const walk = (nodes: VNode[]) =>
    nodes.forEach((n) => {
      if (n.type === "option")
        out.push({
          value: n.props?.value ?? content(n.children).trim(),
          label: content(n.children).trim(),
          disabled: n.props?.disabled === true || n.props?.disabled === "",
        });
      else if (Array.isArray(n.children)) walk(n.children as VNode[]);
    });
  walk(slots.default?.() || []);
  return out;
}
function label() {
  return (
    items().find((i) => String(i.value) === String(props.modelValue))?.label ||
    "请选择"
  );
}
function visibleItems() {
  const query = search.value.trim().toLocaleLowerCase();
  return props.searchable && query
    ? items().filter((item) => item.label.toLocaleLowerCase().includes(query))
    : items();
}
watch(
  search,
  () => {
    active.value = Math.max(
      0,
      visibleItems().findIndex((item) => !item.disabled),
    );
    popup.value?.querySelector(".select-results")?.scrollTo({ top: 0 });
  },
  { flush: "sync" },
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
  if (props.disabled) return;
  if (open.value) return close();
  search.value = "";
  open.value = true;
  active.value = Math.max(
    0,
    visibleItems().findIndex(
      (i) => String(i.value) === String(props.modelValue),
    ),
  );
  await nextTick();
  const rect = trigger.value!.getBoundingClientRect(),
    p = popup.value!;
  const width = Math.min(Math.max(rect.width, 190), innerWidth - 20),
    height = Math.min(
      props.searchable ? 320 : 280,
      items().length * 32 + (props.searchable ? 50 : 8),
    );
  Object.assign(p.style, {
    width: width + "px",
    left: Math.min(rect.left, innerWidth - width - 10) + "px",
    top:
      (rect.bottom + height + 8 < innerHeight
        ? rect.bottom + 5
        : Math.max(8, rect.top - height - 5)) + "px",
  });
  cancelMotion(p);
  p.inert = false;
  p.showPopover();
  const selected = p.querySelector<HTMLElement>('[aria-selected="true"]');
  if (selected) {
    const scroller = props.searchable
      ? p.querySelector<HTMLElement>(".select-results")!
      : p;
    scroller.scrollTop = Math.max(
      0,
      selected.offsetTop - (scroller.clientHeight - selected.offsetHeight) / 2,
    );
  }
  reveal(p, "popover");
  // Safari does not focus buttons on mouse/AX activation by default.
  if (props.searchable) p.querySelector<HTMLInputElement>("input")?.focus();
  else trigger.value?.focus();
}
function choose(i: Item) {
  if (i.disabled) return;
  emit(
    "update:modelValue",
    props.modelModifiers?.number ? Number(i.value) : i.value,
  );
  emit("change", { target: { value: i.value } });
  close(true);
}
async function key(e: KeyboardEvent) {
  if (e.isComposing) return;
  if (
    ["ArrowDown", "ArrowUp", "Home", "End", "Enter", " ", "Escape"].includes(
      e.key,
    )
  )
    e.preventDefault();
  if (e.key === "Escape") {
    e.stopPropagation();
    close(true);
    return;
  }
  if (e.key === "Tab") return close();
  if (!open.value && ["ArrowDown", "ArrowUp", "Enter", " "].includes(e.key))
    return toggle();
  const rows = visibleItems();
  if (!rows.length) return;
  if (e.key === "Enter" || e.key === " ") {
    if (rows[active.value]) choose(rows[active.value]);
  } else if (["ArrowDown", "ArrowUp", "Home", "End"].includes(e.key)) {
    let n =
      e.key === "Home"
        ? 0
        : e.key === "End"
          ? rows.length - 1
          : (active.value + (e.key === "ArrowDown" ? 1 : -1) + rows.length) %
            rows.length;
    for (let i = 0; i < rows.length && rows[n]?.disabled; i++)
      n = (n + (e.key === "ArrowUp" ? -1 : 1) + rows.length) % rows.length;
    active.value = n;
    await nextTick();
    popup.value
      ?.querySelector<HTMLElement>('[data-active="true"]')
      ?.scrollIntoView({ block: "nearest" });
  }
}
function searchKey(e: KeyboardEvent) {
  if (e.isComposing || composing.value) return;
  if (["ArrowDown", "ArrowUp", "Enter", "Escape", "Tab"].includes(e.key)) {
    e.stopPropagation();
    void key(e);
  }
}
function outside(e: Event) {
  if (!(e.target instanceof Node) || !host.value?.contains(e.target)) close();
}
onMounted(() => {
  document.addEventListener("pointerdown", outside);
  window.addEventListener("resize", outside);
});
onUnmounted(() => {
  if (popup.value) {
    cancelMotion(popup.value);
    if (popup.value.matches(":popover-open")) popup.value.hidePopover();
  }
  document.removeEventListener("pointerdown", outside);
  window.removeEventListener("resize", outside);
});
</script>
<template>
  <div ref="host" class="custom-select" :class="attrs.class">
    <button
      ref="trigger"
      type="button"
      class="select-trigger"
      role="combobox"
      :aria-label="String(attrs['aria-label'] || label())"
      :aria-expanded="open"
      :aria-haspopup="searchable ? 'dialog' : 'listbox'"
      :aria-controls="id"
      :aria-activedescendant="
        open && !searchable ? id + '-' + active : undefined
      "
      :disabled="disabled"
      @click="toggle"
      @keydown.stop="key"
    >
      <span>{{ label() }}</span
      ><ChevronDown :size="13" />
    </button>
    <div
      :id="id"
      ref="popup"
      popover="manual"
      :role="searchable ? 'dialog' : 'listbox'"
      class="select-options"
      :class="{ searchable }"
      :aria-label="String(attrs['aria-label'] || '选项')"
    >
      <label v-if="searchable" class="select-search">
        <Search :size="13" />
        <input
          v-model="search"
          role="combobox"
          :aria-label="searchPlaceholder || '搜索选项'"
          :placeholder="searchPlaceholder || '搜索选项'"
          :aria-controls="id + '-results'"
          aria-autocomplete="list"
          :aria-expanded="open"
          :aria-activedescendant="
            visibleItems().length ? id + '-' + active : undefined
          "
          @keydown="searchKey"
          @compositionstart="composing = true"
          @compositionend="composing = false"
        />
      </label>
      <div
        :id="searchable ? id + '-results' : undefined"
        :role="searchable ? 'listbox' : undefined"
        :aria-label="
          searchable ? String(attrs['aria-label'] || '选项') : undefined
        "
        :class="{ 'select-results': searchable }"
      >
        <button
          v-for="(item, i) in visibleItems()"
          :id="id + '-' + i"
          :key="String(item.value)"
          type="button"
          role="option"
          :aria-selected="String(item.value) === String(modelValue)"
          :disabled="item.disabled"
          :data-active="active === i"
          @pointermove="active = i"
          @click="choose(item)"
        >
          <span>{{ item.label }}</span
          ><Check v-if="String(item.value) === String(modelValue)" :size="13" />
        </button>
        <p
          v-if="searchable && !visibleItems().length"
          class="select-empty"
          role="status"
        >
          没有匹配的任务
        </p>
      </div>
    </div>
  </div>
</template>
