<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from "vue";
import {
  CalendarDays,
  ListTodo,
  Timer,
  Settings,
  Upload,
  Download,
  Search,
  CheckCheck,
  Trash2,
  Copy,
  ClipboardPaste,
  Pin,
  Minus,
  Maximize2,
  X,
  ChevronLeft,
  ChevronRight,
  ChevronsUpDown,
  MoreHorizontal,
  Bell,
  AlertCircle,
} from "lucide-vue-next";
import { useTodo } from "./stores/todo";
import { api, listen, confirmAction } from "./api";
import {
  type Task,
  type Holiday,
  type Preview,
  dayString,
  parseDay,
  shiftDay,
} from "./api/types";
import { canLeaveEditor, hasDirtyEditor } from "./composables/editorGuard";
import MotionPresence from "./components/MotionPresence.vue";
import { setMotionEnabled, installControlMotion } from "./composables/motion";
import DayTimeline from "./components/DayTimeline.vue";
import TaskList from "./components/TaskList.vue";
import TaskEditor from "./components/TaskEditor.vue";
import CalendarTaskScroll from "./components/CalendarTaskScroll.vue";
import { startTaskDrag, cancelTaskDrag } from "./composables/taskDrag";
import CustomSelect from "./components/CustomSelect.vue";
import FocusPanel from "./components/FocusPanel.vue";
import SettingsDialog from "./components/SettingsDialog.vue";
import TransferDialog from "./components/TransferDialog.vue";
const s = useTodo(),
  panel = ref("tasks"),
  listView = ref("day"),
  editor = ref<Task | null>(null),
  settingsOpen = ref(false),
  settingsTab = ref("general"),
  transfer = ref<"import" | "export" | "batch" | null>(null),
  batchAction = ref("complete"),
  batch = ref(false),
  search = ref(""),
  menu = ref(false),
  notices = ref(false),
  dragging = ref(false),
  holidays = ref<Holiday[]>([]),
  activeDay = ref(s.date),
  drawer = ref(false),
  month = ref(s.date.slice(0, 7)),
  calendarMode = ref("month"),
  rangeAnchor = ref(""),
  rangeEnd = ref(""),
  rangeDragging = ref(false),
  rangeMoved = ref(false),
  pastedText = ref(""),
  pastedPreview = ref<Preview | null>(null);
const stopListeners: (() => void)[] = [];
let poll: ReturnType<typeof setInterval>,
  previousToday = s.date;
const dateLabel = computed(() =>
  parseDay(s.date).toLocaleDateString("zh-CN", {
    month: "long",
    day: "numeric",
    weekday: "long",
  }),
);
const monthDays = computed(() => {
  const first = month.value + "-01",
    n = (parseDay(first).getDay() + 6) % 7;
  return Array.from({ length: 42 }, (_, i) => shiftDay(first, i - n));
});
const calendarDays = computed(() =>
  calendarMode.value === "month"
    ? monthDays.value
    : Array.from({ length: 7 }, (_, i) =>
        shiftDay(
          activeDay.value,
          i - ((parseDay(activeDay.value).getDay() + 6) % 7),
        ),
      ),
);
const year = computed(() => month.value.slice(0, 4)),
  monthNumber = computed(() => month.value.slice(5));
const windowState = computed(() =>
  s.boot?.windows.find((w) => w.key === "main"),
);
const roots = computed(() =>
  s.roots.filter((t) => {
    const dateMatches =
      panel.value !== "calendar" || t.date === activeDay.value;
    const textMatches =
      !search.value.trim() ||
      s.snapshot.tasks.some(
        (n) =>
          n.rootId === t.id &&
          n.title.toLowerCase().includes(search.value.toLowerCase()),
      );
    return dateMatches && textMatches;
  }),
);
const completed = computed(
  () => roots.value.filter((t) => t.status === "completed").length,
);
const calendarTasks = (day: string) =>
  s.snapshot.tasks.filter(
    (t) => !t.parentId && t.date === day && t.status !== "skipped",
  );
const isRange = (day: string) =>
  !!(
    rangeAnchor.value &&
    rangeEnd.value &&
    day >= [rangeAnchor.value, rangeEnd.value].sort()[0] &&
    day <= [rangeAnchor.value, rangeEnd.value].sort()[1]
  );
