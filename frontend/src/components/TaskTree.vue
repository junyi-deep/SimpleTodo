<script setup lang="ts">
import { computed, ref, nextTick, onMounted, onUnmounted, watch } from "vue";
import {
  ChevronRight,
  Plus,
  MoreHorizontal,
  GripVertical,
  CornerDownRight,
  Repeat2,
  Bell,
  CalendarDays,
  Trash2,
  IndentIncrease,
  IndentDecrease,
  Focus,
} from "lucide-vue-next";
import TaskEditor from "./TaskEditor.vue";
import ReminderPopover from "./ReminderPopover.vue";
import TaskPlacement from "./TaskPlacement.vue";
import { api, confirmAction } from "../api";
import { type Task, parseDay } from "../api/types";
import { useTodo } from "../stores/todo";
const props = defineProps<{
  root: Task;
  compact?: boolean;
  editing?: string;
}>();
const emit = defineEmits<{
  edit: [Task | null];
  drag: [Task];
  drop: [Task, DragEvent];
}>();
const s = useTodo(),
  expanded = ref(new Set<string>()),
  adding = ref<string | null>(null),
  draft = ref(""),
  menu = ref<string | null>(null),
  placement = ref<Task | null>(null),
  reminding = ref(false),
  dropHint = ref<string | null>(null),
  draggedId = ref<string | null>(null),
  focused = ref<string | null>(null),
  rowFocus = ref(props.root.id),
  composing = ref(false),
  menuPosition = ref({ left: 0, top: 0 });
watch(placement, () => {
  if (s.key === "mini") void api("SetEdgeModalBusy", !!placement.value);
});
watch(menu, () => {
  if (s.key === "mini") void api("SetEdgeMenuBusy", menu.value !== null);
});
const nodes = computed(() =>
    s.snapshot.tasks.filter((t) => t.rootId === props.root.id),
  ),
  index = computed(() => new Map(nodes.value.map((t) => [t.id, t])));
const children = (id: string) =>
  nodes.value
    .filter((t) => t.parentId === id)
    .sort((a, b) => {
      if (s.boot?.settings.completedBottom === "true" && a.status !== b.status)
        return a.status === "completed" ? 1 : -1;
      return a.order - b.order || a.id.localeCompare(b.id);
    });
