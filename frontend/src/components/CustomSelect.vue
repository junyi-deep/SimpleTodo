<script setup lang="ts">
import {
  ref,
  useSlots,
  useAttrs,
  useId,
  onMounted,
  onUnmounted,
  nextTick,
  type VNode,
} from "vue";
import { reveal, leave, cancelMotion } from "../composables/motion";
import { ChevronDown, Check } from "lucide-vue-next";
defineOptions({ inheritAttrs: false });
const props = defineProps<{
  modelValue?: string | number;
  disabled?: boolean;
  modelModifiers?: Record<string, boolean>;
}>();
const emit = defineEmits(["update:modelValue", "change"]);
const slots = useSlots(),
  attrs = useAttrs(),
  id = useId();
const host = ref<HTMLElement>(),
  trigger = ref<HTMLButtonElement>(),
  popup = ref<HTMLElement>();
const open = ref(false),
  active = ref(0);
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
  open.value = true;
  active.value = Math.max(
    0,
    items().findIndex((i) => String(i.value) === String(props.modelValue)),
  );
  await nextTick();
  const rect = trigger.value!.getBoundingClientRect(),
    p = popup.value!;
  const width = Math.min(Math.max(rect.width, 190), innerWidth - 20),
    height = Math.min(280, items().length * 32 + 8);
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
  if (selected)
    p.scrollTop = Math.max(
      0,
      selected.offsetTop - (p.clientHeight - selected.offsetHeight) / 2,
    );
  reveal(p, "popover");
  // Safari does not focus buttons on mouse/AX activation by default.
  trigger.value?.focus();
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
  const rows = items();
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
function outside(e: Event) {
  if (!host.value?.contains(e.target as Node)) close();
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
      aria-haspopup="listbox"
      :aria-controls="id"
      :aria-activedescendant="open ? id + '-' + active : undefined"
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
      role="listbox"
      class="select-options"
      :aria-label="String(attrs['aria-label'] || '选项')"
    >
      <button
        v-for="(item, i) in items()"
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
    </div>
  </div>
</template>
