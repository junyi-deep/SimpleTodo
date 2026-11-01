<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import { animate, inView, stagger } from "motion";
import {
  ArrowDown,
  ArrowRight,
  ArrowUpRight,
  Archive,
  CalendarDays,
  Check,
  ChartNoAxesCombined,
  ChevronDown,
  Database,
  FileText,
  Focus,
  Github,
  HardDrive,
  Keyboard,
  Laptop,
  Layers,
  ListTodo,
  Menu,
  Monitor,
  ShieldCheck,
  WifiOff,
  X,
} from "lucide-vue-next";

const base = import.meta.env.BASE_URL;
const asset = (file: string) => `${base}${file}`;
const github = "https://github.com/junyi-deep/SimpleTodo";
const release = `${github}/releases/latest`;
const downloads = {
  mac: `${release}/download/SimpleTodo-macos-arm64.zip`,
  windows: `${release}/download/SimpleTodo-windows-amd64.zip`,
};
const menuOpen = ref(false),
  menuButton = ref<HTMLButtonElement>();
const platform = ref<"mac" | "windows" | null>(null);
const downloadUrl = computed(() =>
  platform.value ? downloads[platform.value] : "#download",
);
const views = [
  {
    id: "list",
    label: "任务清单",
    icon: ListTodo,
    image: "list-dark.png",
    title: "记下想法，就开始行动。",
    description:
      "单击编辑，Enter 连续记录。把大任务拆成最多五层的小步骤，让下一步一目了然。",
    details: ["日计划与收集箱", "多级子任务", "标签与提醒"],
  },
  {
    id: "calendar",
    label: "日历规划",
    icon: CalendarDays,
    image: "calendar-dark.png",
    title: "给每一件事，合适的位置。",
    description:
      "在月历、周历与甘特图里安排时间。拖动即可改期，选中日期就能在侧栏处理当天任务。",
    details: ["拖动调整时间", "月／周／甘特图", "日期任务抽屉"],
  },
  {
    id: "focus",
    label: "专注时钟",
    icon: Focus,
    image: "focus-dark.png",
    title: "这一刻，只做好一件事。",
    description:
      "搜索选择要专注的任务，留出一段完整的时间。暂停、继续，按照自己的节奏前进。",
    details: ["搜索关联任务", "自由专注", "事项倒计时"],
  },
  {
    id: "review",
    label: "复盘笔记",
    icon: ChartNoAxesCombined,
    image: "notes-dark.png",
    title: "看见进展，也记下思考。",
    description:
      "用统计、热力图与任务报告回顾进展。把经验写进本地 Markdown 笔记，留下自己的答案。",
    details: ["状态与标签筛选", "完成热力图", "Markdown 笔记"],
  },
];
const activeView = ref(0),
  previewImage = ref<HTMLElement>();
