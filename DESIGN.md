---
name: SimpleTodo
description: "本地桌面待办工具的现行界面设计记录"
colors:
  surface: "#fcfdfb"
  secondary: "#f3f5f1"
  hover: "#eef2ec"
  text: "#27372d"
  muted: "#68736b"
  quiet: "#657168"
  border: "#e6ebe3"
  accent: "#4e7055"
  accent-strong: "#3f6147"
  accent-soft: "#e7eee4"
  danger: "#b74848"
  surface-dark: "#202720"
  secondary-dark: "#252e26"
  hover-dark: "#303b31"
  text-dark: "#e0e8de"
  muted-dark: "#a1afa0"
  quiet-dark: "#9dac9b"
  border-dark: "#354036"
  accent-dark: "#a5c49e"
  accent-strong-dark: "#bddab5"
  accent-soft-dark: "#354832"
  danger-dark: "#f18d87"
  task-red: "#c45353"
  task-orange: "#c7823c"
  task-yellow: "#b69b30"
  task-green: "#528369"
  task-blue: "#5c87bb"
  task-purple: "#9377b6"
  task-none: "transparent"
typography:
  headline:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "22px"
    fontWeight: 600
    letterSpacing: "-0.03em"
  title:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "18px"
    fontWeight: 600
    letterSpacing: "-0.02em"
  section:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "14px"
    fontWeight: 600
  body:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "14px"
  task:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "13px"
    lineHeight: 1.65
  task-mini:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "12px"
    lineHeight: 1.65
  button:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "12px"
    fontWeight: 500
  label:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "11px"
  small:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "11px"
    lineHeight: 1.6
  support:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "10px"
  footer:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "9px"
  clock:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "62px"
    fontWeight: 300
    lineHeight: 1.4
    letterSpacing: "-0.04em"
rounded:
  badge: "4px"
  check: "5px"
  control: "6px"
  field: "7px"
  entry: "8px"
  menu: "9px"
  dialog: "12px"
  circle: "50%"
spacing:
  "2": "2px"
  "3": "3px"
  "4": "4px"
  "5": "5px"
  "6": "6px"
  "7": "7px"
  "8": "8px"
  "9": "9px"
  "10": "10px"
  "12": "12px"
  "13": "13px"
  "14": "14px"
  "15": "15px"
  "16": "16px"
  "17": "17px"
  "18": "18px"
  "20": "20px"
  "21": "21px"
  "23": "23px"
  "24": "24px"
  "25": "25px"
  "26": "26px"
  "27": "27px"
  "35": "35px"
components:
  button-primary:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.surface}"
    typography: "{typography.button}"
    rounded: "{rounded.field}"
    padding: "8px 12px"
  button-primary-hover:
    backgroundColor: "{colors.accent-strong}"
  button-secondary:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    typography: "{typography.button}"
    rounded: "{rounded.field}"
    padding: "8px 12px"
  button-destructive:
    backgroundColor: "{colors.danger}"
    textColor: "{colors.surface}"
    typography: "{typography.button}"
    rounded: "{rounded.field}"
    padding: "8px 12px"
  field:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    rounded: "{rounded.field}"
    padding: "9px 10px"
  tab-active:
    backgroundColor: "{colors.accent-soft}"
    textColor: "{colors.accent}"
    rounded: "{rounded.control}"
    padding: "9px 4px"
  notice-card:
    backgroundColor: "{colors.secondary}"
    rounded: "{rounded.entry}"
    padding: "14px"
  task-row:
    rounded: "{rounded.control}"
    padding: "6px 3px 6px calc(var(--depth) * 16px + 1px)"
  add-task:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.accent}"
    rounded: "{rounded.entry}"
    padding: "9px 10px"
  calendar-day:
    padding: "7px 8px"
  dialog:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    rounded: "{rounded.dialog}"
    width: "min(540px, calc(100vw - 28px))"
---
# Design System: SimpleTodo

## Overview

SimpleTodo 是 macOS 优先、兼顾 Windows 的本地桌面待办工具。界面以紧凑任务树、日期选择和连续录入为中心：温和近白或深绿灰底色、低饱和绿色强调、系统字体、细分隔线以及 Lucide 线性图标。文字与状态承担信息，空间服务于频繁勾选、拆分、改期和新增。

