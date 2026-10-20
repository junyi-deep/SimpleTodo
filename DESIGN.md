---
name: SimpleTodo
description: "黑白灰、紧凑、单窗口的本地桌面待办界面"
colors:
  surface: "#ffffff"
  secondary: "#f5f5f5"
  hover: "#ededed"
  text: "#202020"
  muted: "#666666"
  quiet: "#666666"
  border: "#e5e5e5"
  accent: "#262626"
  accent-strong: "#111111"
  accent-soft: "#eeeeee"
  danger: "#555555"
  surface-dark: "#191919"
  secondary-dark: "#222222"
  hover-dark: "#2c2c2c"
  text-dark: "#eeeeee"
  muted-dark: "#b0b0b0"
  quiet-dark: "#a9a9a9"
  border-dark: "#383838"
  accent-dark: "#e4e4e4"
  accent-strong-dark: "#ffffff"
  accent-soft-dark: "#333333"
  danger-dark: "#c9c9c9"
typography:
  brand:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "13px"
    fontWeight: 650
    letterSpacing: "-0.02em"
  workspace-title:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "16px"
    fontWeight: 650
    letterSpacing: "-0.02em"
  calendar-title:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "17px"
    fontWeight: 600
  body:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "14px"
  task:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "13px"
    lineHeight: "21px"
  button:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "12px"
    fontWeight: 500
  label:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "11px"
  reminder:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "10px"
    lineHeight: "16px"
  drawer-title:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "15px"
    fontWeight: 600
  clock:
    fontFamily: "-apple-system, BlinkMacSystemFont, \"Segoe UI\", \"PingFang SC\", \"Microsoft YaHei\", sans-serif"
    fontSize: "62px"
    fontWeight: 300
    lineHeight: 1.4
    letterSpacing: "-0.04em"
rounded:
  check: "4px"
  control: "5px"
  row: "6px"
  field: "7px"
  popover: "8px"
  menu: "9px"
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
  "14": "14px"
  "16": "16px"
  "18": "18px"
  "28": "28px"
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
  icon-button:
    textColor: "{colors.muted}"
    rounded: "{rounded.control}"
    padding: "5px"
    size: "27px"
  field:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    rounded: "{rounded.field}"
    padding: "9px 10px"
  select-trigger:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    rounded: "{rounded.control}"
    padding: "6px 9px"
  task-row:
    rounded: "{rounded.row}"
    padding: "5px 3px 5px calc(var(--depth) * 14px + 1px)"
  inline-add:
    textColor: "{colors.muted}"
    rounded: "{rounded.check}"
    padding: "3px 7px 3px 28px"
  inline-editor:
    backgroundColor: "{colors.secondary}"
    textColor: "{colors.text}"
    rounded: "{rounded.field}"
  calendar-day:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    padding: "8px"
  day-drawer:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    width: "clamp(360px, 42vw, 480px)"
  reminder-popover:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    rounded: "{rounded.popover}"
    padding: "12px"
    width: "284px"
---

# Design System: SimpleTodo

## Overview

SimpleTodo 的界面把频繁勾选、拆分、改期和连续录入放在同一工作现场。白色与黑色表面、层次灰文字、系统字体和细分隔线保持任务密度；品牌只保留 SimpleTodo 文字，不增加状态徽标或装饰图案。

这是用户明确指定的替代视觉，取代原绿色、多窗口与底部录入设计。本文以当前 CSS 覆盖规则、Vue 组件及实际浏览器截图为依据；设计记录与浏览器评审不构成 macOS/Windows 原生窗口、托盘、通知或快捷键的实机验收结论。

**Key Characteristics:**
- 黑白灰浅深主题配对，系统字体与 Lucide 线性图标。
- 唯一原生主窗口；列表、月/周日历、专注、设置和数据工具原位切换。
- 35px 最小任务行，单行标题；新增、详情和提醒均紧邻任务。
- 自定义下拉、宽右侧日期抽屉、任务拖动预览与清楚落点。

## Colors

所有界面色均为无彩色灰阶。浅色是无后缀 token，深色使用对应 `-dark` token；精确值以前置 YAML 和 `frontend/src/style.css` 为准。

### Primary

`accent` 用于主动作、勾选、可见焦点与关键状态，`accent-strong` 是主动作悬停色，`accent-soft` 是柔和选中面。选中周日期直接使用 `text` 背景和 `surface` 文字，形成浅深主题中的反转。