const currentView = computed(() => views[activeView.value]);
const shortcuts = [
  {
    keys: ["Enter"],
    name: "接着记下一件事",
    detail: "保存当前任务，立即创建新的同级任务。",
  },
  {
    keys: ["Shift", "Enter"],
    name: "把事情拆成小步骤",
    detail: "为当前任务添加子任务，让行动更具体。",
  },
  {
    keys: ["Shift", "Tab"],
    name: "让层级随想法调整",
    detail: "将子任务往上提升一级，带上它的全部步骤。",
  },
];
const shortcut = ref(0);
const questions = [
  {
    question: "数据会存在哪里？",
    answer:
      "任务、标签、笔记和设置都保存在本机 SQLite 数据库中。无需注册账号，也不需要把任务上传到服务器。",
  },
  {
    question: "断网时还能使用吗？",
    answer:
      "任务管理、日历、专注和笔记都可离线使用。版本检查与安装包下载会访问 GitHub，你的任务与笔记始终留在本机。",
  },
  {
    question: "如何备份或换一台电脑？",
    answer:
      "在设置的“本地数据”中创建或恢复完整备份，也支持导入与导出任务。自动、手动和保护备份合计保留最近 3 次。",
  },
  {
    question: "支持哪些系统？",
    answer:
      "当前提供 macOS Apple Silicon 和 Windows 64 位安装包。macOS 解压后将应用放进“应用程序”；Windows 解压后运行应用。当前发布包未配置开发者证书签名。",
  },
  {
    question: "能同时打开多个工作窗口吗？",
    answer:
      "可以。列表、日历、专注、设置和复盘各自使用独立窗口，并分别记住大小、位置和最大化状态。",
  },
];
const reduced = matchMedia("(prefers-reduced-motion: reduce)");
let cleanupMotion = () => {};
function setupMotion() {
  cleanupMotion();
  if (reduced.matches) return;
  const controls: ReturnType<typeof animate>[] = [];
  controls.push(
    animate(
      ".hero-copy > *",
      { opacity: [0.55, 1], y: [18, 0] },
      { duration: 0.8, delay: stagger(0.08), ease: [0.16, 1, 0.3, 1] },
    ),
  );
  controls.push(
    animate(
      ".hero-product",
      { y: [28, 0], opacity: [0.5, 1] },
      { duration: 1.1, delay: 0.12, ease: [0.16, 1, 0.3, 1] },
    ),
  );
  const stop = inView(
    "[data-reveal]",
    (element) => {
      controls.push(
        animate(
          element,
          { opacity: [0.6, 1], y: [24, 0] },
          { duration: 0.7, ease: [0.16, 1, 0.3, 1] },
        ),
      );
    },
    { amount: 0.15 },
  );
  cleanupMotion = () => {
    stop();
    controls.forEach((control) => control.complete());
  };
}
watch(activeView, async () => {
  await nextTick();
  if (!reduced.matches && previewImage.value)
    animate(
      previewImage.value,
      { opacity: [0.55, 1], y: [10, 0] },
      { duration: 0.35, ease: [0.16, 1, 0.3, 1] },
    );
});
async function tabKey(event: KeyboardEvent) {
  if (!["ArrowRight", "ArrowLeft", "Home", "End"].includes(event.key)) return;
  event.preventDefault();
  activeView.value =
    event.key === "Home"
      ? 0
      : event.key === "End"
        ? views.length - 1
        : (activeView.value +
            (event.key === "ArrowRight" ? 1 : -1) +
            views.length) %
          views.length;
  await nextTick();
  document
    .getElementById(`view-${currentView.value.id}`)
    ?.focus({ preventScroll: true });
}
function closeMenu(event?: KeyboardEvent) {
  menuOpen.value = false;
  if (event) {
    event.preventDefault();
    menuButton.value?.focus();
  }
}
function outside(event: PointerEvent) {
  if (
    menuOpen.value &&
    event.target instanceof Element &&
    !event.target.closest(".site-header")
  )
    closeMenu();
}
onMounted(() => {
  const userPlatform = navigator.platform.toLowerCase();
  platform.value = userPlatform.includes("mac")
    ? "mac"
    : userPlatform.includes("win")
      ? "windows"
      : null;
  setupMotion();
  reduced.addEventListener("change", setupMotion);
  document.addEventListener("pointerdown", outside);
});
onUnmounted(() => {
  cleanupMotion();
  reduced.removeEventListener("change", setupMotion);
  document.removeEventListener("pointerdown", outside);
});
</script>