本文从已有代码扫描记录现状，不建立新的视觉方向，也不表示截图或原生能力已通过验收。颜色和组件原语以前置 YAML 为准，浅色为无后缀默认值，深色使用对应 `-dark` 值；JSON sidecar 保存不能放入前置 token 的交互、阴影和示例。

**Key Characteristics:**
- 桌面操作工具的紧凑密度与清楚任务状态。
- 浅深主题共享语义；系统字体与本地 Lucide 图标。
- 主窗口、总览、浮窗、快速录入各有固定用途。
- 任务、月历、设置内容分区滚动，新增入口保持可达。

## Colors

颜色按功能而非装饰使用。默认颜色和深色配对的精确值见 frontmatter；颜色源为 `frontend/src/style.css`，任务颜色源为 `frontend/src/api/types.ts`。

### Primary

- **低饱和绿色**（`accent` / `accent-dark`）：主要动作、完成勾选、选中日期、焦点描边和关键数字。
- **更强的绿色**（`accent-strong` 配对）：主按钮悬停。
- **浅绿状态面**（`accent-soft` 配对）：活动页签、日期范围、批量操作条、专注圆形状态面与文本选择。

### Neutral

- **主表面**（`surface` 配对）：窗口、字段、菜单、对话框；主按钮使用它作为文字色。
- **次级表面**（`secondary` 配对）：设置侧栏、任务悬停、提示容器、空状态图形底面。
- **悬停表面**（`hover` 配对）：普通按钮、图标按钮、菜单项与未选日期。
- **正文**（`text` 配对）、**次要文字**（`muted` 配对）、**安静标签**（`quiet` 配对）：分配到任务标题、说明与结构标签。不要把后两者互相替换为任意浅灰。
- **边界**（`border` 配对）：细分隔线、字段边框、日历网格和滚动条。

### Semantic status and task labels

`danger` 配对只表达危险操作、逾期或错误。`task-red`、`task-orange`、`task-yellow`、`task-green`、`task-blue`、`task-purple` 与 `task-none` 是任务标记，保持跨主题原值，用于根任务圆点和编辑器色样；它们不是新的主色体系。

## Typography

所有层级使用同一系统字体栈；无远程字体、独立展示字体或专用等宽字体。`kbd` 继承字体，日期、进度与时钟数字使用 tabular-nums。

### Hierarchy

- **Headline**：主日期标题（22px / 600 / -0.03em），窄窗口降为（20px）。
- **Title**：二级标题（18px / 600 / -0.02em）；编辑器标题字段另为（18px / 550）。
- **Section**：常规三级标题（14px / 600）；设置页标题为（17px）。
- **Body / Task**：根字号（14px），任务标题（13px / 1.65），浮窗任务标题（12px / 1.65）。
- **Button / Label**：标准按钮（12px / 500）；字段标签、页签与日期标签多为（11px）。辅助信息（10px），页脚（9px），小字说明（11px / 1.6）。
- **Clock**：专注计时（62px / 300 / 1.4 / -0.04em），只在计时界面使用。

任务标题可在任意位置换行；月历任务使用单行省略。没有正文最大行长规则。不要由现有尺寸推导未经实现的固定比例字阶。

## Layout

外壳占满窗口（100dvh），纵向 flex，隐藏外壳溢出。头部、日期、工具条和底部录入围绕可伸缩任务列表组织；任务列表（min-height: 0、overflow-y: auto、overscroll-behavior: contain）独立滚动。专注页另行滚动。

| 原生窗口 | 新建默认尺寸 | 原生最小尺寸 | 组织方式 |
| --- | --- | --- | --- |
| 主窗口 | 460×720 | 360×480 | 日期、视图、周条、任务树、底部新增与状态 |
| 月度总览 | 1080×760 | 840×600 | 42 日网格和选中日期/范围的任务明细 |
| 迷你浮窗 | 300×420 | 240×62 | 今日待办、快速勾选、紧凑录入；折叠高度 62px |
| 快速记录 | 420×230 | 360×200 | 单一标题字段与日期/收集箱选择、记录按钮 |