### Neutral

`surface` 承载窗口、字段和浮层；`secondary` 承载悬停任务、行内编辑和日历条目；`hover` 用于按钮与菜单的活动反馈。`text`、`muted`、`quiet` 分别承担正文、辅助信息和安静标签；浅色后两者同值，深色保留各自配对。`border` 用于一像素边界与网格。`danger` 也是灰色，危险操作通过文字、图标、影响范围和确认共同表达，不重新引入红色。

历史数据中的颜色字段不再形成可见圆点或编辑入口。不要由保留的旧 CSS 选择器推断仍有多色任务体系。

## Typography

所有角色使用前置 token 中的系统字体栈，没有远程字体或额外展示字体。品牌 13px/650，窄窗口降至 12px；日计划/收集箱切换标题 16px/650；日历年月 17px/600；抽屉标题 15px/600。任务 13px/21px，按钮与行内字段 12px，辅助标签 11px，提醒 10px/16px。专注时钟仍为 62px/300，仅用于计时页。

任务标题使用 `white-space: nowrap`、`overflow: hidden` 和 `text-overflow: ellipsis`，`title` 提供完整悬停内容。提醒时间在标题下另占一行，日期、计数和时间使用 tabular-nums。完成项通过勾号与划线表达。

## Layout

唯一主窗口中的 `.compact-app` 占满可用宽度和 100dvh，高度溢出由内部区域滚动。列表不再在大窗口中居中限宽。头部固定 44px，普通内边距 8px 14px；macOS 头部左侧为原生窗口控制留 82px，520px 以下留 78px。右侧依次提供视图切换与固定、最小化、放大/还原等控制，不显示品牌图标或本地徽标。

原生主窗口默认 600×580、最小 440×380；日历切换调用当前窗口调整到 1080×720，返回时恢复保存的列表尺寸，缺省 600×580。评审截图覆盖列表 600×580、440×480 和日历 1080×720；这些截图尺寸不另定义平台限制。原生实机验收另行记录。

列表标题区高 43px，周日期高 32px，工具条最小高 34px。任务列表内边距 4px 10px 14px，根任务之间保留细分隔线；每行最小 35px，提醒、子输入和展开编辑会增加内容高度。视觉缩进每级 14px、最多显示五级缩进，ARIA 层级仍表示真实深度。新增行直接接在列表末端，最小高 36px，不占用固定页脚。

日计划与收集箱通过同一个标题入口直接切换，没有独立标签条或所有未完成视图。收集箱显示全量任务，日计划按日期筛选。专注、设置和数据工具是同一主窗口的嵌入工作面，不新开原生窗口。

月历为七列六行 `minmax(66px, 1fr)`，周历为七列一行 `minmax(240px, 1fr)`，网格内部可滚动。选日期后右侧抽屉覆盖网格一部分，宽 `clamp(360px, 42vw, 480px)`，最大不超过容器宽减 28px；日期任务不在网格下面重复排列。抽屉含独立任务滚动区，Esc 或外部点击关闭，未保存编辑先确认。

520px 以下缩小头部间距、搜索宽至 80px、行内编辑横向间距和设置内容边距；设置侧栏仅保留紧凑图标区域。没有手机专用导航或营销布局。

## Elevation & Depth

默认工作面平坦，以灰阶表面和细分隔线组织信息。阴影只帮助辨认暂时浮出的内容、拖动对象或覆盖工作面：

- 任务菜单：`0 8px 30px rgb(0 0 0 / 0.15)`。
- 自定义下拉：`0 8px 28px rgb(0 0 0 / .12)`。
- 提醒气泡：`0 8px 28px rgb(0 0 0 / .14)`。
- 日期抽屉：`-12px 0 40px rgb(0 0 0 / .14)`。
- 拖动预览：`0 8px 24px rgb(0 0 0 / .2)`。

选中日期使用 `inset 0 0 0 1.5px var(--text)`，落点使用 `inset 0 2px 0 var(--text)`。嵌入设置/数据工作面明确取消外层阴影和圆角；行内详情仅有边界与次级表面。

## Shapes