const progress = (id: string) => {
  let done = 0,
    total = 0;
  const visit = (key: string) => {
    const cs = children(key);
    if (!cs.length) {
      total++;
      if (index.value.get(key)?.status === "completed") done++;
    } else cs.forEach((c) => visit(c.id));
  };
  visit(id);
  return { done, total };
};
const rows = computed(() => {
  const out: { task: Task; depth: number }[] = [];
  const walk = (t: Task, depth: number) => {
    out.push({ task: t, depth });
    if (expanded.value.has(t.id))
      children(t.id)
        .filter((c) => !s.hideCompleted || c.status !== "completed")
        .forEach((c) => walk(c, depth + 1));
  };
  walk(
    focused.value ? index.value.get(focused.value) || props.root : props.root,
    0,
  );
  return out;
});
const breadcrumbs = computed(() => {
  const path: Task[] = [];
  let task = focused.value ? index.value.get(focused.value) : undefined;
  while (task) {
    path.unshift(task);
    task = task.parentId ? index.value.get(task.parentId) : undefined;
  }
  return path;
});
function toggle(id: string) {
  const next = new Set(expanded.value);
  if (next.has(id)) next.delete(id);
  else next.add(id);
  expanded.value = next;
}
let titleTimer: ReturnType<typeof setTimeout> | undefined;
function titleClick(t: Task) {
  if (s.boot?.settings.titleClick === "complete") {
    clearTimeout(titleTimer);
    titleTimer = setTimeout(() => void s.complete(t), 250);
  } else emit("edit", t);
}
function titleDoubleClick(t: Task) {
  clearTimeout(titleTimer);
  emit("edit", t);
}
async function moveOrder(t: Task, step: number) {
  const peers = t.parentId
    ? children(t.parentId)
    : s.roots.filter((n) => n.date === t.date);
  const i = peers.findIndex((n) => n.id === t.id),
    j = i + step;
  if (j < 0 || j >= peers.length) return;
  const ids = peers.map((n) => n.id);
  [ids[i], ids[j]] = [ids[j], ids[i]];
  if (t.parentId) {
    const before = ids[ids.indexOf(t.id) + 1] || "";
    await s.run(() =>
      api("Reparent", {
        ...s.command(t),
        parentId: t.parentId,
        beforeId: before,
      }),
    );
  } else await s.run(() => api("Reorder", ids, s.snapshot.revision));
  menu.value = null;
}
function openMenu(t: Task, e: MouseEvent) {
  if (menu.value === t.id) {
    menu.value = null;
    return;
  }
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
  menuPosition.value = {
    left: Math.max(8, Math.min(window.innerWidth - 220, rect.right - 210)),
    top: Math.max(
      8,
      Math.min(
        window.innerHeight - Math.min(430, window.innerHeight - 20),
        rect.bottom + 4,
      ),
    ),
  };
  menu.value = t.id;
}
function outside(e: PointerEvent) {
  const target = e.target as HTMLElement;
  if (
    !target.closest(`[data-owner="${props.root.id}"]`) &&
    !target.closest(`[data-popup="${props.root.id}"]`)
  )
    menu.value = null;
}
function addRequested(e: Event) {
  const detail = (e as CustomEvent).detail;
  if (detail.rootId === props.root.id) {
    const t = index.value.get(detail.id);
    if (t) void add(t);
  }
}
onMounted(() => {
  document.addEventListener("pointerdown", outside);
  document.addEventListener("todo:add-child", addRequested);
});
onUnmounted(() => {
  clearTimeout(titleTimer);
  document.removeEventListener("pointerdown", outside);
  document.removeEventListener("todo:add-child", addRequested);
});
async function add(t: Task) {
  const ancestors = [t.id];
  let current = t;
  while (current.parentId) {
    ancestors.push(current.parentId);
    current = index.value.get(current.parentId)!;
    if (!current) break;
  }
  expanded.value = new Set([...expanded.value, ...ancestors]);
  adding.value = t.id;
  draft.value = "";
  menu.value = null;
  await nextTick();
  document
    .querySelector<HTMLInputElement>(`[data-new-child="${t.id}"]`)
    ?.focus();
}
async function saveChild(parent: Task, event: KeyboardEvent) {
  if (event.isComposing || composing.value) return;
  if (event.key === "Escape") {
    event.preventDefault();
    event.stopPropagation();
    adding.value = null;
    await nextTick();
    document.querySelector<HTMLElement>(`[data-node="${parent.id}"]`)?.focus();
    return;
  }
  if (
    event.key !== "Enter" &&
    !(event.key === "Tab" && s.boot?.settings.tabEntry === "true")
  )
    return;
  event.preventDefault();
  if (!draft.value.trim()) {
    adding.value = null;
    return;
  }
  const result = await s.create(draft.value, parent);
  if (result) {
    draft.value = "";
    await nextTick();
    document
      .querySelector<HTMLInputElement>(`[data-new-child="${parent.id}"]`)
      ?.focus();
  }
}
async function del(t: Task) {
  const count = nodes.value.filter((n) => {
    let cur = n;
    while (cur.parentId) {
      if (cur.parentId === t.id) return true;
      cur = index.value.get(cur.parentId)!;
      if (!cur) break;
    }
    return false;
  }).length;
  if (
    await confirmAction(
      `删除「${t.title}」${count ? `及全部 ${count} 个后代` : ""}？此操作无法逐条撤销。`,
    )
  ) {
    await s.remove(t);
  }
  menu.value = null;
}
async function indent(t: Task, out = false) {
  if (!t.parentId) return;
  let parent: string | undefined,
    before = "";
  if (out) {
    const p = index.value.get(t.parentId);
    if (!p?.parentId) {
      s.inform("已是根任务的直属子任务");
      return;
    }
    parent = p.parentId;
    const peers = children(parent),
      i = peers.findIndex((x) => x.id === p.id);
    before = peers[i + 1]?.id || "";
  } else {
    const peers = children(t.parentId),
      i = peers.findIndex((x) => x.id === t.id);
    if (i <= 0) {
      s.inform("前面没有可作为父节点的任务");
      return;
    }
    parent = peers[i - 1].id;
  }
  await s.run(() =>
    api("Reparent", { ...s.command(t), parentId: parent, beforeId: before }),
  );
  if (parent) expanded.value = new Set([...expanded.value, parent]);
  menu.value = null;
}
async function keydown(e: KeyboardEvent, t: Task) {
  if (e.key === "Escape" && menu.value) {
    e.preventDefault();
    e.stopPropagation();
    menu.value = null;
    return;
  }
  if ((e.target as HTMLElement).matches("input,textarea")) return;
  if (
    (e.target as HTMLElement).matches("button") &&
    ((e.key === "Enter" && !e.shiftKey) || e.key === " ")
  )
    return;
  const mod = e.metaKey || e.ctrlKey;
  if (e.isComposing) return;
  if (e.key === "Enter" && e.shiftKey) {
    e.preventDefault();
    await add(t);
  } else if (e.key === "Enter") {
    e.preventDefault();
    emit("edit", t);
  } else if (e.key === " ") {
    e.preventDefault();
    await s.complete(t);
  } else if (e.key === "Backspace" || e.key === "Delete") {
    e.preventDefault();
    await del(t);
  } else if (mod && e.key === "]") {
    e.preventDefault();
    await indent(t);
  } else if (mod && e.key === "[") {
    e.preventDefault();
    await indent(t, true);
  } else if (e.key.startsWith("Arrow")) {
    e.preventDefault();
    const i = rows.value.findIndex((r) => r.task.id === t.id);
    if (e.key === "ArrowRight") {
      if (children(t.id).length && !expanded.value.has(t.id)) toggle(t.id);
      else rowFocus.value = rows.value[i + 1]?.task.id || t.id;
    } else if (e.key === "ArrowLeft") {
      if (expanded.value.has(t.id)) toggle(t.id);
      else rowFocus.value = t.parentId || t.id;
    } else {
      const visible = Array.from(
        document.querySelectorAll<HTMLElement>("[data-node]"),
      );
      const position = visible.findIndex((el) => el.dataset.node === t.id);
      const target = visible[position + (e.key === "ArrowDown" ? 1 : -1)];
      if (target) {
        target.focus();
        return;
      }
    }
    await nextTick();
    document
      .querySelector<HTMLElement>(`[data-node="${rowFocus.value}"]`)
      ?.focus();
  }
}
function startDrag(e: DragEvent, t: Task) {
  if (!e.dataTransfer || s.saving || props.editing) {
    e.preventDefault();
    return;
  }
  e.dataTransfer.effectAllowed = "move";
  e.dataTransfer.setData("text/plain", t.id);
  const row = (e.target as HTMLElement).closest(".task-row") as HTMLElement;
  const ghost = row.cloneNode(true) as HTMLElement;
  ghost.className = "task-row drag-preview";
  ghost.style.width = Math.min(row.clientWidth, 420) + "px";
  ghost.style.position = "fixed";
  ghost.style.left = "-10000px";
  document.body.appendChild(ghost);
  e.dataTransfer.setDragImage(ghost, 28, 18);
  requestAnimationFrame(() => ghost.remove());
  draggedId.value = t.id;
  emit("drag", t);
}
async function dropNode(e: DragEvent, t: Task) {
  e.preventDefault();
  const id = e.dataTransfer?.getData("text/plain"),
    source = s.taskMap.get(id || "");
  if (!source) return;
  if (source.parentId && source.rootId === t.rootId) {
    await s.run(() =>
      api("Reparent", {
        ...s.command(source),
        parentId: e.altKey ? t.parentId || t.rootId : t.id,
        beforeId: e.altKey ? t.id : "",
      }),
    );
    expanded.value = new Set([...expanded.value, t.id]);
  } else emit("drop", t, e);
}
</script>
<template>
  <section
    class="task-tree"
    :class="{ compact }"
    role="tree"
    :aria-label="root.title"
  >
    <nav v-if="focused" class="branch-path" aria-label="任务层级路径">
      <template v-for="(ancestor, i) in breadcrumbs" :key="ancestor.id">
        <ChevronRight v-if="i" :size="12" />
        <button
          class="branch-back"
          :aria-current="ancestor.id === focused ? 'location' : undefined"
          @click="focused = ancestor.parentId ? ancestor.id : null"
        >
          {{ ancestor.title }}
        </button>
      </template>
    </nav>
    <template v-for="{ task: t, depth } in rows" :key="t.id">
      <div
        class="task-row group"
        :class="{
          done: t.status === 'completed',
          root: !t.parentId,
          'is-dragging': draggedId === t.id,
          'drop-target': dropHint === t.id,
        }"
        :style="{ '--depth': Math.min(depth, 5) }"
        role="treeitem"
        :aria-level="depth + 1"
        :aria-setsize="t.parentId ? children(t.parentId).length : 1"
        :aria-posinset="
          t.parentId
            ? children(t.parentId).findIndex((n) => n.id === t.id) + 1
            : 1
        "
        :aria-expanded="children(t.id).length ? expanded.has(t.id) : undefined"
        :aria-checked="
          t.status === 'completed'
            ? true
            : progress(t.id).done > 0
              ? 'mixed'
              : false
        "
        :tabindex="rowFocus === t.id ? 0 : -1"
        :data-node="t.id"
        @dragenter.prevent="dropHint = t.id"
        @dragleave="
          if (
            !($event.currentTarget as HTMLElement).contains(
              $event.relatedTarget as Node,
            )
          )
            dropHint = null;
        "
        @dragend="
          draggedId = null;
          dropHint = null;
        "
        @drop="dropHint = null"
        @keydown="keydown($event, t)"
        @focus="rowFocus = t.id"
        @dragover.prevent
        @drop.stop="dropNode($event, t)"
      >
        <span
          class="drag-handle"
          :draggable="!s.saving && s.sortMode === 'manual'"
          @dragstart="startDrag($event, t)"
          title="拖动；子任务放到节点下，按住 Alt 可同级排序"
          ><GripVertical :size="13"
        /></span>
        <button
          v-if="children(t.id).length"
          class="tree-toggle"
          :aria-label="expanded.has(t.id) ? '折叠子任务' : '展开子任务'"
          @click="toggle(t.id)"
        >
          <ChevronRight
            :size="13"
            :class="{ rotate: expanded.has(t.id) }"
          /></button
        ><span v-else class="tree-toggle spacer" />
        <button
          class="task-check"
          :class="{
            checked: t.status === 'completed',
            partial: t.status !== 'completed' && progress(t.id).done > 0,
          }"
          :aria-label="
            t.status === 'completed' ? '取消完成' : '完成任务及所有后代'
          "
          :disabled="s.saving"
          @click="s.complete(t)"
        >
          <svg v-if="t.status === 'completed'" viewBox="0 0 16 16">
            <path d="m3 8 3 3 7-7" /></svg
          ><span v-else-if="progress(t.id).done > 0" />
        </button>
        <div class="task-copy">
          <button
            class="task-title"
            :title="t.title"
            @click="titleClick(t)"
            @dblclick="titleDoubleClick(t)"
            :disabled="s.saving"
          >
            {{ t.title }}</button
          ><small v-if="!t.parentId && t.remindAt" class="task-reminder"
            ><Bell :size="10" />{{
              new Date(t.remindAt).toLocaleString("zh-CN", {
                month: "2-digit",
                day: "2-digit",
                hour: "2-digit",
                minute: "2-digit",
              })
            }}</small
          >
        </div>
        <span v-if="children(t.id).length" class="task-progress"
          >{{ progress(t.id).done }}/{{ progress(t.id).total }}</span
        >
        <Repeat2
          v-if="t.seriesId"
          class="task-indicator"
          :size="12"
          title="重复任务"
        />
        <button
          v-if="!t.parentId"
          class="icon-button reminder-trigger"
          :class="{ scheduled: !!t.remindAt }"
          :data-reminder-trigger="t.id"
          aria-label="设置提醒"
          @click.stop="reminding = !reminding"
        >
          <Bell :size="14" />
        </button>
        <details
          class="node-more"
          :data-owner="root.id"
          :open="menu === t.id"
          @toggle="
            (e) => {
              if (!(e.target as HTMLDetailsElement).open && menu === t.id)
                menu = null;
            }
          "
        >
          <summary
            class="icon-button"
            aria-label="更多任务操作"
            @click.prevent="openMenu(t, $event)"
          >
            <MoreHorizontal :size="16" />
          </summary>
          <Teleport to="body"
            ><div
              v-if="menu === t.id"
              class="menu node-menu"
              :data-popup="root.id"
              :style="{
                left: menuPosition.left + 'px',
                top: menuPosition.top + 'px',
              }"
              @keydown.esc.stop="menu = null"
            >
              <button v-if="t.parentId" @click="indent(t)">
                <IndentIncrease :size="14" />缩进一级</button
              ><button v-if="t.parentId" @click="indent(t, true)">
                <IndentDecrease :size="14" />提升一级
              </button>
              <button
                v-if="children(t.id).length"
                @click="
                  focused = t.id;
                  expanded = new Set(nodes.map((n) => n.id));
                  menu = null;
                "
              >
                <Focus :size="14" />聚焦此分支</button
              ><button
                v-if="!t.parentId && !t.seriesId"
                @click="
                  s.move(t, t.date ? null : s.date);
                  menu = null;
                "
              >
                <CalendarDays :size="14" />{{
                  t.date ? "移到收集箱" : "安排到所选日期"
                }}
              </button>
              <button
                v-if="t.parentId || !t.seriesId"
                @click="
                  placement = t;
                  menu = null;
                "
              >
                {{ t.parentId ? "移动到其他父任务…" : "调整计划日期…" }}
              </button>
              <hr />
              <button class="danger" @click="del(t)">
                <Trash2 :size="14" />删除{{
                  children(t.id).length ? "整个分支" : ""
                }}
              </button>
            </div></Teleport
          >
        </details>
      </div>
      <TaskEditor
        v-if="editing === t.id"
        :key="t.id"
        :task="t"
        @close="emit('edit', null)"
      />
      <div
        v-if="adding === t.id"
        class="child-entry"
        :style="{ paddingLeft: `${Math.min(depth + 1, 5) * 16 + 31}px` }"
      >
        <CornerDownRight :size="14" /><input
          v-model="draft"
          :data-new-child="t.id"
          placeholder="添加子任务，Enter 连续录入"
          maxlength="500"
          @compositionstart="composing = true"
          @compositionend="composing = false"
          @keydown="saveChild(t, $event)"
        /><button
          class="icon-button"
          @click="adding = null"
          aria-label="取消新增"
        >
          ×
        </button>
      </div>
    </template>
  </section>
  <ReminderPopover
    v-if="reminding"
    :task="s.taskMap.get(root.id) || root"
    @close="reminding = false"
  />
  <TaskPlacement v-if="placement" :task="placement" @close="placement = null" />
</template>
