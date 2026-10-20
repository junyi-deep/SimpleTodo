<script setup lang="ts">
import CustomSelect from "./CustomSelect.vue";
import { restoreDialogFocus } from "../composables/dialogFocus";
import { ref, onMounted } from "vue";
import {
  X,
  Settings2,
  Palette,
  Database,
  Keyboard,
  Info,
  FolderOpen,
  Archive,
  RotateCcw,
  Bell,
  ExternalLink,
} from "lucide-vue-next";
import { api, confirmAction } from "../api";
import { type Backup } from "../api/types";
import { useTodo } from "../stores/todo";
const props = defineProps<{ initialTab?: string }>();
const s = useTodo(),
  emit = defineEmits<{ close: [] }>(),
  dialog = ref<HTMLElement>(),
  tab = ref(props.initialTab || "general"),
  backups = ref<Backup[]>([]),
  windowKey = ref(s.key === "quick" ? "main" : s.key),
  windowMode = ref("normal"),
  opacity = ref(1),
  bindings = ref<Record<string, string>>({ main: "", quick: "", mini: "" }),
  release = ref(""),
  emoji = ref("");
const tabs = [
  { id: "general", label: "常规", icon: Settings2 },
  { id: "appearance", label: "外观与窗口", icon: Palette },
  { id: "keys", label: "快捷键", icon: Keyboard },
  { id: "data", label: "本地数据", icon: Database },
  { id: "about", label: "关于", icon: Info },
];
restoreDialogFocus();
onMounted(async () => {
  await loadBackups();
  release.value = s.boot?.settings.releaseURL || "";
  emoji.value = s.boot?.settings.emoji || "☺";
  for (const key of ["main", "quick", "mini"])
    bindings.value[key] =
      s.boot?.settings["shortcut" + key[0].toUpperCase() + key.slice(1)] || "";
  readWindow();
});
function readWindow() {
  const w = s.boot?.windows.find((w) => w.key === windowKey.value);
  windowMode.value = w?.mode || (windowKey.value === "mini" ? "top" : "normal");
  opacity.value = w?.opacity || 1;
}
function checkbox(e: Event) {
  return (e.target as HTMLInputElement).checked ? "true" : "false";
}
async function set(key: string, e: Event) {
  await s.setting(key, (e.target as HTMLInputElement).value);
}
async function auto(e: Event) {
  await s.run(() => api("Autostart", (e.target as HTMLInputElement).checked));
  await s.metadata();
}
async function configure() {
  await s.run(() =>
    api(
      "WindowConfig",
      windowKey.value,
      windowMode.value,
      opacity.value,
      false,
    ),
  );
  await s.metadata();
}
async function binding(key: string) {
  await s.run(() => api("Shortcut", key, bindings.value[key]));
  await s.metadata();
  if (!s.error) s.inform("快捷键已生效");
}
async function loadBackups() {
  try {
    backups.value = await api<Backup[]>("Backups");
  } catch (e) {
    s.error = String(e);
  }
}
async function backup() {
  await s.run(() => api("Backup"));
  await loadBackups();
  if (!s.error) s.inform("完整备份已保存");
}
async function restore(name: string) {
  const info = await s.run(() =>
    api<{
      path: string;
      name: string;
      createdAt: number;
      roots: number;
      nodes: number;
    } | null>("PrepareRestore", name),
  );
  if (
    !info ||
    !(await confirmAction(
      `恢复「${info.name}」？\n备份时间：${new Date(info.createdAt).toLocaleString("zh-CN")}\n${info.roots} 个根任务，${info.nodes} 个节点。\n将整体替换当前数据，并先创建保护备份。`,
    ))
  )
    return;
  await s.run(() => api("RestorePrepared", info.path));
  await s.metadata();
  await loadBackups();
  if (!s.error) s.inform("数据已恢复");
}
async function migrate() {
  if (
    !(await confirmAction(
      "将暂停写入，校验并切换到新的本地目录，保留原副本。请勿选择网络共享目录。",
    ))
  )
    return;
  await s.run(() => api("Migrate"));
  await s.metadata();
  await loadBackups();
}
</script>
<template>
  <section
    :inert="s.maintenance || undefined"
    ref="dialog"
    class="dialog settings-dialog embedded-pane"
    @keydown.esc.stop.prevent="emit('close')"
  >
    <header class="dialog-header">
      <span>设置 · {{ tabs.find((t) => t.id === tab)?.label }}</span
      ><button class="icon-button" aria-label="关闭设置" @click="emit('close')">
        <X :size="18" />
      </button>
    </header>
    <div class="settings-layout">
      <nav aria-label="设置分类">
        <button
          v-for="t in tabs"
          :key="t.id"
          :class="{ active: tab === t.id }"
          :title="t.label"
          :aria-current="tab === t.id ? 'page' : undefined"
          @click="tab = t.id"
        >
          <component :is="t.icon" :size="16" />{{ t.label }}
        </button>
      </nav>
      <div class="settings-content">
        <template v-if="tab === 'general'"
          ><h2>按自己的节奏安排</h2>
          <p class="muted section-description"></p>
          <label class="setting-row"
            ><span>登录时启动<small>静默进入菜单栏或托盘</small></span
            ><input
              type="checkbox"
              :checked="s.boot?.autostart"
              @change="auto" /></label
          ><label class="setting-row"
            ><span
              >未完成任务自动延期<small
                >仅移动普通任务，重复任务不延期</small
              ></span
            ><input
              type="checkbox"
              :checked="s.boot?.settings.autoPostpone === 'true'"
              @change="s.setting('autoPostpone', checkbox($event))" /></label
          ><label class="setting-row"
            ><span>完成项置底<small>在每个同级列表中保留完整子树</small></span
            ><input
              type="checkbox"
              :checked="s.boot?.settings.completedBottom === 'true'"
              @change="s.setting('completedBottom', checkbox($event))" /></label
          ><label class="setting-row"
            ><span>点击任务标题</span
            ><CustomSelect
              :model-value="s.boot?.settings.titleClick"
              aria-label="点击任务标题"
              @change="set('titleClick', $event)"
            >
              <option value="edit">编辑任务</option>
              <option value="complete">切换完成</option>
            </CustomSelect></label
          ><label class="setting-row"
            ><span>Tab 连续录入<small>关闭时使用系统键盘导航</small></span
            ><input
              type="checkbox"
              :checked="s.boot?.settings.tabEntry === 'true'"
              @change="s.setting('tabEntry', checkbox($event))" /></label
          ><label class="setting-row"
            ><span>提醒声音</span
            ><input
              type="checkbox"
              :checked="s.boot?.settings.sound === 'true'"
              @change="s.setting('sound', checkbox($event))" /></label
          ><label class="setting-row"
            ><span>默认专注分钟数</span
            ><input
              class="small-input"
              type="number"
              min="1"
              max="180"
              :model-value="s.boot?.settings.focusMinutes"
              @change="set('focusMinutes', $event)"
          /></label>
          <div class="setting-row">
            <span
              >系统通知<small>{{
                s.boot?.notificationAllowed
                  ? "已获得通知权限"
                  : "未授权或系统已关闭"
              }}</small></span
            ><button
              class="button"
              @click="s.run(() => api('Permission')).then(() => s.metadata())"
            >
              <Bell :size="14" />申请权限
            </button>
          </div>
          <button
            v-if="!s.boot?.notificationAllowed"
            class="text-button"
            @click="s.run(() => api('OpenNotificationSettings'))"
          >
            打开系统通知设置
          </button>
          <p class="quiet-note">
            关闭主窗口会继续在后台运行；明确退出后停止提醒调度。
          </p></template
        >
        <template v-if="tab === 'appearance'"
          ><h2>轻一点，清楚一点</h2>
          <label class="setting-row"
            ><span>主题</span
            ><CustomSelect
              :model-value="s.boot?.settings.theme"
              aria-label="主题"
              @change="set('theme', $event)"
            >
              <option value="system">跟随系统</option>
              <option value="light">浅色</option>
              <option value="dark">深色</option>
            </CustomSelect></label
          ><label class="setting-row"
            ><span>动态效果<small>始终遵从系统减少动态效果设置</small></span
            ><input
              type="checkbox"
              :checked="s.boot?.settings.animation === 'true'"
              @change="s.setting('animation', checkbox($event))"
          /></label>
          <h3>窗口显示</h3>
          <div class="field-grid">
            <label class="field"
              ><span>显示层级</span
              ><CustomSelect v-model="windowMode">
                <option value="normal">普通窗口</option>
                <option value="top">始终置顶</option>
                <option
                  value="desktop"
                  :disabled="!s.boot?.capabilities.desktopMode"
                >
                  桌面模式 / 置底
                </option>
              </CustomSelect></label
            ><label class="field"
              ><span>背景不透明度 {{ Math.round(opacity * 100) }}%</span
              ><input
                v-model.number="opacity"
                type="range"
                min="0.4"
                max="1"
                step="0.05"
            /></label>
          </div>
          <button class="button" @click="configure">应用窗口设置</button
          ><button
            class="text-button"
            @click="s.run(() => api('ResetWindows'))"
          >
            <RotateCcw :size="13" />重置窗口位置
          </button>
          <p class="quiet-note">
            背景透明度保留文字可读性。桌面模式在系统安全界面和全屏应用下的行为受系统限制。
          </p></template
        >
        <template v-if="tab === 'keys'"
          ><h2>少一点切换</h2>
          <p class="muted section-description">
            全局快捷键在应用后台运行时同样生效。留空可取消绑定；系统冲突时保留原配置。
          </p>
          <label
            v-for="(label, key) in {
              main: '显示 / 隐藏主窗口',
              quick: '快速新增任务',
            }"
            :key="key"
            class="field"
            ><span>{{ label }}</span>
            <div class="inline-actions">
              <input
                v-model="bindings[key]"
                placeholder="如 CmdOrCtrl+Shift+J"
                :aria-label="label"
              /><button class="button" @click="binding(key)">应用</button>
            </div></label
          >
          <h3>应用内操作</h3>
          <dl class="key-help">
            <dt>新增任务</dt>
            <dd>⌘ / Ctrl + N</dd>
            <dt>添加任意层级子任务</dt>
            <dd>Shift + Enter</dd>
            <dt>缩进 / 提升子树</dt>
            <dd>⌘ / Ctrl + ] / [</dd>
            <dt>编辑 / 切换完成</dt>
            <dd>Enter / Space</dd>
            <dt>展开 / 折叠 / 导航</dt>
            <dd>方向键</dd>
            <dt>复制 / 粘贴（月历）</dt>
            <dd>⌘ / Ctrl + C / V</dd>
            <dt>回到今天</dt>
            <dd>⌘ / Ctrl + T</dd>
            <dt>保存详情</dt>
            <dd>⌘ / Ctrl + Enter</dd>
          </dl></template
        >
        <template v-if="tab === 'data'"
          ><h2>数据由你掌握</h2>
          <p class="muted section-description">
            完整备份用于恢复全部规则与任务树。
          </p>
          <label class="field"
            ><span>当前数据目录</span
            ><output class="path-output">{{ s.boot?.dataDir }}</output></label
          >
          <div class="inline-actions">
            <button class="button" @click="s.run(() => api('OpenDataDir'))">
              <FolderOpen :size="14" />打开目录</button
            ><button class="button" @click="migrate">迁移目录</button>
          </div>
          <h3>完整数据库备份</h3>
          <div class="inline-actions">
            <button class="button primary" @click="backup">
              <Archive :size="14" />立即备份</button
            ><button class="button" @click="restore('')">从文件恢复</button>
          </div>
          <p class="muted small">
            每日首次启动自动备份，保留最近 7 份。恢复前另建保护备份。
          </p>
          <div class="backup-list">
            <p v-if="!backups.length" class="muted">还没有备份。</p>
            <div v-for="b in backups" :key="b.name">
              <span
                ><strong>{{
                  new Date(b.createdAt).toLocaleString("zh-CN")
                }}</strong
                ><small
                  >{{
                    b.name.startsWith("protect")
                      ? "保护备份"
                      : b.name.startsWith("daily")
                        ? "每日备份"
                        : "手动备份"
                  }}
                  · {{ (b.size / 1024).toFixed(0) }} KB</small
                ></span
              ><button class="button" @click="restore(b.name)">恢复</button>
            </div>
          </div>
          <h3>节假日</h3>
          <p class="muted small">
            导入 JSON 数组，字段为 date、name、kind（holiday/workday）、source
            和 version。未配置年份仅显示周末，不推断调休。
          </p>
          <button class="button" @click="s.run(() => api('ImportHolidays'))">
            导入本地节假日数据
          </button></template
        >
        <template v-if="tab === 'about'"
          ><div class="about-mark"><CheckMark /></div>
          <h2>SimpleTodo</h2>
          <p class="muted">版本 {{ s.boot?.version }} · 把每一天安排得刚刚好</p>
          <p class="about-copy">
            从一个想法开始，拆成可执行的步骤。按天规划、桌面待办、专注计时，所有任务和数据都留在你的电脑上。
          </p>
          <h3>发布与更新</h3>
          <label class="field"
            ><span>项目发布页（HTTPS）</span
            ><input v-model="release" placeholder="尚未配置发布页"
          /></label>
          <div class="inline-actions">
            <button class="button" @click="s.setting('releaseURL', release)">
              保存地址</button
            ><button class="button" @click="s.run(() => api('CheckUpdate'))">
              <ExternalLink :size="14" />手动检查更新
            </button>
          </div>
          <h3>使用提示</h3>
          <p class="muted small">
            Enter 连续记录，Shift+Enter
            拆分子任务；从任务菜单可调整层级。拖动根任务到日期栏即可改期，收集箱保存还没安排日期的想法。
          </p>
          <p class="quiet-note">
            由 Go、Wails、Vue 和 SQLite 构建。运行不依赖远程字体或在线服务。
          </p></template
        >
        <p v-if="s.error" class="inline-error" role="alert">{{ s.error }}</p>
      </div>
    </div>
  </section>
</template>
<script lang="ts">
import { defineComponent, h } from "vue";
const CheckMark = defineComponent({
  render: () =>
    h(
      "svg",
      {
        viewBox: "0 0 24 24",
        width: 28,
        height: 28,
        fill: "none",
        stroke: "currentColor",
        "stroke-width": 2,
      },
      [h("path", { d: "m5 12 4 4L19 6" })],
    ),
});
export default { components: { CheckMark } };
</script>
