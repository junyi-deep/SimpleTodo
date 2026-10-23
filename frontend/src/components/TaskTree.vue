<script setup lang="ts">
import { startTaskDrag, cancelTaskDrag } from "../composables/taskDrag";
onUnmounted(cancelTaskDrag);
import { reveal, leave, cancelMotion } from "../composables/motion";
import { computed, ref, nextTick, onMounted, onUnmounted, watch } from "vue";
import {
  ChevronRight,
  GripVertical,
  CornerDownRight,
  Repeat2,
  Bell,
  CalendarDays,
  Trash2,
  Pencil,
  Check,
} from "lucide-vue-next";
import InlineTitle from "./InlineTitle.vue";
import { canLeaveEditor } from "../composables/editorGuard";
import TaskEditor from "./TaskEditor.vue";
import ReminderPopover from "./ReminderPopover.vue";
import TaskPlacement from "./TaskPlacement.vue";
import { api, confirmAction } from "../api";
import { type Task } from "../api/types";
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
  placement = ref<Task | null>(null),
  reminding = ref(false),
  dropHint = ref<string | null>(null),
  draggedId = ref<string | null>(null),
  rowFocus = ref(props.root.id),
  composing = ref(false),
  deletePending = ref<string | null>(null);
let deleteTimer: ReturnType<typeof setTimeout>;
watch(placement, () => {
  if (s.key === "mini") void api("SetEdgeModalBusy", !!placement.value);
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
  walk(props.root, 0);
  return out;
});
function toggle(id: string) {
  const next = new Set(expanded.value);
  if (next.has(id)) next.delete(id);
  else next.add(id);
  expanded.value = next;
}
async function titleClick(t: Task) {
  if (s.titleEditing === t.id || !(await canLeaveEditor())) return;
  s.activeTask = t.id;
  s.titleEditing = t.id;
}
function resetDelete() {
  clearTimeout(deleteTimer);
  deletePending.value = null;
}
function outside(e: PointerEvent) {
  if (
    !(e.target as HTMLElement).closest(`[data-delete-owner="${props.root.id}"]`)
  )
    resetDelete();
}
async function del(t: Task) {
  if (!(await canLeaveEditor())) return;
  if (deletePending.value !== t.id) {
    resetDelete();
    deletePending.value = t.id;
    deleteTimer = setTimeout(resetDelete, 4000);
    return;
  }
  const latest = s.taskMap.get(t.id);
  if (latest) await s.remove(latest);
  resetDelete();
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
  resetDelete();
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
  if (ancestors.length >= 5) {
    s.inform("任务最多支持 5 层（包含根任务）");
    return;
  }
  expanded.value = new Set([...expanded.value, ...ancestors]);
  adding.value = t.id;
  draft.value = "";
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
}
async function keydown(e: KeyboardEvent, t: Task) {
  if (e.key === "Escape" && deletePending.value) {
    e.preventDefault();
    e.stopPropagation();
    resetDelete();
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
    await titleClick(t);
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
function dropPosition(e: DragEvent) {
  const r = (e.currentTarget as HTMLElement).getBoundingClientRect();
  const y = e.clientY - r.top;
  return e.altKey || y < r.height * 0.25
    ? "before"
    : y > r.height * 0.75
      ? "after"
      : "inside";
}
async function dropNode(e: DragEvent, t: Task) {
  e.preventDefault();
  dropHint.value = null;
  const source = s.taskMap.get(e.dataTransfer?.getData("text/plain") || "");
  if (!source || source.id === t.id) return;
  const mode = dropPosition(e);
  if (mode !== "inside" && !source.parentId && !t.parentId) {
    emit("drop", t, e);
    return;
  }
  if (mode !== "inside" && source.parentId && source.rootId === t.rootId) {
    const parent = t.parentId || t.rootId;
    const peers = children(parent).filter((node) => node.id !== source.id);
    const beforeId = t.parentId
      ? mode === "before"
        ? t.id
        : peers[peers.findIndex((node) => node.id === t.id) + 1]?.id || ""
      : mode === "before"
        ? peers[0]?.id || ""
        : "";
    await s.run(() =>
      api("Reparent", {
        ...s.command(source),
        parentId: t.parentId || t.rootId,
        beforeId,
      }),
    );
    return;
  }
  if (
    !source.parentId &&
    (source.remindAt || source.seriesId || source.description) &&
    !(await confirmAction(
      "作为子任务后将继承目标任务的日期与提醒。原描述会追加到目标根任务的描述；重复任务仅移动当前实例。继续？",
    ))
  )
    return;
  const parentId = mode !== "inside" && t.parentId ? t.parentId : t.id;
  const peers = children(parentId).filter((node) => node.id !== source.id);
  const beforeId =
    mode !== "inside" && t.parentId
      ? mode === "before"
        ? t.id
        : peers[peers.findIndex((node) => node.id === t.id) + 1]?.id || ""
      : "";
  const result = await s.run(() =>
    api(
      "Nest",
      { ...s.command(source), parentId, beforeId },
      s.snapshot.revision,
    ),
  );
  if (result) expanded.value = new Set([...expanded.value, parentId]);
}
</script>
<template>
  <section class="task-tree" :class="{ compact }" :aria-label="root.title">
    <TransitionGroup
      tag="div"
      role="list"
      :aria-label="root.title"
      :css="false"
      @enter="(el, done) => reveal(el, 'row', done)"
      @leave="(el, done) => leave(el, 'row', done)"
      @enter-cancelled="cancelMotion"
      @leave-cancelled="cancelMotion"
      move-class="task-move"
    >
      <div v-for="{ task: t, depth } in rows" :key="t.id" role="listitem">
        <div
          class="task-row group"
          :class="{
            done: t.status === 'completed',
            root: !t.parentId,
            'is-dragging': draggedId === t.id,
            'drop-target': dropHint === t.id,
            selected: s.activeTask === t.id,
          }"
          :style="{ '--depth': Math.min(depth, 5) }"
          role="group"
          :aria-label="`${t.title}，第 ${depth + 1} 级任务`"
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
          @focus="
            rowFocus = t.id;
            s.activeTask = t.id;
          "
          @click="
            ($event.target as HTMLElement).closest(
              'button,input,textarea,.drag-handle',
            ) || titleClick(t)
          "
          @dragover.prevent="dropHint = t.id"
          @drop.stop="dropNode($event, t)"
        >
          <span
            class="drag-handle"
            :draggable="!s.saving && s.sortMode === 'manual'"
            @dragstart="startDrag($event, t)"
            @pointerdown="
              !s.saving &&
              !editing &&
              s.sortMode === 'manual' &&
              startTaskDrag($event, t.id)
            "
            title="拖动；子任务放到节点下，按住 Alt 可同级排序"
            ><GripVertical :size="13"
          /></span>
          <button
            v-if="children(t.id).length"
            class="tree-toggle"
            :aria-label="expanded.has(t.id) ? '折叠子任务' : '展开子任务'"
            :aria-expanded="expanded.has(t.id)"
            @click="toggle(t.id)"
          >
            <ChevronRight
              :size="13"
              :class="{ rotate: expanded.has(t.id) }"
            /></button
          ><span v-else class="tree-toggle spacer" />
          <button
            v-complete="t.status === 'completed'"
            class="task-check"
            :class="{
              checked: t.status === 'completed',
              partial: t.status !== 'completed' && progress(t.id).done > 0,
            }"
            :aria-label="
              t.status === 'completed' ? '取消完成' : '完成任务及所有后代'
            "
            :disabled="s.saving"
            :aria-pressed="t.status === 'completed'"
            @click="s.complete(t)"
          >
            <svg v-if="t.status === 'completed'" viewBox="0 0 16 16">
              <path d="m3 8 3 3 7-7" /></svg
            ><span v-else-if="progress(t.id).done > 0" />
          </button>
          <div class="task-copy">
            <InlineTitle
              v-if="s.titleEditing === t.id"
              :task="t"
              @done="s.titleEditing = ''"
            />
            <button
              v-else
              class="task-title"
              :title="t.title"
              @click="titleClick(t)"
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
          <div
            class="task-actions"
            :class="{ 'confirming-delete': deletePending === t.id }"
          >
            <button
              v-if="!t.parentId"
              class="icon-button reminder-trigger action-reminder"
              :data-reminder-trigger="t.id"
              aria-label="设置提醒"
              title="设置提醒"
              @click.stop="reminding = !reminding"
            >
              <Bell :size="14" />
            </button>
            <button
              class="icon-button task-settings-trigger action-edit"
              aria-label="任务设置"
              title="任务设置"
              @click.stop="emit('edit', t)"
            >
              <Pencil :size="14" />
            </button>
            <button
              v-if="t.parentId || !t.seriesId"
              class="icon-button action-date"
              :aria-label="t.parentId ? '移动到其他父任务' : '调整计划日期'"
              :title="t.parentId ? '移动到其他父任务' : '调整计划日期'"
              @click.stop="placement = t"
            >
              <CalendarDays :size="14" />
            </button>
            <button
              class="icon-button action-delete"
              :class="{ armed: deletePending === t.id }"
              :data-delete-owner="root.id"
              :aria-label="deletePending === t.id ? '确认删除任务' : '删除任务'"
              :title="
                deletePending === t.id
                  ? '再次点击确认删除整个分支；Esc 取消'
                  : '删除任务及其子任务'
              "
              :disabled="s.saving"
              @click.stop="del(t)"
            >
              <Check v-if="deletePending === t.id" :size="15" /><Trash2
                v-else
                :size="14"
              />
            </button>
          </div>
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
          :style="{ paddingLeft: `${Math.min(depth + 1, 5) * 22 + 31}px` }"
        >
          <CornerDownRight :size="14" /><input
            v-model="draft"
            v-sparks
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
      </div>
    </TransitionGroup>
  </section>
  <ReminderPopover
    v-if="reminding"
    :task="s.taskMap.get(root.id) || root"
    @close="reminding = false"
  />
  <TaskPlacement v-if="placement" :task="placement" @close="placement = null" />
</template>
