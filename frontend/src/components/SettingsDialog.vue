<script setup lang="ts">
import CustomSelect from "./CustomSelect.vue";
import { restoreDialogFocus } from "../composables/dialogFocus";
import { ref, onMounted, onUnmounted } from "vue";
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
import { api, confirmAction, listen } from "../api";
import { type Backup } from "../api/types";
import { useTodo } from "../stores/todo";
const props = defineProps<{ initialTab?: string }>();
const s = useTodo(),
  emit = defineEmits<{ close: []; "window-opacity": [number | null] }>(),
  dialog = ref<HTMLElement>(),
  tab = ref(props.initialTab || "general"),
  backups = ref<Backup[]>([]),
  windowKey = ref(s.key === "quick" ? "main" : s.key),
  windowMode = ref("normal"),
  opacity = ref(1),
  bindings = ref<Record<string, string>>({ main: "", quick: "", mini: "" }),
  latestRelease = ref<{
    version: string;
    available: boolean;
    downloadURL: string;
  } | null>(null),
  checking = ref(false),
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
  await loadProxy();

  emoji.value = s.boot?.settings.emoji || "☺";
  for (const key of ["main", "quick", "mini"])
    bindings.value[key] =
      s.boot?.settings["shortcut" + key[0].toUpperCase() + key.slice(1)] || "";
  readWindow();
});
onUnmounted(() => emit("window-opacity", null));
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
type ProxyConfig = {
  mode: string;
  address: string;
  username: string;
  password: string;
  hasPassword?: boolean;
  keepPassword?: boolean;
};
const proxy = ref<ProxyConfig>({
  mode: "system",
  address: "",
  username: "",
  password: "",
});
const proxyLoaded = ref(false),
  proxySaving = ref(false),
  downloading = ref(false),
  downloadPath = ref(""),
  updateProgress = ref("");
