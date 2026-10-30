<script setup lang="ts">
import CustomSelect from "./CustomSelect.vue";
import { restoreDialogFocus } from "../composables/dialogFocus";
import { computed, ref, onMounted } from "vue";
import {
  X,
  Clipboard,
  FolderOpen,
  Download,
  ArrowRight,
  Check,
  AlertCircle,
} from "lucide-vue-next";
import { api } from "../api";
import { type Preview, type Task } from "../api/types";
import { useTodo } from "../stores/todo";
const props = defineProps<{
    mode: "import" | "export" | "batch";
    action?: string;
    initial?: Preview | null;
    initialText?: string;
  }>(),
  emit = defineEmits<{ close: [] }>(),
  s = useTodo(),
  dialog = ref<HTMLElement>(),
  content = ref(props.initialText || ""),
  format = ref(props.mode === "export" ? "markdown" : "text"),
  replace = ref(false),
  keepCompleted = ref(true),
  preview = ref<Preview | null>(props.initial || null),
  includeInbox = ref(false),
  scope = ref("current"),
  from = ref(s.date),
  until = ref(s.date),
  copyStatus = ref("");
const previewNodes = computed(() => {
  const tasks = preview.value?.tasks || [];
  const children = new Map<string, Task[]>();
  const ids = new Set(tasks.map((t) => t.id));
  for (const task of tasks) {
    const key = task.parentId && ids.has(task.parentId) ? task.parentId : "";
    children.set(key, [...(children.get(key) || []), task]);
  }
  const ordered: Task[] = [];
  const visit = (parent: string) => {
    for (const task of (children.get(parent) || []).sort(
      (a, b) => a.order - b.order || a.id.localeCompare(b.id),
    )) {
      ordered.push(task);
      visit(task.id);
    }
  };
  visit("");
  return ordered;
});
const batchVerb = computed(() =>
  props.action === "delete"
    ? "删除"
    : props.action === "pending"
      ? "恢复未完成"
      : "完成",
);
const confirmLabel = computed(() =>
  props.mode === "batch" && preview.value
    ? `${batchVerb.value} ${preview.value.roots} 项任务及其后代`
    : "确认执行",
);
restoreDialogFocus();
onMounted(async () => {
  if (props.mode === "batch") await getBatch();
});
function query() {
  return scope.value === "all"
    ? { ...s.query, view: "all" }
    : scope.value === "range"
      ? { ...s.query, view: "month", date: from.value, end: until.value }
      : s.query;
}
function depth(t: Task) {
  let n = 0,
    parent = t.parentId;
  const map = new Map(preview.value?.tasks.map((t) => [t.id, t]) || []);
  while (parent && n < 201) {
    n++;
    parent = map.get(parent)?.parentId || null;
  }
  return n;
}
async function getPreview() {
  const p = await s.run(() =>
    api<Preview>(
      "PreviewImport",
      keepCompleted.value || format.value === "json"
        ? content.value
        : content.value.replace(/^(\s*(?:[-*]\s+)?)\[x\](?=\s)/gim, "$1[ ]"),
      format.value,
      s.view === "inbox" ? null : s.date,
      replace.value,
    ),
  );
  if (p) preview.value = p;
}
async function getBatch() {
  const p = await s.run(() =>
    api<Preview>(
      "PreviewBatch",
      s.selected,
      props.action || "complete",
      s.snapshot.revision,
    ),
  );
  if (p) preview.value = p;
}
async function importFile() {
  const r = await s.run(() =>
    api<{ content: string; format: string } | null>("PickImport"),
  );
  if (r) {
    content.value = r.content;
    format.value = r.format;
    preview.value = null;
  }
}
async function clipboard() {
  const r = await s.run(() => api<string>("ClipboardRead"));
  if (r !== undefined) content.value = r;
}
function submitShortcut(e: KeyboardEvent) {
  if (e.isComposing || s.saving) return;
  if (props.mode === "export") void exportFile();
  else if (preview.value) void commit();
  else void getPreview();
}
async function commit() {
  if (!preview.value) return;
  let ok = false;
  await s.run(async () => {
    await api(
      props.mode === "batch" ? "CommitBatch" : "CommitImport",
      preview.value!.token,
    );
    ok = true;
  });
  if (ok) {
    s.selected = [];
    await s.metadata();
    s.inform(props.mode === "batch" ? "已完成批量操作" : "导入成功");
    emit("close");
  }
}
async function exportFile() {
  if (scope.value === "range" && from.value > until.value) {
    s.error = "结束日期不能早于开始日期";
    return;
  }
  const result = await s.run(() =>
    api<string>(
      "SaveExport",
      format.value,
      query(),
      includeInbox.value,
      scope.value === "selected" ? s.selected : [],
    ),
  );
  if (result) {
    s.inform("已导出到文件");
    emit("close");
  }
}
async function copy() {
  const text = await s.run(() =>
    api<string>(
      "Export",
      format.value,
      query(),
      includeInbox.value,
      scope.value === "selected" ? s.selected : [],
    ),
  );
  if (text !== undefined) {
    await s.run(() => api("ClipboardWrite", text));
    copyStatus.value = "已复制";
  }
}
</script>
<template>
  <section
    :inert="s.maintenance || undefined"
    ref="dialog"
    v-motion
    class="dialog transfer-dialog embedded-pane"
    @keydown.esc.stop.prevent="emit('close')"
    @keydown.meta.enter.prevent="submitShortcut"
    @keydown.ctrl.enter.prevent="submitShortcut"
  >
    <header class="dialog-header">
      <span>{{
        mode === "import"
          ? "导入待办"
          : mode === "export"
            ? "导出待办"
            : "确认批量操作"
      }}</span
      ><button class="icon-button" aria-label="关闭" @click="emit('close')">
        <X :size="18" />
      </button>
    </header>
    <div class="dialog-body">
      <template v-if="mode === 'import' && !preview"
        ><div class="inline-actions">
          <button class="button" @click="clipboard">
            <Clipboard :size="14" />从剪贴板</button
          ><button class="button" @click="importFile">
            <FolderOpen :size="14" />选择文件</button
          ><CustomSelect v-model="format" aria-label="导入格式">
            <option value="text">纯文本</option>
            <option value="markdown">Markdown 多级清单</option>
            <option value="json">完整 JSON</option>
          </CustomSelect>
        </div>
        <textarea
          v-model="content"
          class="import-content"
          placeholder="每行一件待办事项。&#10;Markdown 用每级 2 个空格缩进子任务。"
        /><label v-if="format === 'json'" class="check-label"
          ><input
            v-model="replace"
            type="checkbox"
          />整体替换现有数据（导入前自动保护备份）</label
        >
        <label v-if="format !== 'json'" class="check-label"
          ><input
            v-model="keepCompleted"
            type="checkbox"
          />保留清单中的完成状态</label
        >
        <p class="muted small">
          {{
            format === "json"
              ? "完整结构导入将校验父子关系与重复规则。"
              : "确认后添加到 " + (s.view === "inbox" ? "收集箱" : s.date)
          }}
          · 不会静默丢弃无效行。
        </p></template
      >
      <template v-if="preview"
        ><div class="preview-summary">
          <Check :size="20" />
          <div>
            <strong
              >{{ preview.roots }}
              {{ mode === "batch" ? "个所选分支" : "个根任务" }} ·
              {{ preview.nodes }} 个节点</strong
            >
            <p v-if="mode === 'batch'">
              本次将{{ batchVerb }}这些任务及其全部后代。{{
                action === "delete" ? "删除后无法逐条撤销。" : ""
              }}
            </p>
            <p v-else>请核对内容和影响范围后再确认。</p>
          </div>
        </div>
        <p v-for="w in preview.warnings" :key="w" class="muted small">
          {{ w }}
        </p>
        <div class="import-tree">
          <div
            v-for="t in previewNodes"
            :key="t.id"
            :style="{ paddingLeft: `${Math.min(depth(t), 8) * 14}px` }"
          >
            <span>{{ t.status === "completed" ? "☑" : "□" }}</span
            >{{ t.title
            }}<small v-if="!t.parentId" class="preview-date">{{
              t.date || "收集箱"
            }}</small>
          </div>
        </div>
        <p v-if="preview.replacing" class="inline-error">
          确认后替换当前数据。恢复需要使用完整备份。
        </p></template
      >
      <template v-if="mode === 'export'"
        ><label class="field"
          ><span>文件格式</span
          ><CustomSelect v-model="format">
            <option value="markdown">Markdown (.md)</option>
            <option value="text">纯文本 (.txt)</option>
            <option value="csv">表格 CSV (.csv)</option>
            <option value="html">离线网页 (.html)</option>
            <option value="json">完整 JSON (.json)</option>
          </CustomSelect></label
        ><label v-if="format !== 'json'" class="field"
          ><span>导出范围</span
          ><CustomSelect v-model="scope">
            <option value="current">当前视图</option>
            <option value="range">日期范围</option>
            <option value="all">全部任务</option>
            <option value="selected" :disabled="!s.selected.length">
              选中的 {{ s.selected.length }} 个根任务
            </option>
          </CustomSelect></label
        >
        <div v-if="scope === 'range' && format !== 'json'" class="field-grid">
          <label class="field"
            ><span>起始日期</span><input v-model="from" type="date" /></label
          ><label class="field"
            ><span>结束日期</span
            ><input v-model="until" type="date" :min="from"
          /></label>
        </div>
        <label v-if="format !== 'json'" class="check-label"
          ><input v-model="includeInbox" type="checkbox" />包含收集箱</label
        >
        <p class="muted small">
          {{
            format === "json"
              ? "JSON 始终保存完整领域数据和重复规则。"
              : "导出完整子任务层级，中文与 emoji 使用 UTF-8。可读清单不能代替完整数据库备份。"
          }}
        </p></template
      >
      <p v-if="s.error" class="inline-error" role="alert">{{ s.error }}</p>
    </div>
    <footer class="dialog-footer">
      <button
        class="button"
        @click="preview && mode === 'import' ? (preview = null) : emit('close')"
      >
        {{ preview && mode === "import" ? "返回编辑" : "取消" }}</button
      ><template v-if="mode === 'export'"
        ><button class="button" @click="copy">
          <Clipboard :size="14" />{{ copyStatus || "复制内容" }}</button
        ><button
          class="button primary"
          :disabled="s.saving"
          @click="exportFile"
        >
          <Download :size="14" />导出文件
        </button></template
      ><button
        v-else
        class="button primary"
        :class="{ destructive: action === 'delete' }"
        :disabled="s.saving || (!preview && !content.trim())"
        @click="preview ? commit() : getPreview()"
      >
        {{ preview ? confirmLabel : "预览导入" }}<ArrowRight :size="14" />
      </button>
    </footer>
  </section>
</template>