小圆角服务于紧凑工具：勾选框与新增行 4px，图标按钮、周日期和下拉触发器 5px，任务行与拖动预览 6px，字段、下拉列表、行内详情和日历网格 7px，提醒气泡 8px，任务菜单 9px。边界通常 1px；任务勾选框 1.25px，选中日期内描边 1.5px。日历单格保持直角，避免七列网格成为卡片墙。

## Components

### Buttons / fields

标准主/次按钮保持 12px/500、8px 12px 内边距、最小高 34px 和 7px 圆角；主按钮用 accent/surface 配对。图标按钮高 27px、宽度声明 27px，保留基础样式的 29px 最小宽度；内边距 5px，更多按钮更窄。禁用透明度 .45。键盘可见焦点为 `2px solid var(--accent)`，普通控件 outline-offset 3px。

标准字段为 surface、border 和 7px 圆角；行内新增字段无框，容器 focus-within 使用 secondary 表面，避免重复焦点框。新增输入支持连续 Enter，保留输入法组合保护。

### Custom select

统一 `CustomSelect` 使用按钮 `role=combobox` 和弹出 `role=listbox`，选项 `role=option`。触发器 6px 9px 内边距、最小高 30px、5px 圆角；列表 4px 内边距、最大高 280px、7px 圆角。弹出定位跟随触发器并避让窗口边界，最小目标宽 190px、视口左右保留 10px。活动选项用 hover，已选项有勾号。方向键/Home/End 移动，Enter/空格选择，Esc 关闭并返回触发器；Tab 或外部点击关闭。不使用系统 select 弹出菜单。

### Task tree / inline editor

任务行最小 35px、gap 5px；完成框高 15px、宽度声明 15px，保留基础样式的 16px 最小宽度，圆角 4px，部分完成为短横线。根任务分隔线清楚，标题保持单行省略。更多与提醒在悬停或 focus-within 出现；根任务提醒时间在标题下面以 10px 小字显示。

Enter 打开详情，Shift+Enter 添加子任务，详情在对应任务行下展开；编辑容器为 secondary 表面、7px 圆角、1px 边框，头部与底部各紧凑排列。更多菜单不重复提供编辑、添加子任务、复制、上移和下移。切换日期、工作面或抽屉前先处理未保存草稿；拖动在编辑期间停用。

拖动使用当前行克隆构造最大宽 420px 的预览，移除操作控件并加阴影；原行透明度 .35。落点以顶部内线和 secondary 背景表达，避免仅依赖光标猜测目的地。行背景/透明度过渡 140ms，树展开箭头 180ms。

### Reminder popover

提醒图标位于更多操作左边；根任务点击后显示宽 284px 的原位气泡，12px 内边距、8px 圆角。气泡包含时间字段、移除与保存动作，输入立即获得焦点；Esc 与外部点击关闭。气泡不新建原生窗口。已设置的提醒持续显示为任务下方时间小字。

### Calendar / day drawer

月/周共用工作区；年月与模式均用自定义下拉。日格 8px 内边距，任务条 11px/20px、灰底、单行省略；完成条划线。今天用加粗和下划线，当前选日同时使用内描边，非本月格透明度 .48。

日期抽屉有 16px 侧边距、15px 标题和 11px 数量说明，右上角关闭按钮；内部复用完整任务列表、行内录入与编辑。进入/退出为 200ms 横移 22px 加透明度，位移缓动 `cubic-bezier(.16,1,.3,1)`。减少动画或关闭动画设置会移除所有动画与过渡。

## Do's and Don'ts

### Do:
- Do 沿用前置 token 的浅深配对、系统字体和 Lucide 线性图标。
- Do 保持任务单行省略、完整悬停标题、明确完成勾号和键盘焦点。
- Do 将新增行与详情保留在任务列表现场，切换视图前保护未保存草稿。
- Do 使用自定义 combobox/listbox，保留键盘选择和焦点返回。
- Do 同时检查紧凑窗口、日历抽屉、深色及减少动画设置。

### Don't:
- Don't 恢复绿色主题、任务颜色入口、品牌对钩、本地徽标或底部本地存储说明。
- Don't 新建今日待办、总览或快速记录原生窗口。
- Don't 用大卡片、大留白或多行任务标题降低列表密度。
- Don't 把日历日期任务再次排在网格下方；使用右侧抽屉。
- Don't 把浏览器截图、detector 或代码检查写成原生系统能力通过。
