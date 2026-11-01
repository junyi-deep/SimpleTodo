<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import {
  House,
  NotebookPen,
  Plus,
  Search,
  ArrowLeft,
  Save,
  Eye,
  Pencil,
  Trash2,
} from "lucide-vue-next";
import MarkdownIt from "markdown-it";
import CustomSelect from "./CustomSelect.vue";
import TagFilter from "./TagFilter.vue";
import TaskEditor from "./TaskEditor.vue";
import { api, confirmAction, listen, previewMode } from "../api";
import { useTodo } from "../stores/todo";
import {
  type Task,
  type Note,
  dayString,
  shiftDay,
  parseDay,
} from "../api/types";
import { registerEditorGuard } from "../composables/editorGuard";
const s = useTodo(),
  page = ref("home"),
  from = ref(`${new Date().getFullYear()}-01-01`),
  until = ref(""),
  children = ref(false),
  status = ref("all"),
  tags = ref<string[]>([]),
  search = ref(""),
  tasks = ref<Task[]>([]),
  loading = ref(false),
  error = ref(""),
  editor = ref<Task | null>(null),
  notes = ref<Note[]>([]),
  noteSearch = ref(""),
  draft = ref<Note | null>(null),
  preview = ref(false),
  saving = ref(false),
  noteWidth = ref(33),
  split = ref<HTMLElement>(),
  initialNote = ref("");
watch(
  () => s.boot?.settings.reviewNoteWidth,
  (v) => {
    if (v) noteWidth.value = Math.max(25, Math.min(50, Number(v)));
  },
  { immediate: true },
);
async function editTask(t: Task) {
  try {
    const tree = await api<Task[]>("ReadTree", t.rootId);
    s.snapshot.tasks = [
      ...s.snapshot.tasks.filter((n) => n.rootId !== t.rootId),
      ...tree,
    ];
    editor.value = t;
  } catch (e) {
    error.value = String(e);
  }
}
function noteLink(event: MouseEvent) {
  const a = (event.target as HTMLElement).closest<HTMLAnchorElement>("a[href]");
  if (!a || a.getAttribute("href")?.startsWith("#")) return;
  event.preventDefault();
  const link = a.href;
  if (!/^https?:|^mailto:/i.test(link)) {
    error.value = "仅支持 http、https 和邮件链接";
    return;
  }
  if (previewMode) window.open(link, "_blank", "noopener,noreferrer");
  else void api("OpenExternal", link).catch((e) => (error.value = String(e)));
}
const markdown = new MarkdownIt({ html: false, linkify: false, breaks: true });
const rendered = computed(() => markdown.render(draft.value?.content || ""));
const dirty = () =>
  !!draft.value && JSON.stringify(draft.value) !== initialNote.value;
async function leave() {
  return (
    !saving.value && (!dirty() || (await confirmAction("放弃未保存的笔记？")))
  );
}
const release = registerEditorGuard(dirty, leave);
let stop: (() => void) | undefined;
onMounted(async () => {
  noteWidth.value = Number(s.boot?.settings.reviewNoteWidth || 33);
  await load();
  await loadNotes();
  stop = await listen("todo:changed", () => {
    if (!dirty()) void loadNotes();
    void load();
  });
});
onUnmounted(() => {
  release();
  stop?.();
  stopResize();
});
async function load() {
  loading.value = true;
  error.value = "";
  try {
    tasks.value = await api("TaskReport", {
      view: "all",
      date: "",
      end: "",
      limit: 10000,
    });
  } catch (e) {
    error.value = String(e);
  } finally {
    loading.value = false;
  }
}
const reportMap = computed(() => new Map(tasks.value.map((t) => [t.id, t])));
const report = computed(() =>
  tasks.value.filter((t) => {
    const date =
      t.date ||
      reportMap.value.get(t.rootId)?.date ||
      dayString(new Date(t.startAt || t.createdAt));
    if (
      t.parentId &&
      ((t.startAt && date < dayString(new Date(t.startAt))) ||
        (t.endAt && date > dayString(new Date(t.endAt))))
    )
      return false;
    return (
      (children.value || !t.parentId) &&
      t.status !== "skipped" &&
      (!from.value || date >= from.value) &&
      date <= (until.value || s.boot?.today || dayString(new Date())) &&
      (status.value === "all" || t.status === status.value) &&
      (page.value !== "review" ||
        !tags.value.length ||
        t.tagIds?.some((id) => tags.value.includes(id))) &&
      (page.value !== "review" ||
        !search.value ||
        `${t.title} ${t.description || ""}`
          .toLocaleLowerCase()
          .includes(search.value.toLocaleLowerCase()))
    );
  }),
);
const done = computed(() =>
    report.value.filter((t) => t.status === "completed"),
  ),
  rate = computed(() =>
    report.value.length
      ? Math.round((done.value.length / report.value.length) * 100)
      : 0,
  );