function applyTheme() {
  const theme = s.boot?.settings.theme;
  document.documentElement.dataset.theme =
    theme === "dark" ||
    (theme === "system" && matchMedia("(prefers-color-scheme: dark)").matches)
      ? "dark"
      : "light";
}
watch(() => s.boot?.settings.theme, applyTheme, { immediate: true });
watch(
  () => s.boot?.settings.animation,
  (v) => {
    document.documentElement.classList.toggle("no-motion", v === "false");
    setMotionEnabled(v !== "false");
  },
  { immediate: true },
);
watch(
  () => s.boot?.today,
  (v) => {
    if (v && v !== previousToday) {
      if (
        s.date === previousToday &&
        !editor.value &&
        !dragging.value &&
        !document.querySelector<HTMLInputElement>("[data-new-child]")?.value &&
        !document.querySelector<HTMLInputElement>(
          '[aria-label="新待办标题"]',
        ) &&
        !hasDirtyEditor() &&
        panel.value === "tasks"
      ) {
        s.date = v;
        void s.load();
      }
      previousToday = v;
    }
  },
);
watch(
  () => s.date,
  () => {
    if (panel.value === "tasks") {
      s.selected = [];
      void s.load();
    }
  },
);

function clamp(day: string) {
  return day < "2000-01-01"
    ? "2000-01-01"
    : day > "2100-12-31"
      ? "2100-12-31"
      : day;
}
async function loadCalendar() {
  if (panel.value !== "calendar") return;
  s.view = "month";
  s.date = clamp(calendarDays.value[0]);
  s.end = clamp(calendarDays.value.at(-1)!);
  await s.load();
}
async function chooseEditor(task: Task | null) {
  if (task?.id === editor.value?.id) return;
  if (task && !(await canLeaveEditor())) return;
  editor.value = task;
}
async function closeDrawer() {
  if (!(await canLeaveEditor())) return;
  drawer.value = false;
  editor.value = null;
}
async function openSettings(tab: string) {
  if (!(await canLeaveEditor())) return;
  editor.value = null;
  drawer.value = false;
  settingsTab.value = tab;
  settingsOpen.value = true;
  menu.value = false;
}
async function openTransfer(mode: "import" | "export") {
  if (!(await canLeaveEditor())) return;
  editor.value = null;
  drawer.value = false;
  transfer.value = mode;
  menu.value = false;
}
async function changeCalendar(
  kind: "year" | "month" | "mode",
  value: string | number,
) {
  if (!(await canLeaveEditor())) return;
  editor.value = null;
  drawer.value = false;
  if (kind === "year") month.value = String(value) + month.value.slice(4);
  else if (kind === "month")
    month.value = month.value.slice(0, 5) + String(value);
  else calendarMode.value = String(value);
  await loadCalendar();
}
async function selectListDay(day: string) {
  if (!(await canLeaveEditor())) return;
  editor.value = null;
  s.date = day;
  activeDay.value = day;
}
async function openBatch(action: string) {
  if (!(await canLeaveEditor())) return;
  editor.value = null;
  batchAction.value = action;
  transfer.value = "batch";
}
async function navigate(next: string) {
  if (!(await canLeaveEditor())) return false;
  settingsOpen.value = false;
  transfer.value = null;
  menu.value = false;
  editor.value = null;
  drawer.value = false;
  s.selected = [];
  panel.value = next;
  if (next === "calendar") {
    month.value = activeDay.value.slice(0, 7);
    await loadCalendar();
  } else {
    s.view = listView.value;
    s.date = activeDay.value;
    await s.load();
  }
  await s.run(() => api("SetView", next));
  return true;
}
async function toggleList() {
  if (!(await canLeaveEditor())) return;
  listView.value = listView.value === "day" ? "inbox" : "day";
  s.view = listView.value;
  search.value = "";
  s.selected = [];
  editor.value = null;
  await s.load();
}
async function step(n: number) {
  if (!(await canLeaveEditor())) return;
  editor.value = null;
  if (panel.value !== "calendar") {
    s.date = clamp(shiftDay(s.date, n));
    activeDay.value = s.date;
    return;
  }
  if (calendarMode.value === "week") {
    activeDay.value = clamp(shiftDay(activeDay.value, n * 7));
    month.value = activeDay.value.slice(0, 7);
  } else {
    const d = parseDay(month.value + "-01");
    d.setMonth(d.getMonth() + n);
    if (d.getFullYear() < 2000 || d.getFullYear() > 2100) return;
    month.value = dayString(d).slice(0, 7);
    activeDay.value = month.value + "-01";
  }
  drawer.value = false;
  rangeAnchor.value = "";
  rangeEnd.value = "";
  await loadCalendar();
}
async function today() {
  if (!(await canLeaveEditor())) return;
  editor.value = null;
  activeDay.value = s.boot?.today || dayString(new Date());
  month.value = activeDay.value.slice(0, 7);
  if (panel.value === "calendar") await loadCalendar();
  else {
    s.date = activeDay.value;
    listView.value = "day";
    s.view = "day";
    await s.load();
  }
}
async function pickDay(day: string, e?: MouseEvent) {
  if (day < "2000-01-01" || day > "2100-12-31") return;
  if (!(await canLeaveEditor())) return;
  if (rangeMoved.value) {
    rangeMoved.value = false;
    drawer.value = true;
    return;
  }
  activeDay.value = day;
  drawer.value = true;
  editor.value = null;
  if (e?.shiftKey && rangeAnchor.value) rangeEnd.value = day;
  else {
    rangeAnchor.value = day;
    rangeEnd.value = day;
  }
  selectRange();
}
function calendarNativeDrag(event: DragEvent, task: Task) {
  if (task.seriesId || s.saving) {
    event.preventDefault();
    return;
  }
  event.dataTransfer?.setData("text/plain", task.id);
  if (event.dataTransfer) event.dataTransfer.effectAllowed = "move";
}
async function openCalendarTask(task: Task) {
  if (!(await canLeaveEditor())) return;
  activeDay.value = task.date || activeDay.value;
  drawer.value = true;
  rangeAnchor.value = activeDay.value;
  rangeEnd.value = activeDay.value;
  selectRange();
  s.activeTask = task.id;
  editor.value = task;
}
function selectRange() {
  s.selected = s.snapshot.tasks
    .filter((t) => !t.parentId && isRange(t.date || ""))
    .map((t) => t.id);
}
function rangeStart(day: string, e: PointerEvent) {
  if (day < "2000-01-01" || day > "2100-12-31") return;
  if (
    hasDirtyEditor() ||
    (e.target as HTMLElement).closest(".calendar-task-list")
  )
    return;
  if (e.button !== 0 || e.shiftKey) return;
  rangeDragging.value = true;
  rangeMoved.value = false;
  rangeAnchor.value = day;
  rangeEnd.value = day;
}
function rangeOver(day: string, e: PointerEvent) {
  if (!rangeDragging.value || e.buttons !== 1) return;
  rangeMoved.value = rangeEnd.value !== day;
  rangeEnd.value = day;
  selectRange();
}
function rangeStop() {
  rangeDragging.value = false;
}
async function outside(e: PointerEvent) {
  const el = e.target as HTMLElement;
  if (!el.closest(".tools")) menu.value = false;
  if (
    drawer.value &&
    !el.closest(
      ".day-drawer,.calendar-day,.select-options,.reminder-popover,.node-menu,.app-header",
    )
  ) {
    if (!(await canLeaveEditor())) return;
    drawer.value = false;
    editor.value = null;
  }
}
async function dropDate(e: DragEvent, date: string | null) {
  const t = s.taskMap.get(e.dataTransfer?.getData("text/plain") || "");
  if (!t) return;
  if (t.parentId) {
    s.inform("子任务随根任务日期移动");
    return;
  }
  if (t.date === date || !(await canLeaveEditor())) return;
  const moved = await s.move(t, date);
  if (moved) editor.value = null;
  dragging.value = false;
}
async function rootDrop(target: Task, e: DragEvent) {
  const source = s.taskMap.get(e.dataTransfer?.getData("text/plain") || "");
  if (!source || source.id === target.id) return;
  if (source.date !== target.date) {
    await s.move(source, target.date);
    return;
  }
  const peers = s.snapshot.tasks
      .filter(
        (t) => !t.parentId && t.date === source.date && t.status !== "skipped",
      )
      .sort((a, b) => a.order - b.order),
    ids = peers.filter((t) => t.id !== source.id).map((t) => t.id);
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
  ids.splice(
    ids.indexOf(target.id) +
      (!e.altKey && e.clientY > rect.top + rect.height * 0.75 ? 1 : 0),
    0,
    source.id,
  );
  await s.run(() => api("Reorder", ids, s.snapshot.revision));
  dragging.value = false;
}
async function importLines(text: string) {
  if (!(await canLeaveEditor())) return;
  editor.value = null;
  pastedText.value = text;
  transfer.value = "import";
}
async function copySelection() {
  const sourceQuery =
    panel.value === "calendar" && rangeAnchor.value && rangeEnd.value
      ? {
          ...s.query,
          date: [rangeAnchor.value, rangeEnd.value].sort()[0],
          end: [rangeAnchor.value, rangeEnd.value].sort()[1],
          limit: 10000,
        }
      : s.query;
  const text = await s.run(() =>
    api<string>("CopyTasks", sourceQuery, s.selected),
  );
  if (text !== undefined) {
    await s.run(() => api("ClipboardWrite", text));
    s.inform("已复制任务");
  }
}
async function pasteSelection() {
  if (!(await canLeaveEditor())) return;
  editor.value = null;
  await s.run(async () => {
    const text = await api<string>("ClipboardRead");
    pastedPreview.value = await api<Preview>(
      "PreviewPaste",
      text,
      rangeAnchor.value && rangeEnd.value
        ? [rangeAnchor.value, rangeEnd.value].sort()[0]
        : activeDay.value,
    );
    transfer.value = "import";
  });
}
async function pin() {
  await s.run(() =>
    api(
      "WindowConfig",
      "main",
      windowState.value?.mode === "top" ? "normal" : "top",
      1,
      false,
    ),
  );
  await s.metadata();
}
async function quit() {
  if (
    document.querySelector<HTMLInputElement>(".inline-add input")?.value ||
    editor.value
  ) {
    if (!(await confirmAction("有正在编辑的内容，确定退出？"))) return;
  }
  if (
    (await api<boolean>("QuitStatus")) &&
    !(await confirmAction("有专注或近期提醒，退出后将停止后台提醒。仍然退出？"))
  )
    return;
  await api("Quit");
}
async function recover() {
  const info = await s.run(() =>
    api<{ path: string; name: string; roots: number; nodes: number } | null>(
      "PrepareRestore",
      "",
    ),
  );
  if (
    info &&
    (await confirmAction(
      `恢复「${info.name}」？${info.roots} 个根任务，${info.nodes} 个节点。原数据库文件将被保留。`,
    ))
  ) {
    await s.run(() => api("RestorePrepared", info.path));
    await s.metadata();
    await s.load();
  }
}
function confirmationFailed(e: Event) {
  s.error = String((e as CustomEvent).detail);
}
async function keyboard(e: KeyboardEvent) {
  if (e.defaultPrevented || e.isComposing) return;
  const editing = (e.target as HTMLElement).matches(
      'input,textarea,[contenteditable="true"],[role="combobox"]',
    ),
    mod = e.metaKey || e.ctrlKey;
  if (mod && e.key.toLowerCase() === "n") {
    e.preventDefault();
    if (!(await canLeaveEditor())) return;
    editor.value = null;
    if (settingsOpen.value || transfer.value) {
      settingsOpen.value = false;
      transfer.value = null;
    }
    if (panel.value === "focus") await navigate("tasks");
    if (panel.value === "calendar") drawer.value = true;
    await nextTick();
    document.dispatchEvent(new Event("todo:new-task"));
  } else if (mod && e.key.toLowerCase() === "t" && !editing) {
    e.preventDefault();
    await today();
  } else if (
    mod &&
    e.key.toLowerCase() === "c" &&
    !editing &&
    s.selected.length
  ) {
    e.preventDefault();
    await copySelection();
  } else if (
    mod &&
    e.key.toLowerCase() === "v" &&
    !editing &&
    panel.value === "calendar"
  ) {
    e.preventDefault();
    await pasteSelection();
  } else if (e.key === "Escape" && !editing) {
    if (transfer.value) {
      transfer.value = null;
    } else if (settingsOpen.value) {
      settingsOpen.value = false;
    } else if (menu.value) {
      menu.value = false;
    } else if (drawer.value) {
      await closeDrawer();
    } else {
      batch.value = false;
      s.selected = [];
      notices.value = false;
    }
  }
}
onMounted(async () => {
  stopListeners.push(installControlMotion());
  await s.init();
  previousToday = s.boot?.today || s.date;
  try {
    holidays.value = await api<Holiday[]>("Holidays");
  } catch (e) {
    s.error = String(e);
  }
  stopListeners.push(
    await listen("todo:navigate", async (value) => {
      if (value === "calendar" || value === "overview")
        await navigate("calendar");
      else if (value === "focus") await navigate("focus");
      else {
        await navigate("tasks");
        if (value === "quick") {
          await nextTick();
          document.dispatchEvent(new Event("todo:new-task"));
        }
      }
    }),
  );
  stopListeners.push(await listen("todo:quit-request", () => void quit()));
  stopListeners.push(
    await listen("todo:notification", async (data) => {
      if (data?.sourceType !== "task" || !data.sourceId) {
        await navigate("focus");
        return;
      }
      try {
        const root = await api<Task>("FindTask", data.sourceId);
        if (!(await navigate("tasks"))) return;
        s.hideCompleted = false;
        listView.value = root.date ? "day" : "inbox";
        s.view = listView.value;
        activeDay.value = root.date || s.boot?.today || s.date;
        s.date = activeDay.value;
        await s.load();
        await nextTick();
        document
          .querySelector<HTMLElement>(`[data-node="${root.id}"]`)
          ?.focus();
      } catch (e) {
        s.inform("提醒对应的任务可能已被删除");
      }
    }),
  );
  poll = setInterval(() => {
    void s.metadata().catch(() => {});
    void s.load();
  }, 5000);
  document.addEventListener("keydown", keyboard);
  document.addEventListener("pointerdown", outside);
  document.addEventListener("pointerup", rangeStop);
  window.addEventListener("todo:confirm-error", confirmationFailed);
  matchMedia("(prefers-color-scheme: dark)").addEventListener(
    "change",
    applyTheme,
  );
});
onUnmounted(() => {
  cancelTaskDrag();
  clearInterval(poll);
  stopListeners.forEach((fn) => fn());
  document.removeEventListener("keydown", keyboard);
  document.removeEventListener("pointerdown", outside);
  document.removeEventListener("pointerup", rangeStop);
  window.removeEventListener("todo:confirm-error", confirmationFailed);
  matchMedia("(prefers-color-scheme: dark)").removeEventListener(
    "change",
    applyTheme,
  );
});
</script>
<template>
  <main
    class="app-shell compact-app"
    :class="{
      'calendar-view': panel === 'calendar',
      'on-mac': s.boot?.capabilities.platform === 'darwin',
    }"
    :inert="s.maintenance || undefined"
    @dragend="dragging = false"
  >
    <header
      class="app-header"
      @dblclick.self="s.run(() => api('WindowAction', 'maximize'))"
    >
      <strong class="brand">SimpleTodo</strong>
      <nav class="header-actions" aria-label="窗口与视图">
        <button
          class="icon-button"
          :class="{ active: panel === 'tasks' && !settingsOpen }"
          aria-label="任务列表"
          title="任务列表"
          @click="navigate('tasks')"
        >
          <ListTodo :size="16" />
        </button>
        <button
          class="icon-button"
          :class="{ active: panel === 'calendar' && !settingsOpen }"
          aria-label="日历视图"
          title="日历视图"
          @click="navigate('calendar')"
        >
          <CalendarDays :size="16" />
        </button>
        <button
          class="icon-button"
          :class="{ active: panel === 'focus' && !settingsOpen }"
          aria-label="专注与倒计时"
          title="专注与倒计时"
          @click="navigate('focus')"
        >
          <Timer :size="16" />
        </button>
        <span class="toolbar-divider" />
        <button
          class="icon-button"
          :class="{ active: windowState?.mode === 'top' }"
          aria-label="固定窗口"
          title="固定窗口"
          @click="pin"
        >
          <Pin :size="14" />
        </button>
        <div class="tools">
          <button
            :aria-expanded="menu"
            class="icon-button"
            aria-label="工具菜单"
            @click.prevent="menu = !menu"
          >
            <MoreHorizontal :size="17" />
          </button>
          <MotionPresence
            ><div v-if="menu" class="menu">
              <button @click="openSettings('general')">
                <Settings :size="14" />设置</button
              ><button @click="openTransfer('import')">
                <Upload :size="14" />导入待办</button
              ><button @click="openTransfer('export')">
                <Download :size="14" />导出待办</button
              ><button
                @click="
                  notices = !notices;
                  menu = false;
                "
              >
                <Bell :size="14" />错过的提醒 {{ s.boot?.notices.length || 0 }}
              </button>
              <hr />
              <button @click="openSettings('about')">帮助与关于</button
              ><button
                @click="
                  quit();
                  menu = false;
                "
              >
                退出 SimpleTodo
              </button>
            </div></MotionPresence
          >
        </div>
        <button
          class="icon-button"
          aria-label="最小化窗口"
          title="最小化"
          @click="s.run(() => api('WindowAction', 'minimize'))"
        >
          <Minus :size="15" />
        </button>
        <button
          class="icon-button"
          aria-label="放大或还原窗口"
          title="放大 / 还原"
          @click="s.run(() => api('WindowAction', 'maximize'))"
        >
          <Maximize2 :size="14" />
        </button>
        <button
          class="icon-button window-close"
          aria-label="关闭窗口"
          title="关闭"
          @click="s.run(() => api('WindowAction', 'close'))"
        >
          <X :size="15" />
        </button>
      </nav>
    </header>
    <section v-if="s.boot?.startupError" class="recovery-panel">
      <h1>恢复任务数据</h1>
      <p role="alert">{{ s.boot.startupError }}</p>
      <button class="button primary" @click="recover">选择完整备份恢复</button>
    </section>
    <SettingsDialog
      v-else-if="settingsOpen"
      :initial-tab="settingsTab"
      @close="settingsOpen = false"
    />
    <TransferDialog
      v-else-if="transfer"
      :mode="transfer"
      :action="batchAction"
      :initial="pastedPreview"
      :initial-text="pastedText"
      @close="
        transfer = null;
        pastedPreview = null;
        pastedText = '';
      "
    />
    <FocusPanel v-else-if="panel === 'focus'" />
    <template v-else>
      <section v-if="panel === 'tasks'" class="planning-bar">
        <div class="list-heading">
          <button
            class="view-switch"
            @click="toggleList"
            @dragover.prevent
            @drop="dropDate($event, null)"
          >
            {{ listView === "day" ? "日计划" : "收集箱"
            }}<ChevronsUpDown :size="13" /></button
          ><span class="muted small">{{
            listView === "day" ? dateLabel : "全部任务"
          }}</span
          ><span class="grow" /><button
            class="icon-button"
            aria-label="打开设置"
            title="设置"
            @click="openSettings('general')"
          >
            <Settings :size="14" />
          </button>
        </div>
        <DayTimeline
          v-if="listView === 'day'"
          :date="s.date"
          :today="s.boot?.today"
          @select="selectListDay"
          @step="step"
          @today="today"
          @drop="(event, day) => dropDate(event, day)"
        />
      </section>
      <section v-else class="calendar-toolbar">
        <div class="calendar-month-picker">
          <CustomSelect
            :model-value="year"
            @update:model-value="changeCalendar('year', $event)"
            aria-label="年份"
            ><option v-for="y in 101" :key="y" :value="String(1999 + y)">
              {{ 1999 + y }}年
            </option></CustomSelect
          ><CustomSelect
            :model-value="monthNumber"
            @update:model-value="changeCalendar('month', $event)"
            aria-label="月份"
            ><option
              v-for="m in 12"
              :key="m"
              :value="String(m).padStart(2, '0')"
            >
              {{ m }}月
            </option></CustomSelect
          >
        </div>
        <span class="grow" /><CustomSelect
          :model-value="calendarMode"
          @update:model-value="changeCalendar('mode', $event)"
          aria-label="日历范围"
          ><option value="month">月视图</option>
          <option value="week">周视图</option></CustomSelect
        ><button class="icon-button" aria-label="上一个周期" @click="step(-1)">
          <ChevronLeft :size="15" /></button
        ><button class="text-button" @click="today">今天</button
        ><button class="icon-button" aria-label="下一个周期" @click="step(1)">
          <ChevronRight :size="15" />
        </button>
      </section>
      <div v-if="panel === 'tasks'" class="list-toolbar">
        <label class="search-control"
          ><Search :size="13" /><input
            v-model="search"
            placeholder="搜索任务"
            aria-label="搜索当前视图"
        /></label>
        <button
          class="text-button"
          @click="
            batch = !batch;
            s.selected = [];
          "
        >
          {{ batch ? "完成选择" : "选择" }}
        </button>
        <template v-if="batch"
          ><button
            class="icon-button"
            aria-label="批量完成"
            :disabled="!s.selected.length"
            @click="openBatch('complete')"
          >
            <CheckCheck :size="15" /></button
          ><button
            class="icon-button"
            aria-label="批量取消完成"
            :disabled="!s.selected.length"
            @click="openBatch('pending')"
          >
            <ListTodo :size="15" /></button
          ><button
            class="icon-button"
            aria-label="批量删除"
            :disabled="!s.selected.length"
            @click="openBatch('delete')"
          >
            <Trash2 :size="14" /></button
        ></template>
      </div>
      <div
        v-if="
          s.selected.length && panel === 'calendar' && rangeAnchor !== rangeEnd
        "
        class="batch-bar"
      >
        <span>已选 {{ s.selected.length }} 项</span><span class="grow" /><button
          class="icon-button"
          aria-label="批量完成"
          @click="openBatch('complete')"
        >
          <CheckCheck :size="15" /></button
        ><button
          class="icon-button"
          aria-label="批量取消完成"
          @click="openBatch('pending')"
        >
          <ListTodo :size="15" /></button
        ><button
          class="icon-button"
          aria-label="复制任务"
          @click="copySelection"
        >
          <Copy :size="14" /></button
        ><button
          class="icon-button"
          aria-label="粘贴到选中日期"
          @click="pasteSelection"
        >
          <ClipboardPaste :size="14" /></button
        ><button
          class="icon-button"
          aria-label="批量删除"
          @click="openBatch('delete')"
        >
          <Trash2 :size="14" />
        </button>
      </div>
      <TaskList
        v-if="panel === 'tasks'"
        :roots="roots"
        :date="listView === 'inbox' ? null : s.date"
        :editing="editor?.id"
        :batch="batch"
        @edit="chooseEditor"
        @drag="dragging = true"
        @drop="rootDrop"
        @paste="importLines"
      />
      <footer
        v-if="panel === 'tasks'"
        v-motion="completed"
        class="list-footer"
        aria-label="任务完成统计"
      >
        <span
          v-if="roots.length && completed === roots.length"
          v-motion
          class="all-done"
          ><CheckCheck :size="13" /> 都完成了</span
        >
        {{ completed }} / {{ roots.length }} 已完成
      </footer>
      <section
        v-else
        class="calendar-workspace"
        :class="{ 'week-mode': calendarMode === 'week' }"
      >
        <div class="month-weekdays">
          <span v-for="d in ['一', '二', '三', '四', '五', '六', '日']" :key="d"
            >周{{ d }}</span
          >
        </div>
        <div v-motion="calendarMode + calendarDays[0]" class="calendar-grid">
          <div
            v-for="day in calendarDays"
            :key="day"
            class="calendar-day"
            :class="{
              outside: calendarMode === 'month' && day.slice(0, 7) !== month,
              today: day === s.boot?.today,
              selected: drawer && day === activeDay,
              inrange: isRange(day) && rangeAnchor !== rangeEnd,
              weekend: [0, 6].includes(parseDay(day).getDay()),
            }"
            :data-day="day"
            role="group"
            :aria-label="'日历日期 ' + day"
            @pointerdown="rangeStart(day, $event)"
            @pointerenter="rangeOver(day, $event)"
            @click="pickDay(day, $event)"
            @dragover.prevent
            @drop="dropDate($event, day)"
          >
            <button
              class="calendar-number"
              :aria-label="day"
              :disabled="day < '2000-01-01' || day > '2100-12-31'"
            >
              {{ Number(day.slice(-2))
              }}<small v-if="holidays.find((h) => h.date === day)">{{
                holidays.find((h) => h.date === day)?.name
              }}</small>
            </button>
            <CalendarTaskScroll :label="day + ' 任务'">
              <button
                v-for="t in calendarTasks(day)"
                :key="t.id"
                class="calendar-task"
                :class="{ done: t.status === 'completed' }"
                :title="
                  t.seriesId
                    ? t.title + '（重复任务请在详情中调整规则）'
                    : t.title
                "
                :data-calendar-task="t.id"
                :aria-label="'编辑任务：' + t.title"
                :draggable="!t.seriesId && !s.saving"
                @pointerdown.stop="
                  !t.seriesId && !s.saving && startTaskDrag($event, t.id)
                "
                @click.stop="openCalendarTask(t)"
                @dragstart.stop="calendarNativeDrag($event, t)"
              >
                {{ t.title }}
              </button>
            </CalendarTaskScroll>
          </div>
        </div>
        <MotionPresence kind="drawer"
          ><aside v-if="drawer" class="day-drawer" aria-label="日期任务抽屉">
            <header class="drawer-header">
              <div>
                <h2>
                  {{
                    parseDay(activeDay).toLocaleDateString("zh-CN", {
                      month: "long",
                      day: "numeric",
                      weekday: "long",
                    })
                  }}
                </h2>
                <small>{{ roots.length }} 项任务</small>
              </div>
              <button
                class="icon-button"
                aria-label="关闭日期任务"
                @click="closeDrawer"
              >
                <X :size="16" />
              </button>
            </header>
            <TaskEditor
              v-if="editor"
              :key="editor.id"
              :task="editor"
              embedded
              @close="editor = null"
            />
            <TaskList
              v-else
              details-on-click
              :roots="roots"
              :date="activeDay"
              @edit="chooseEditor"
              @drag="dragging = true"
              @drop="rootDrop"
              @paste="importLines"
            /></aside
        ></MotionPresence>
      </section>
      <section v-if="notices" v-motion:popover class="notice-panel">
        <header>
          <strong>错过的提醒</strong
          ><button
            class="icon-button"
            aria-label="关闭提醒记录"
            @click="notices = false"
          >
            <X :size="14" />
          </button>
        </header>
        <p v-if="!s.boot?.notices.length" class="muted">没有错过的提醒。</p>
        <p v-for="n in s.boot?.notices" :key="n.id">
          {{ n.title }} <small>{{ new Date(n.fireAt).toLocaleString() }}</small>
        </p>
        <button
          v-if="s.boot?.notices.length"
          class="text-button"
          @click="s.run(() => api('DismissNotices')).then(() => s.metadata())"
        >
          清空记录
        </button>
      </section>
    </template>
    <div
      v-if="s.error && !settingsOpen && !transfer"
      class="error-banner"
      role="alert"
    >
      <AlertCircle :size="14" /><span>{{ s.error }}</span
      ><button
        class="text-button"
        @click="
          s.error = '';
          s.load();
        "
      >
        重新加载任务</button
      ><button
        class="icon-button"
        aria-label="关闭错误提示"
        @click="s.error = ''"
      >
        <X :size="13" />
      </button>
    </div>
    <MotionPresence kind="toast"
      ><div v-if="s.toast" class="toast" role="status">
        {{ s.toast }}
      </div></MotionPresence
    >
  </main>
  <div v-if="s.maintenance" class="maintenance-mask" role="status">
    正在处理数据，请稍候…
  </div>
</template>