这些默认和最小尺寸来自 `todoservice.go`，保存的窗口状态会覆盖默认尺寸。评审方向文档另使用主窗口（460×760）与紧凑检查视口（360×600）；这是截图检查尺寸，并非代码中的原生默认/最小值。

主界面常用横向内边距（25px），任务列表（17px）；日期区域（20px 25px 0），工具条（19px 25px 12px）。主窗口宽度达到（700px）时居中，最大宽度（640px），增加两侧细边界；日期与录入横向间距（35px），任务列表（27px）。

宽度不超过（420px）时主横向间距多缩为（17px），页签间距缩为（3px），隐藏搜索输入。设置宽度不超过（600px）时侧栏由（145px）缩为（44px），仅保留图标可见区域（34px）；设置内容内边距由（24px）缩为（18px）。高度不超过（620px）时隐藏欢迎提示、压缩头部/日期/周条和工具条。

macOS 标题栏留白通过平台类处理：主头部顶部内边距（40px）、最小高度（81px）；浮窗（34px / 62px），折叠浮窗（25px / 62px）。一般头部为（22px 23px 13px）、最小高度（63px）。拖动区在头部，动作区显式取消拖动。

总览月历使用七列与六行；外层（flex: 0 1 58dvh、min-height: 220px），网格行（minmax(58px, 1fr)）且可滚动，下面任务列表也独立滚动。最终覆盖规则取消早期日格最小高度；日期/任务行高均为（16px），溢出提示行高（12px）。不要沿用被覆盖的 99px 日格最小高度。

## Elevation & Depth

以表面色、细边界和状态底色分层；阴影集中在浮出菜单、对话框、消息及轻微焦点反馈，普通任务行没有阴影。窗口表面使用 `color-mix(in srgb, var(--surface) calc(var(--window-alpha) * 100%), transparent)`；默认透明度变量为 1，由窗口状态提供设置值。

### Shadow Vocabulary

- **Menu**：`0 8px 30px rgb(24 45 29 / 0.15)`，菜单浮层。
- **Dialog**：`0 20px 70px rgb(20 35 24 / 0.24)`；背板 `rgb(20 30 23 / 0.28)`。
- **Toast**：`0 5px 18px rgb(20 35 24 / 0.15)`。
- **Entry focus**：`0 2px 9px rgb(32 58 35 / 0.05)`，与绿色录入边框一起出现。
- **Selected segment**：`0 1px 5px rgb(20 35 24 / 0.05)`。
- **Selected calendar day**：`inset 0 0 0 1.5px var(--accent)`，标示当前日期选择。

## Shapes

轮廓以小圆角矩形为主：本地标签（4px）、复选框和小日期按钮（5px）、图标按钮/任务行/页签（6px）、字段与标准按钮（7px）、录入/提示面板（8px）、菜单（9px）、对话框（12px）。任务颜色点与星期选择是圆形；品牌标记（24×24px / 7px），浮窗缩为（21×21px）。边界通常是（1px），任务勾选框为（1.25px），日期选择内描边（1.5px）。

## Components

### Buttons

标准按钮以细边界与小圆角保持密度（8px 12px、最小高度 34px、12px/500）；次要按钮用主表面，悬停用 hover。主按钮使用 accent/surface，悬停 accent-strong。危险按钮在主按钮基础上使用 danger 背景和边框；其文字继承主按钮 surface。图标按钮（29px、内边距 6px、圆角 6px）默认 muted，悬停或活动状态使用 hover 背景与 accent。文本按钮（12px、5px 0）悬停下划线，偏移（3px）。禁用按钮透明度（0.45），光标 not-allowed。

### Inputs / Fields

标准字段使用主表面、细边框和圆角（7px），内边距（9px 10px）；placeholder 为 muted，光标 accent。文本域垂直可缩放，行高（1.65）。键盘可见焦点为（2px solid accent），按钮、字段及链接的 outline-offset 为（3px）；任务树节点有同样焦点描边，其 offset 未另指定。原生复选框（15×15px），range 最小高度（32px）。