const counts = computed(() => {
  const out = new Map<string, number>();
  for (const t of done.value) {
    const day = dayString(new Date(t.completedAt || t.updatedAt));
    out.set(day, (out.get(day) || 0) + 1);
  }
  return out;
});
const heatDays = computed(() => {
  const end = until.value || s.boot?.today || dayString(new Date());
  const begin = from.value || end.slice(0, 4) + "-01-01";
  const out: string[] = [];
  for (
    let day = begin;
    day <= end && out.length < 36700;
    day = shiftDay(day, 1)
  )
    out.push(day);
  return out;
});
const trend = computed(() => {
  const end = until.value || s.boot?.today || dayString(new Date());
  return Array.from({ length: 14 }, (_, i) => {
    const day = shiftDay(end, i - 13);
    return { day, n: counts.value.get(day) || 0 };
  });
});
const trendMax = computed(() => Math.max(1, ...trend.value.map((d) => d.n)));
const pieSlices = computed(() => {
  let offset = 0;
  return [
    { key: "completed", label: "已完成", count: done.value.length },
    {
      key: "pending",
      label: "未完成",
      count: report.value.length - done.value.length,
    },
  ].map((item) => {
    const start = offset,
      ratio = report.value.length ? item.count / report.value.length : 0;
    offset += ratio;
    const point = (fraction: number) => {
      const angle = fraction * Math.PI * 2 - Math.PI / 2;
      return `${60 + Math.cos(angle) * 56} ${60 + Math.sin(angle) * 56}`;
    };
    return {
      ...item,
      ratio,
      path: `M60 60 L${point(start)} A56 56 0 ${ratio > 0.5 ? 1 : 0} 1 ${point(offset)} Z`,
    };
  });
});
const tagStats = computed(() =>
  s.tags
    .map((tag) => ({
      ...tag,
      n: report.value.filter((task) => task.tagIds?.includes(tag.id)).length,
    }))
    .sort((a, b) => b.n - a.n)
    .slice(0, 8),
);
const heatLevel = (day: string) => Math.min(4, counts.value.get(day) || 0);
async function switchPage(next: string) {
  if (await leave()) {
    page.value = next;
    draft.value = null;
  }
}
async function loadNotes() {
  try {
    notes.value = await api("Notes", noteSearch.value);
  } catch (e) {
    error.value = String(e);
  }
}
async function openNote(note?: Note) {
  if (!(await leave())) return;
  draft.value = note
    ? { ...note }
    : {
        id: "",
        title: "",
        content: "",
        createdAt: 0,
        updatedAt: 0,
        version: 0,
      };
  initialNote.value = JSON.stringify(draft.value);
  preview.value = false;
}
async function backNotes() {
  if (await leave()) {
    draft.value = null;
    await loadNotes();
  }
}
async function deleteNote() {
  const note = draft.value;
  if (!note?.id || saving.value) return;
  if (
    !(await confirmAction(
      `删除笔记“${note.title}”？已保存内容及未保存修改都会移除。`,
    ))
  )
    return;
  saving.value = true;
  error.value = "";
  try {
    await api("DeleteNote", note.id, note.version);
    draft.value = null;
    initialNote.value = "";
    await loadNotes();
    s.inform("笔记已删除");
  } catch (e) {
    error.value = String(e);
  } finally {
    saving.value = false;
  }
}
async function saveNote() {
  if (!draft.value?.title.trim()) {
    error.value = "笔记标题必填";
    return;
  }
  saving.value = true;
  error.value = "";
  try {
    const submitted = { ...draft.value };
    const saved = await api<Note>("SaveNote", submitted);
    const unchanged = JSON.stringify(draft.value) === JSON.stringify(submitted);
    draft.value = unchanged
      ? saved
      : {
          ...draft.value,
          id: saved.id,
          createdAt: saved.createdAt,
          updatedAt: saved.updatedAt,
          version: saved.version,
        };
    initialNote.value = JSON.stringify(saved);
    await loadNotes();
    s.inform("笔记已保存");
  } catch (e) {
    error.value = String(e);
  } finally {
    saving.value = false;
  }
}
let dragStart: { x: number; width: number; ratio: number } | undefined;
function resize(e: PointerEvent) {
  if (e.button !== 0 || !split.value) return;
  dragStart = {
    x: e.clientX,
    width: split.value.clientWidth,
    ratio: noteWidth.value,
  };
  (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
}
function moveResize(e: PointerEvent) {
  if (dragStart)
    noteWidth.value = Math.max(
      25,
      Math.min(
        50,
        dragStart.ratio - ((e.clientX - dragStart.x) / dragStart.width) * 100,
      ),
    );
}
function stopResize() {
  if (dragStart) {
    dragStart = undefined;
    void s.setting("reviewNoteWidth", String(Math.round(noteWidth.value)));
  }
}
function keyboardResize(e: KeyboardEvent) {
  if (e.key === "ArrowLeft" || e.key === "ArrowRight") {
    e.preventDefault();
    noteWidth.value = Math.max(
      25,
      Math.min(50, noteWidth.value + (e.key === "ArrowLeft" ? 2 : -2)),
    );
    void s.setting("reviewNoteWidth", String(noteWidth.value));
  }
}
</script>
<template>
  <section class="review-window">
    <aside class="review-nav" aria-label="复盘菜单">
      <button :class="{ active: page === 'home' }" @click="switchPage('home')">
        <House :size="17" />首页</button
      ><button
        :class="{ active: page === 'review' }"
        @click="switchPage('review')"
      >
        <NotebookPen :size="17" />复盘
      </button>
    </aside>
    <main
      class="review-main"
      :class="{ 'review-report-page': page === 'review' }"
    >
      <header v-if="page === 'home'" class="review-heading">
        <h1>任务统计</h1>
        <div class="review-filters">
          <label
            >开始日期<input
              v-model="from"
              type="date"
              aria-label="统计开始日期" /></label
          ><label
            >结束日期<input
              v-model="until"
              type="date"
              aria-label="统计结束日期"
              :min="from" /></label
          ><label class="review-status-field"
            >任务状态
            <CustomSelect v-model="status" aria-label="统计任务状态">
              <option value="all">全部状态</option>
              <option value="pending">未完成</option>
              <option value="completed">已完成</option>
            </CustomSelect> </label
          ><label class="check-label"
            ><input v-model="children" type="checkbox" />统计子任务</label
          >
        </div>
      </header>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <p v-if="from && until && from > until" class="error">
        开始日期不能晚于结束日期
      </p>
      <p v-if="loading" class="muted" role="status">读取任务报告…</p>
      <template v-if="page === 'home'"
        ><div class="review-overview-grid">
          <section class="review-distribution" aria-label="任务状态分布">
            <h2>任务分布</h2>
            <div class="review-pie-content">
              <svg
                class="review-pie"
                viewBox="0 0 120 120"
                role="img"
                :aria-label="`任务完成状态饼图：共 ${report.length} 项，已完成 ${done.length} 项，未完成 ${report.length - done.length} 项`"
              >
                <circle
                  v-if="!report.length"
                  cx="60"
                  cy="60"
                  r="56"
                  class="pie-empty"
                />
                <template v-for="slice in pieSlices" :key="slice.key">
                  <circle
                    v-if="slice.ratio === 1"
                    cx="60"
                    cy="60"
                    r="56"
                    :class="'pie-' + slice.key"
                  >
                    <title>{{ slice.label }} {{ slice.count }} 项</title>
                  </circle>
                  <path
                    v-else-if="slice.count"
                    :d="slice.path"
                    :class="'pie-' + slice.key"
                  >
                    <title>{{ slice.label }} {{ slice.count }} 项</title>
                  </path>
                </template>
              </svg>
              <div class="review-pie-legend">
                <p class="pie-total">
                  <strong>{{ report.length }}</strong
                  ><span>项任务</span>
                </p>
                <p v-for="slice in pieSlices" :key="slice.key">
                  <i :class="'pie-' + slice.key" /><span>{{ slice.label }}</span
                  ><strong>{{ slice.count }}</strong>
                </p>
                <p class="pie-rate">
                  <span>完成率</span><strong>{{ rate }}%</strong>
                </p>
              </div>
            </div>
          </section>
          <section class="review-heatmap">
            <h2>完成热力图</h2>
            <p class="muted small">
              按实际完成日期统计 · {{ from || "全部时间" }} —
              {{ until || "今天" }}
            </p>
            <div class="heatmap-scroll">
              <div class="heatmap-grid">
                <span
                  v-for="day in heatDays"
                  :key="day"
                  :data-level="heatLevel(day)"
                  :aria-label="day + ' 完成 ' + (counts.get(day) || 0) + ' 项'"
                  :title="day + ' · ' + (counts.get(day) || 0) + ' 项完成'"
                />
              </div>
            </div>
            <div class="heatmap-key">
              <small>少</small
              ><span
                v-for="level in 5"
                :key="level"
                :data-level="level - 1"
              /><small>多</small>
            </div>
          </section>
        </div>
        <div class="review-chart-grid">
          <section>
            <h2>最近两周</h2>
            <svg
              viewBox="0 0 450 135"
              role="img"
              aria-label="最近两周完成任务柱状图"
            >
              <g v-for="(d, i) in trend" :key="d.day">
                <rect
                  :x="i * 32 + 5"
                  :y="100 - (d.n / trendMax) * 84"
                  width="20"
                  :height="Math.max(1, (d.n / trendMax) * 84)"
                  fill="currentColor"
                >
                  <title>{{ d.day }}：{{ d.n }} 项</title>
                </rect>
                <text :x="i * 32 + 15" y="122" text-anchor="middle">
                  {{ d.day.slice(-2) }}
                </text>
              </g>
            </svg>
          </section>
          <section>
            <h2>标签分布</h2>
            <p v-if="!tagStats.some((t) => t.n)" class="muted small">
              给任务添加标签后，这里会显示分布。
            </p>
            <div
              v-for="tag in tagStats.filter((t) => t.n)"
              :key="tag.id"
              class="tag-stat"
            >
              <span>{{ tag.name }}</span
              ><meter
                :value="tag.n"
                :max="Math.max(1, report.length)"
              /><small>{{ tag.n }}</small>
            </div>
          </section>
        </div>
        <p v-if="!report.length" class="muted">
          当前范围暂无任务，可调整日期或勾选子任务。
        </p></template
      >
      <div
        v-else
        ref="split"
        class="review-split"
        :style="{
          gridTemplateColumns: `minmax(0,1fr) 8px minmax(260px,${noteWidth}%)`,
        }"
      >
        <section class="task-report">
          <header class="report-heading">
            <h1>任务总览</h1>
            <small class="muted">{{ report.length }} 项</small>
          </header>
          <div class="review-filters report-range-filters">
            <label
              >开始日期<input
                v-model="from"
                type="date"
                aria-label="统计开始日期"
            /></label>
            <label
              >结束日期<input
                v-model="until"
                type="date"
                aria-label="统计结束日期"
                :min="from"
            /></label>
            <label class="check-label"
              ><input v-model="children" type="checkbox" />统计子任务</label
            >
          </div>
          <div class="report-filters">
            <label class="search-control"
              ><Search :size="14" /><input
                v-model="search"
                placeholder="搜索任务内容"
                aria-label="搜索任务报告" /></label
            ><CustomSelect v-model="status" aria-label="报告任务状态"
              ><option value="all">全部状态</option>
              <option value="pending">未完成</option>
              <option value="completed">已完成</option></CustomSelect
            >
            <TagFilter v-model="tags" :options="s.tags" />
          </div>
          <p v-if="!report.length" class="muted">没有匹配的任务</p>
          <button
            v-for="t in report"
            :key="t.id"
            class="report-task"
            @click="editTask(t)"
          >
            <span :class="{ done: t.status === 'completed' }">{{
              t.title
            }}</span
            ><small
              >{{ t.status === "completed" ? "已完成" : "未完成" }} ·
              {{ t.date || dayString(new Date(t.startAt || t.createdAt))
              }}<template v-if="t.parentId"> · 子任务</template></small
            ><small v-if="t.tagIds?.length">{{
              s.tags
                .filter((x) => t.tagIds.includes(x.id))
                .map((x) => x.name)
                .join(" · ")
            }}</small>
          </button>
        </section>
        <div
          class="review-resizer"
          role="separator"
          tabindex="0"
          aria-label="调整笔记区域宽度"
          aria-orientation="vertical"
          @pointerdown="resize"
          @pointermove="moveResize"
          @pointerup="stopResize"
          @pointercancel="stopResize"
          @keydown="keyboardResize"
        />
        <section class="notes-pane" aria-label="复盘笔记">
          <template v-if="!draft"
            ><header>
              <h2>NOTE</h2>
              <button
                class="icon-button"
                aria-label="新建笔记"
                @click="openNote()"
              >
                <Plus :size="17" />
              </button>
            </header>
            <label class="search-control"
              ><Search :size="13" /><input
                v-model="noteSearch"
                aria-label="搜索笔记"
                placeholder="搜索笔记"
                @input="loadNotes"
            /></label>
            <p v-if="!notes.length" class="muted small">
              还没有笔记，记录一次思考。
            </p>
            <button
              v-for="note in notes"
              :key="note.id"
              class="note-item"
              @click="openNote(note)"
            >
              <strong>{{ note.title }}</strong
              ><small
                >创建
                {{ new Date(note.createdAt).toLocaleString("zh-CN") }}</small
              ><small
                >更新
                {{ new Date(note.updatedAt).toLocaleString("zh-CN") }}</small
              >
            </button></template
          ><template v-else
            ><header>
              <button
                class="icon-button"
                aria-label="返回笔记列表"
                @click="backNotes"
              >
                <ArrowLeft :size="16" /></button
              ><span class="grow">{{
                draft.id ? "编辑笔记" : "新建 NOTE"
              }}</span
              ><button
                v-if="draft.id"
                class="icon-button action-delete"
                aria-label="删除笔记"
                title="删除笔记"
                :disabled="saving"
                @click="deleteNote"
              >
                <Trash2 :size="15" /></button
              ><button
                class="icon-button"
                :aria-label="preview ? '切换笔记编辑' : '切换笔记预览'"
                @click="preview = !preview"
              >
                <Pencil v-if="preview" :size="15" /><Eye
                  v-else
                  :size="15"
                /></button
              ><button
                class="icon-button"
                aria-label="保存笔记"
                :disabled="saving"
                @click="saveNote"
              >
                <Save :size="15" />
              </button>
            </header>
            <input
              v-model="draft.title"
              class="note-title"
              aria-label="笔记标题"
              placeholder="标题（必填）"
              maxlength="200"
            />
            <article
              v-if="preview"
              class="markdown-body note-preview"
              @click="noteLink"
              v-html="rendered"
            />
            <textarea
              v-else
              v-model="draft.content"
              class="note-content"
              aria-label="Markdown 笔记内容"
              placeholder="写下你的复盘，支持 Markdown…"
              @keydown.meta.s.prevent="saveNote"
              @keydown.ctrl.s.prevent="saveNote"
            /><small class="muted"
              >{{ dirty() ? "未保存" : "已保存"
              }}<template v-if="draft.updatedAt">
                ·
                {{
                  new Date(draft.updatedAt).toLocaleString("zh-CN")
                }}</template
              ></small
            ></template
          >
        </section>
      </div>
    </main>
    <TaskEditor
      v-if="editor"
      :task="editor"
      @close="
        editor = null;
        load();
      "
    />
  </section>
</template>