let stopUpdate: (() => void) | undefined;
onMounted(async () => {
  stopUpdate = await listen(
    "todo:update-progress",
    (value) => (updateProgress.value = String(value)),
  );
});
onUnmounted(() => stopUpdate?.());
async function installUpdate() {
  if (!(await saveProxy())) return;
  if (
    !(await confirmAction(
      "升级将自动下载、校验并替换应用，然后重启。进行中的专注与提醒会短暂停止，任务数据保留。继续？",
    ))
  )
    return;
  downloading.value = true;
  updateProgress.value = "正在下载安装包…";
  try {
    await api("InstallUpdate");
    updateProgress.value = "正在安装，即将重启…";
  } catch (e) {
    s.error = String(e);
    updateProgress.value = "";
    downloading.value = false;
  }
}
async function loadProxy() {
  const result = await s.run(() => api<ProxyConfig>("UpdateProxy"));
  if (result) proxy.value = result;
  proxyLoaded.value = true;
}
async function saveProxy() {
  if (!proxyLoaded.value) return false;
  proxySaving.value = true;
  try {
    await api("SaveUpdateProxy", { ...proxy.value });
    proxy.value.hasPassword =
      !!proxy.value.password || !!proxy.value.keepPassword;
    proxy.value.keepPassword = proxy.value.hasPassword;
    proxy.value.password = "";
    s.error = "";
    return true;
  } catch (error) {
    s.error = String(error);
    return false;
  } finally {
    proxySaving.value = false;
  }
}
async function downloadUpdate() {
  if (!(await saveProxy())) return;
  downloading.value = true;
  try {
    const path = await s.run(() => api<string>("DownloadUpdate"));
    if (path) {
      downloadPath.value = path;
      s.inform("安装包已保存");
    }
  } finally {
    downloading.value = false;
  }
}
async function checkUpdate() {
  if (!(await saveProxy())) return;
  checking.value = true;
  latestRelease.value = null;
  try {
    const result = await s.run(() =>
      api<{ version: string; available: boolean; downloadURL: string }>(
        "CheckUpdate",
      ),
    );
    if (result) latestRelease.value = result;
  } finally {
    checking.value = false;
  }
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
    v-motion
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
      <div v-motion="tab" class="settings-content">
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
              :checked="s.boot?.settings.autoPostpone !== 'false'"
              @change="s.setting('autoPostpone', checkbox($event))" /></label
          ><label class="setting-row"
            ><span>完成项置底<small>在每个同级列表中保留完整子树</small></span
            ><input
              type="checkbox"
              :checked="s.boot?.settings.completedBottom === 'true'"
              @change="s.setting('completedBottom', checkbox($event))" /></label
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
                @input="emit('window-opacity', opacity)"
                aria-label="背景不透明度"
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
          ><img class="about-app-icon" src="/appicon.svg" alt="SimpleTodo" />
          <h2>SimpleTodo</h2>
          <p class="muted">版本 {{ s.boot?.version }} · 把每一天安排得刚刚好</p>
          <p class="about-copy">
            简洁地记录、安排与完成。所有任务数据仅存放在本地电脑上。
          </p>
          <h3>发布与更新</h3>
          <p class="muted small">
            从 GitHub Release 检查并下载适合当前电脑的安装包。
          </p>
          <fieldset
            class="update-proxy"
            :disabled="!proxyLoaded || checking || downloading || proxySaving"
          >
            <label class="setting-row"
              ><span>更新代理</span
              ><CustomSelect v-model="proxy.mode" aria-label="更新代理"
                ><option value="system">系统代理</option>
                <option value="none">无代理</option>
                <option value="custom">自定义代理</option></CustomSelect
              ></label
            >
            <div v-if="proxy.mode === 'custom'" v-motion class="proxy-fields">
              <label class="field"
                ><span>HTTP 代理地址</span
                ><input
                  v-model="proxy.address"
                  aria-label="HTTP 代理地址"
                  placeholder="http://127.0.0.1:7890"
                  autocomplete="off"
                  spellcheck="false"
              /></label>
              <div class="proxy-credentials">
                <label class="field"
                  ><span>用户名（可选）</span
                  ><input
                    v-model="proxy.username"
                    aria-label="代理用户名"
                    autocomplete="off"
                /></label>
                <label class="field"
                  ><span>密码（可选）</span
                  ><input
                    v-model="proxy.password"
                    type="password"
                    aria-label="代理密码"
                    :placeholder="
                      proxy.hasPassword && proxy.keepPassword
                        ? '已保存，留空保留'
                        : '无需认证可留空'
                    "
                    autocomplete="new-password"
                    @input="proxy.keepPassword = false"
                /></label>
              </div>
              <button
                v-if="proxy.hasPassword"
                class="text-button"
                @click="
                  proxy.password = '';
                  proxy.keepPassword = false;
                  proxy.hasPassword = false;
                "
              >
                清除已保存的密码
              </button>
            </div>
            <p class="muted small">
              仅用于检查更新和下载安装包。支持 HTTP 代理。
            </p>
            <button
              class="text-button"
              @click="
                saveProxy().then((ok) => {
                  if (ok) s.inform('更新代理已保存');
                })
              "
            >
              保存代理设置
            </button>
          </fieldset>
          <div class="inline-actions">
            <button
              class="button"
              :disabled="checking || downloading || !proxyLoaded"
              @click="checkUpdate"
            >
              <ExternalLink :size="14" />{{
                checking ? "检查中…" : "检查更新"
              }}</button
            ><button
              v-if="latestRelease?.available"
              class="button primary"
              :disabled="downloading || checking"
              @click="installUpdate"
            >
              {{ downloading ? "升级中…" : "升级并重启" }}
            </button>
            <button
              v-if="latestRelease?.available"
              class="button"
              :disabled="downloading || checking"
              @click="downloadUpdate"
            >
              {{ downloading ? "下载中…" : "仅下载安装包" }}
            </button>
          </div>
          <p
            v-for="warning in s.boot?.warnings.filter((w) =>
              w.startsWith('上次自动升级'),
            )"
            :key="warning"
            class="inline-error"
            role="alert"
          >
            {{ warning }}
          </p>
          <p v-if="updateProgress" class="muted small" role="status">
            {{ updateProgress }}
          </p>
          <p v-if="downloadPath" class="muted small path-output" role="status">
            已保存到 {{ downloadPath }}
          </p>
          <p v-if="latestRelease" class="muted small">
            {{
              latestRelease.available
                ? "发现新版本 " + latestRelease.version
                : "当前已是最新版本（" + latestRelease.version + "）"
            }}
          </p>
        </template>
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