底部录入条（8px 圆角、9px 10px 内边距、9px gap）包含无框输入与发送图标，focus-within 改为 accent 边框并出现轻微阴影；内部输入明确不显示自身 outline。浮窗改为（6px 8px 内边距）。快速记录标题字段（16px），容器（12px 24px 23px）。

### Navigation / Chips / Containers

视图页签等宽排列，默认 muted；活动项用 accent-soft 背景、accent 文字和（600）字重，非活动悬停用 secondary。周条日期按钮（8px 圆角），选中项用 accent/surface，并以小点标记今天。设置侧栏当前项用 surface/accent，压缩行为见 Layout。

本地徽标（9px、1px 边框、1px 4px 内边距、4px 圆角）是轻量状态标签；批量操作条使用 accent-soft/accent。通知面板（14px 内边距、8px 圆角）与欢迎提示使用 secondary，无常规卡片阴影。空状态使用小型 Lucide 图标和次级表面圆形，不扩展为插画。

### Task tree

默认行最小高度（43px）、gap（7px）、内边距（6px 3px 6px calc(var(--depth) * 16px + 1px)）；浮窗行（38px）、gap（5px）。组件将视觉缩进层数上限设为 5，逻辑 aria-level 仍表达真实深度。根任务之间用 border 分隔，行悬停用 secondary。

任务勾选框（16×16px / 5px 圆角）；未完成边框混合 muted（65%），完成为 accent 填充和 surface 勾号，部分完成为（7×2px）短横线。完成标题划线（1px）并转为 muted。拖动柄只在悬停可见；更多按钮在悬停、菜单打开或 focus-within 可见。树展开箭头旋转（90deg），过渡（0.18s）。顶部层级路径允许换行；单个路径按钮最大宽（200px），溢出省略。

### Calendar

日期格使用（7px 8px）内边距及右/下细边界；范围填 accent-soft，选中日用 inset accent，今天日期加粗（700）并使用 accent，非本月日期透明度（0.45）。每格任务可省略，并提供额外任务数量提示；已完成任务划线。拖动范围或 Shift+点击选择日期范围，任务可拖入日期。没有额外图表或营销信息。

### Dialogs / Feedback

原生 `<dialog>` 居中；常规宽度（min(540px, calc(100vw - 28px))），最大高度（calc(100dvh - 36px)），外部溢出隐藏。头部（15px 21px）与底部（13px 21px）用细边界；内容（20px 23px）独立滚动，最大高度（calc(100dvh - 195px)）。设置弹窗宽（min(720px, calc(100vw - 24px))），内部高度（min(610px, calc(100dvh - 96px))）、最小高度（260px）；侧栏固定，内容独立滚动。

错误采用 danger 文字、图标与说明；内联错误背景为 danger（7%）的透明混合，错误条贴近底部并保留恢复路径。toast 固定于底部（22px），使用 text 背景和 surface 文字，最大宽（90vw）。菜单通常 z-index（20），底部错误条（21），toast（100）；任务更多菜单使用 fixed 定位以避开列表裁剪。

### Motion and focus

动画仅跟随现有状态：树箭头 transform（0.18s）；加载图标 spin（1s linear infinite）；专注运行圆形 breathe（5s ease-in-out infinite），中点缩放（1.035）并加 `0 8px 28px rgb(72 114 77 / 0.12)` 阴影。没有通用悬停缩放或页面入场动画。用户动画设置关闭或 `prefers-reduced-motion: reduce` 均移除所有 animation/transition。

## Do's and Don'ts

### Do:
- Do 沿用系统字体、Lucide 线性图标和现有语义颜色。
- Do 保持任务标题换行、完成划线、部分完成短横线和键盘焦点。
- Do 保持固定窗口内的独立滚动与底部录入入口。
- Do 同时检查浅色、深色、窄窗口与低高度布局。
- Do 在对话框展示导入预览、错误、批量影响范围等实际业务反馈。

### Don't:
- Don't 加入远程字体、营销页装饰或摄影背景。
- Don't 用大卡片或大留白替换紧凑任务行。
- Don't 把任务标签的多色圆点扩展为整个界面的强调色。
- Don't 用颜色单独表示完成、部分完成、错误或当前选择。
- Don't 把本次代码扫描或浏览器截图视为原生窗口、通知、快捷键的实机验收。