<template>
  <a class="skip-link" href="#main">跳到主要内容</a>
  <header class="site-header" @keydown.esc="closeMenu($event)">
    <div class="nav-shell">
      <a class="brand" :href="base" aria-label="SimpleTodo 首页"
        ><img :src="asset('icon.svg')" alt="" width="32" height="32" /><span
          >SimpleTodo</span
        ></a
      >
      <nav class="desktop-nav" aria-label="主要导航">
        <a href="#features">功能</a><a href="#local">本地优先</a
        ><a href="#faq">常见问题</a>
      </nav>
      <div class="nav-actions">
        <a
          class="github-link"
          :href="github"
          target="_blank"
          rel="noreferrer"
          aria-label="在 GitHub 查看 SimpleTodo"
          ><Github :size="19" /></a
        ><a
          class="button button-small button-white nav-download"
          href="#download"
          >下载 SimpleTodo<ArrowDown :size="14" /></a
        ><button
          ref="menuButton"
          class="menu-button"
          :aria-expanded="menuOpen"
          aria-controls="mobile-nav"
          :aria-label="menuOpen ? '关闭导航菜单' : '打开导航菜单'"
          @click="menuOpen = !menuOpen"
        >
          <X v-if="menuOpen" :size="22" /><Menu v-else :size="22" />
        </button>
      </div>
    </div>
    <nav
      v-if="menuOpen"
      id="mobile-nav"
      class="mobile-nav"
      aria-label="移动端导航"
    >
      <a href="#features" @click="closeMenu()">功能</a
      ><a href="#local" @click="closeMenu()">本地优先</a
      ><a href="#faq" @click="closeMenu()">常见问题</a
      ><a href="#download" @click="closeMenu()"
        >下载 SimpleTodo<ArrowDown :size="16"
      /></a>
    </nav>
  </header>

  <main id="main">
    <section class="hero" aria-labelledby="hero-title">
      <img
        class="hero-light"
        :src="asset('images/blue-prism.webp')"
        alt=""
        aria-hidden="true"
        width="1659"
        height="948"
        fetchpriority="high"
      />
      <div class="hero-copy container">
        <h1 id="hero-title">少一点纷扰，<br /><span>多一点完成。</span></h1>
        <p>
          完全本地的极简待办。<br
            class="mobile-break"
          />记下、安排、专注、复盘，<br
            class="desktop-break"
          />都在自己的节奏里。
        </p>
        <div class="hero-actions">
          <a class="button button-white" :href="downloadUrl"
            >下载 SimpleTodo<ArrowDown :size="17" /></a
          ><a class="button button-quiet" href="#features"
            >浏览功能<ArrowRight :size="17"
          /></a>
        </div>
      </div>
      <div
        class="hero-product container"
        aria-label="SimpleTodo 的实际应用界面，使用演示任务"
      >
        <img
          class="hero-calendar"
          :src="asset('images/calendar-dark.png')"
          alt="SimpleTodo 日历与日期任务抽屉，演示数据"
          width="2160"
          height="1440"
          fetchpriority="high"
        />
        <img
          class="hero-list"
          :src="asset('images/list-portrait.png')"
          alt="SimpleTodo 多级任务清单，演示数据"
          width="960"
          height="1440"
          fetchpriority="high"
        />
      </div>
    </section>

    <section
      id="features"
      class="features section-space container"
      aria-labelledby="features-title"
    >
      <div class="section-intro" data-reveal>
        <h2 id="features-title">
          想做的事，<br /><span>都有自己的位置。</span>
        </h2>
        <p>
          从随手记下的灵感，到有条不紊的计划。<br />你需要的工具，在需要时恰好出现。
        </p>
      </div>
      <div
        class="feature-tabs"
        role="tablist"
        aria-label="查看 SimpleTodo 功能"
        @keydown="tabKey"
      >
        <button
          v-for="(view, index) in views"
          :id="`view-${view.id}`"
          :key="view.id"
          role="tab"
          :aria-selected="activeView === index"
          aria-controls="feature-panel"
          :tabindex="activeView === index ? 0 : -1"
          @click="activeView = index"
        >
          <component :is="view.icon" :size="17" />{{ view.label }}
        </button>
      </div>
      <div
        id="feature-panel"
        class="feature-panel"
        role="tabpanel"
        :aria-labelledby="`view-${currentView.id}`"
      >
        <div class="feature-description">
          <h3>{{ currentView.title }}</h3>
          <p>{{ currentView.description }}</p>
          <div class="feature-details">
            <span v-for="detail in currentView.details" :key="detail"
              ><Check :size="13" />{{ detail }}</span
            >
          </div>
        </div>
        <div ref="previewImage" class="feature-image">
          <img
            :src="asset(`images/${currentView.image}`)"
            :alt="`${currentView.label}实际应用截图，使用演示数据`"
            width="2160"
            height="1440"
            loading="lazy"
          />
        </div>
      </div>
      <p class="image-caption">以上均为实际应用界面，任务内容为演示数据。</p>
    </section>

    <section
      id="local"
      class="local-section section-space container"
      aria-labelledby="local-title"
    >
      <div class="local-heading" data-reveal>
        <ShieldCheck :size="30" />
        <h2 id="local-title">
          属于你的数据，<br /><span>留在你的电脑里。</span>
        </h2>
        <p>无需账号，打开就能用。<br />任务、笔记与备份，由你自己掌握。</p>
      </div>
      <div class="local-principles">
        <div
          class="local-diagram"
          aria-label="任务数据、笔记与备份都保存在本机"
        >
          <div class="device">
            <img :src="asset('icon.svg')" alt="" width="64" height="64" /><span
              >你的电脑</span
            ><HardDrive :size="19" />
          </div>
          <div class="local-files">
            <span><Database :size="18" />任务数据</span
            ><span><FileText :size="18" />本地笔记</span
            ><span><Archive :size="18" />完整备份</span>
          </div>
        </div>
        <div class="local-copy">
          <article>
            <WifiOff :size="20" />
            <div>
              <h3>断网，也不打断思路。</h3>
              <p>
                任务管理、日历与笔记完全离线可用。没有登录流程，随时回到自己的事情上。
              </p>
            </div>
          </article>
          <article>
            <Archive :size="20" />
            <div>
              <h3>随时备份，安心保留。</h3>
              <p>
                自动、手动和保护备份合计保留最近 3
                次。支持导入与导出，数据可以带着走。
              </p>
            </div>
          </article>
          <a class="text-link" :href="github" target="_blank" rel="noreferrer"
            >公开代码，细节可查<ArrowUpRight :size="16"
          /></a>
        </div>
      </div>
    </section>

    <section
      class="flow-section section-space container"
      aria-labelledby="flow-title"
    >
      <div class="section-intro" data-reveal>
        <h2 id="flow-title">让记录顺手，<br /><span>让注意力完整。</span></h2>
      </div>
      <div class="flow-grid">
        <article class="keyboard-feature" data-reveal>
          <Keyboard :size="22" />
          <h3>跟得上想法的速度。</h3>
          <p>手不用离开键盘，<br />下一步就已经准备好。</p>
          <div class="shortcut-list" aria-label="体验任务快捷键说明">
            <button
              v-for="(item, index) in shortcuts"
              :key="item.name"
              :class="{ active: shortcut === index }"
              :aria-pressed="shortcut === index"
              @click="shortcut = index"
            >
              <span class="key-combination"
                ><kbd v-for="key in item.keys" :key="key">{{ key }}</kbd></span
              ><span>{{ item.name }}</span>
            </button>
          </div>
          <p class="shortcut-description" aria-live="polite">
            {{ shortcuts[shortcut].detail }}
          </p>
        </article>
        <article class="windows-feature" data-reveal>
          <div class="windows-copy">
            <Layers :size="22" />
            <h3>每个窗口，各就各位。</h3>
            <p>
              清单、日历、专注、设置与复盘独立打开。<br />各自记住大小与位置，适应你的桌面。
            </p>
          </div>
          <div class="windows-stage">
            <img
              class="window-calendar"
              :src="asset('images/calendar-dark.png')"
              alt="独立日历窗口，演示数据"
              width="2160"
              height="1440"
              loading="lazy"
            /><img
              class="window-focus"
              :src="asset('images/focus-dark.png')"
              alt="独立专注窗口，演示数据"
              width="1600"
              height="1200"
              loading="lazy"
            />
          </div>
        </article>
      </div>
    </section>

    <section
      id="faq"
      class="faq-section section-space container"
      aria-labelledby="faq-title"
    >
      <div class="section-intro" data-reveal>
        <h2 id="faq-title">还有一点，<br /><span>你可能想知道。</span></h2>
      </div>
      <div class="faq-list">
        <details
          v-for="(item, index) in questions"
          :key="item.question"
          :open="index === 0"
        >
          <summary>{{ item.question }}<ChevronDown :size="19" /></summary>
          <p>{{ item.answer }}</p>
        </details>
      </div>
    </section>

    <section
      id="download"
      class="download-section section-space container"
      aria-labelledby="download-title"
    >
      <img
        class="download-logo"
        :src="asset('icon.svg')"
        width="64"
        height="64"
        alt="SimpleTodo"
      />
      <h2 id="download-title" data-reveal>
        把下一件小事，<br /><span>变成一个开始。</span>
      </h2>
      <p>SimpleTodo，回到简单的待办。</p>
      <div class="download-platforms">
        <article>
          <Laptop :size="23" />
          <h3>macOS</h3>
          <span>Apple Silicon</span
          ><a
            class="button button-white"
            :href="downloads.mac"
            aria-label="下载 SimpleTodo macOS Apple Silicon 版"
            >下载 SimpleTodo<ArrowDown :size="16"
          /></a>
        </article>
        <article>
          <Monitor :size="23" />
          <h3>Windows</h3>
          <span>64 位</span
          ><a
            class="button button-white"
            :href="downloads.windows"
            aria-label="下载 SimpleTodo Windows 64 位版"
            >下载 SimpleTodo<ArrowDown :size="16"
          /></a>
        </article>
      </div>
      <a
        class="text-link release-link"
        :href="release"
        target="_blank"
        rel="noreferrer"
        >版本记录与其他下载<ArrowUpRight :size="15"
      /></a>
    </section>
  </main>
  <footer class="site-footer container">
    <a class="brand" :href="base"
      ><img :src="asset('icon.svg')" alt="" width="27" height="27" /><span
        >SimpleTodo</span
      ></a
    >
    <p>专心做事，也认真生活。</p>
    <nav aria-label="页脚导航">
      <a :href="github" target="_blank" rel="noreferrer"
        >GitHub<ArrowUpRight :size="13" /></a
      ><a :href="`${github}/issues`" target="_blank" rel="noreferrer"
        >反馈建议<ArrowUpRight :size="13"
      /></a>
    </nav>
  </footer>
</template>
