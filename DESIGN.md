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
  calendar-task-blue: "#2563eb"
  calendar-task-hover: "#1d4ed8"
  calendar-task-completed: "#4064a5"
  app-icon-blue: "#2563eb"
  fixed-white: "#fff"
  window-close-hover: "#c33"
  action-reminder: "#9b6909"
  action-edit: "#3269aa"
  action-date: "#7853a5"
  action-delete: "#bd4545"
  action-delete-armed: "#248046"
  action-reminder-dark: "#e5b957"
  action-edit-dark: "#86b9f1"
  action-date-dark: "#c0a0e8"
  action-delete-dark: "#ef9696"
  action-delete-armed-dark: "#7cce98"
  typing-star-amber: "#d39b27"
  typing-star-blue: "#579fe5"
  typing-star-purple: "#aa83d8"
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
    lineHeight: "23px"
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
    padding: "5px 3px 5px calc(var(--depth) * 22px + 1px)"
  calendar-day:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    padding: "8px"
  calendar-task:
    backgroundColor: "{colors.calendar-task-blue}"
    textColor: "{colors.fixed-white}"
    rounded: "3px"
    padding: "0 5px"
    height: "22px"
  calendar-task-hover:
    backgroundColor: "{colors.calendar-task-hover}"
  calendar-task-completed:
    backgroundColor: "{colors.calendar-task-completed}"
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
  task-row-selected:
    backgroundColor: "{colors.secondary}"
  add-task-button:
    textColor: "{colors.muted}"
    rounded: "{rounded.control}"
  draft-task-row:
    backgroundColor: "{colors.secondary}"
    rounded: "{rounded.control}"
    padding: "6px 6px 6px 21px"
  header-icon-button:
    textColor: "{colors.muted}"
    rounded: "{rounded.control}"
    width: "25px"
    height: "28px"
  task-action:
    rounded: "{rounded.check}"
    padding: "3px"
    size: "23px"
  search-field:
    textColor: "{colors.text}"
    padding: "2px"
  inline-title:
    textColor: "{colors.text}"
    padding: "0 2px"
  task-settings-dialog:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.text}"
    rounded: "{rounded.menu}"
    width: "min(460px, calc(100vw - 28px))"
  list-footer:
    textColor: "{colors.muted}"
    typography: "{typography.label}"
    padding: "0 13px"
    height: "28px"
---

# Design System: SimpleTodo

## Overview

SimpleTodo 的界面把频繁勾选、拆分、改期与快速记录放在同一工作现场。白色与黑色表面、层次灰文字、系统字体和细分隔线保持任务密度；顶栏品牌只保留 SimpleTodo 文字，不显示对钩或本地徽标。应用图标与关于页共用蓝底白色交叉双对钩，图标蓝色不扩展为全局强调色。任务操作、输入星星与日历任务条使用已确认的局部彩色反馈。

最新任务列表交互延续紧凑单窗口：单击标题或行空白直接编辑，光标在末尾；Enter 保存后立即在当前根任务或子任务后开始同级草稿，Shift+Enter 保存并在可见后代末尾打开同一原位子任务草稿。悬停操作在首行紧凑滑出，删除通过两次点击确认，输入光标附近提供短暂星星。普通清单详细设置、改期与更换父任务继续使用主窗口内 HTML 模态对话框；日历具体任务在日期抽屉内嵌入同一详细编辑器。五层任务、拖动预览和日计划连续日期滑块保持现有密度。日历可连续或不连续选日，通过选区右键在同一抽屉创建重复任务。最新行为见 [统一草稿与组件确认](docs/统一草稿与组件确认.md)。应用确认与提示使用同一窗口内的 Vue 对话框；系统文件和日期选择器保留。本轮 macOS 原生 GUI 尚未完成验证，Windows 原生 GUI 仍未实测。

**Key Characteristics:**
- 黑白灰浅深主题配对，系统字体与 Lucide 线性图标；任务操作、输入星星和日历任务条使用局部彩色反馈。
- 唯一原生主窗口；列表、月/周日历、专注、设置和数据工具原位切换。
- 35px 最小任务行、22px 层级缩进与稳定箭头列；单击原位编辑、加号草稿与独立详细设置。
- 搜索填满剩余行宽、固定底部统计、自定义下拉、宽日期抽屉和拖动预览。

## Colors

工作区表面与全局强调仍使用无彩色灰阶；任务操作、输入反馈与日历任务条是已确认的局部彩色例外。浅色是无后缀 token，深色使用对应 `-dark` token；精确值以前置 YAML、`frontend/src/style.css` 与 `frontend/src/composables/typingStars.ts` 为准。

### Primary

`accent` 用于主动作、勾选与关键状态，`accent-strong` 是主动作悬停色，`accent-soft` 是柔和选中面。选中周日期及重复星期按钮使用 `text` 背景与 `surface` 文字，形成浅深主题中的反转。

### Secondary / scoped task feedback

任务操作使用 `action-reminder`（琥珀）、`action-edit`（蓝）、`action-date`（紫）与 `action-delete`（红）；深色使用各自的 `-dark` 配对。首次点击删除后使用 `action-delete-armed` 的绿色对钩，第二次点击才执行删除；绿色仅表示等待再次确认。确认底色沿用代码中的浅色绿色 12% 混合透明面，深色只替换图标前景。光标星星使用 `typing-star-amber`、`typing-star-blue`、`typing-star-purple`，只在输入时短暂出现。这些颜色不形成任务颜色标记入口，也不替换全局灰阶强调。

日历任务条独立使用 `calendar-task-blue`、`calendar-task-hover` 与 `calendar-task-completed`，文字保持 `fixed-white`；完成项同时显示删除线。浅深主题沿用这组日历局部颜色；拖动目标日期复用日历蓝的柔和底色、内框与日期数字反馈，不推广为全局强调色。

### Neutral

`surface` 承载窗口、字段和浮层；`secondary` 承载悬停/选中任务、草稿行和日历滚动轨道；`hover` 用于按钮、菜单与拖动目标反馈。`text`、`muted`、`quiet` 分配给正文、辅助信息与安静标签；浅色后两者同值，深色保留各自配对。`border` 用于细边界与网格。其他业务危险操作沿用灰色 `danger`，由文字、图标、影响范围和确认补充含义；任务行删除使用下述局部动作色。历史颜色字段不形成任务圆点或配置入口。

### Application asset / platform control

`app-icon-blue` 与 `fixed-white` 来自 `build/appicon.svg`：蓝色圆角底面、白色短笔画相交的两个对钩。只用于应用图标及关于页共用图形，不用于顶栏标识或日常工作区。PNG、ICNS、ICO、Icon Composer 资产与托盘模板从这一几何源派生，模板遵循系统单色渲染。

标题栏关闭按钮在 macOS 与 Windows 上均可用，其悬停使用局部窗口操作反馈：`window-close-hover` / `fixed-white`。它不改变黑白灰主题与业务 `danger` 的定义。

## Typography

所有角色使用前置 token 中的系统字体栈，没有远程字体或额外展示字体。品牌 13px/650，窄窗口降至 12px；日计划/收集箱标题 16px/650；日历年月 17px/600；抽屉标题 15px/600。任务默认 13px/23px，按钮 12px，标签与 footer 11px，提醒 10px/16px。详细设置标题字段 16px/24px，专注时钟 62px/300。

未选任务标题使用单行省略与完整 `title` 悬停内容。选中项改为 `white-space: normal`、`overflow-wrap: anywhere`，行高随完整内容扩展。`InlineTitle` 文本域继承当前文字，行高 23px，最小高 24px、最大高 210px，随内容调整高度。提醒在标题下另占小字行，计数与时间使用 tabular-nums；清单完成项保留勾号，标题同时使用删除线与 muted 次要文字色，浅深主题一致；日历完成项沿用蓝底白字与删除线。

## Layout

唯一主窗口中的 `.compact-app` 占满可用宽度和 100dvh，任务列表独立滚动，固定 footer 不随任务滚走。主窗口默认 480×720，最小 400×440；保存过的用户尺寸保留。日历仍在原窗口调整到 1080×720，并限制在可用屏幕区域，返回时恢复列表尺寸，缺省 480×720。截图覆盖 480×720 清单、400×720 Windows 布局和 1080×720 日历抽屉；截图尺寸不另定义平台下限。

头部高 44px，普通左内边距 14px、右内边距 8px；macOS 左右内边距均为 12px，窄窗口也不再预留交通灯空间。Windows 使用无原生标题栏窗口；两平台顶栏操作顺序均为列表、日历、专注、固定、更多、最小化、放大/还原和关闭，关闭始终在最右侧。头部支持窗口拖动，按钮与 summary 显式不拖动。macOS 主窗口将原生最小化、放大和关闭按钮设为隐藏，使用应用自定义窗口控制。

标题区高 43px、周日期高 32px、工具条最小高 34px。搜索占满工具条剩余宽度，允许收缩至零最小宽度，输入无边框、轮廓与阴影；右侧是“选择”，进入选择后出现完成、未完成和删除三个图标，无选中任务时禁用。固定 footer 高 28px，内边距 0 13px，统计靠右，显示完成数量；非空列表全部完成时补充双对钩与“都完成了”，不恢复本地存储标签。

任务列表内边距 4px 10px 14px，根任务之间用细线分隔；行最小 35px，选中长标题、提醒或原位草稿可增加高度。任务业务层级最多五层（根任务为第 1 层），第 5 层不再新增子任务；视觉缩进每级 22px，箭头与叶节点均保留 12px 列位，并对历史超深数据封顶，可访问名称仍说明真实层级。历史超深任务保留且可编辑、完成与缩浅，不能继续加深；新增、重挂、复制与导入由后端检查完整子树。list/listitem/group 语义保留行内表单的原生可访问性。列表末端仅有居中轻量加号，点击后出现草稿，不保留永久输入框。

日计划与收集箱通过标题入口直接切换，不使用标签条；收集箱显示全量任务。专注、设置和数据工具继续在同一原生主窗口内切换。普通清单详细任务设置是主窗口内的 HTML `dialog`；日历复用为抽屉内嵌编辑器，两者均属于当前窗口。

月历七列六行 `minmax(66px, 1fr)`，周历七列一行 `minmax(240px, 1fr)`。日期抽屉覆盖右侧网格，宽 `clamp(360px, 42vw, 480px)`，最大不超过容器宽减 28px；内部显示当天任务列表和草稿，点击具体任务切换为内嵌详细编辑器，选区右键的“创建重复任务”切换为重复任务表单。日期格只显示全部已加载根任务，月/周任务区独立滚动并按溢出显示持久轨道。Esc 从详情返回当天列表，再按 Esc 或外部点击关闭抽屉；离开未保存详情前确认，标题草稿先保存，冲突时保留编辑现场。

520px 以下压缩头部间距与设置内容；顶栏按钮常规宽与最小宽均为 25px、间距 0，现有窄 macOS 样式仍覆盖为 26px。440px 以下分隔间距 2px。没有手机专用导航或营销布局。

## Elevation & Depth

工作面以灰阶表面与细分隔线分层。阴影只用于临时浮出的内容、拖动对象或覆盖工作面：

- 自定义下拉：`0 8px 28px rgb(0 0 0 / .12)`。
- 提醒气泡：`0 8px 28px rgb(0 0 0 / .14)`。
- 日期抽屉：`-12px 0 40px rgb(0 0 0 / .14)`。
- 指针拖动预览：`0 10px 28px #0002`。
- 详细设置：`0 14px 46px #0002`；背板 `#0004`。

日历选中日期内描边 `inset 0 0 0 1.5px var(--text)`。指针拖动的前/后落点使用 2px text 色细线，移入落点使用 hover 背景与 `inset 0 0 0 1px var(--muted)` 轮廓。普通选中行使用 secondary，无额外卡片阴影。

## Shapes

小圆角服务于紧凑工具：勾选框与重复星期 4px，加号/草稿/周日期/下拉触发器 5px，任务行与拖动预览 6px，字段/下拉列表/日历网格 7px，提醒气泡 8px，详细设置 9px。边界通常 1px；任务勾选框 1.25px，选中日内描边 1.5px。日历单格保持直角，避免七列网格成为卡片墙。

## Components

### Buttons / fields / footer

主/次按钮保持 12px/500、8px 12px 内边距、最小高 34px 和 7px 圆角；主按钮使用 accent/surface。一般图标按钮高 27px、宽度声明 27px，继承基础 29px 最小宽；顶栏图标按钮高 28px、常规宽与最小宽 25px，窄 macOS 仍为 26px。任务操作按钮独立使用 23px 方形、3px 内边距、4px 圆角与零间距。禁用透明度 .45。全局去除黑色外焦点轮廓；键盘按钮焦点使用 `inset 0 0 0 1px var(--muted)` 灰色内框，普通按钮同时使用 secondary 底色。主按钮、选中星期和已勾选任务框保留原有反转填色，并使用 `inset 0 0 0 1px var(--surface)` 内框；浅深主题均保留可见焦点。字段聚焦只将细边界改为 muted；搜索和统一原位标题草稿保持透明无框。列表加号的键盘焦点使用 secondary 底色、text 图标与灰色内框。

标准字段使用 surface、border 与 7px 圆角。固定 footer 用 11px 次要文字与顶部细边界表达完成统计，搜索和选择操作保留在列表上方；不把统计与新增输入拼成底部工具条。

### Custom select

统一 `CustomSelect` 使用 combobox 触发器、listbox 浮层与 option。触发器 6px 9px 内边距、最小高 30px、5px 圆角；列表 4px 内边距、最大高 280px、7px 圆角。活动项 hover，选中项勾号。每次展开先将当前选中项滚至可见列表中部，年/月长列表也沿用这一定位。方向键/Home/End 移动，Enter/空格选择，Esc 关闭并返回触发器，Tab 或外部点击关闭。提醒气泡内嵌同一组件，不切回系统 select。

### Task selection / inline title / draft

普通清单单击标题或任务行空白直接进入 `InlineTitle` 文本域，并选中当前任务、完整换行；键盘聚焦任务行后 Enter 在当前任务后打开同级草稿，Shift+Enter 在可见后代末尾打开子任务草稿。聚焦时光标置于文字末尾，不选中原文。所有标题编辑、新建、同级与子任务草稿共用 `InlineTitle`，输入表面透明，无边框、外轮廓、阴影与圆角。Enter 保存成功后在当前根任务或子任务后立即打开同级草稿，失焦只保存，Esc 取消；Shift+Enter 在保存成功后展开父任务，并在其可见后代末尾打开新增子草稿，达到五层限制时保留限制提示。保存尚未结束时按 Esc 会停止后续连续录入。输入法组合期间不提交。标题更新使用进入编辑时的任务树版本，冲突保留草稿、显示错误并回到输入。

完成框高 15px、宽度声明 15px、继承最小宽 16px、4px 圆角；部分完成显示短横线。箭头、完成框、统计和操作固定对齐首行，长标题换行不改变对齐：箭头列宽 12px、叶节点保留相同占位，箭头高 23px、顶部偏移 0，标题、统计和操作区共用 23px 首行，完成框顶部偏移 4px；批量选择框为 16×15px，外边距 9px 4px 0 2px，独立对齐首行而非整个任务树；统计高与行高均为 23px，右边距 2px。行内间距 4px。箭头使用独立 `rotate`（90deg / 180ms），不参与通用按钮图标缩放。

列表末端加号最小高 34px，满可用宽度、图标居中；按下时保留草稿焦点直到点击处理，避免失焦使按钮提前移动。点击后草稿行最小高 35px、6px 6px 6px 21px 内边距、secondary 背景和 5px 圆角；左侧空框只作界面占位。空白失焦或 Esc 不写入数据库，非空 Enter 保存后继续在新任务后开始同级草稿，失焦只保存。根任务和子任务的同级插入均紧跟当前任务，顺序与创建同一事务提交。根任务草稿支持多行粘贴进入导入预览，保留输入法保护。子任务草稿左侧内边距按下一层 22px 缩进加 31px 计算，使用同一 `InlineTitle`；Enter 继续同级、Shift+Enter 继续下一级，失焦保存、空白取消和 Esc 取消与其他草稿一致，不保留专用的常驻子输入。

操作区在鼠标悬停、操作按钮 focus-within 或等待删除确认时显示，其余时间收起。操作区从右侧 9px 滑入，最大宽 96px，宽度/位移过渡 200ms、透明度 160ms；按钮各 23px、零间距。根任务显示提醒，所有任务显示铅笔和删除，日期入口用于根任务改期（重复根隐藏）或子任务更换父任务。铅笔打开详细设置。任务更多菜单、分支聚焦、面包屑和移到收集箱入口已经移除；缩进/提升仍支持 Cmd/Ctrl+[、]。详细设置打开或正在保存时阻止开始拖动；离开标题编辑时先尝试保存，冲突保留草稿。

### Delete confirmation / typing stars

第一次点击红色删除按钮后变为绿色对钩，并保持操作区可见；第二次点击删除当前任务及其完整子树，不再弹出任务删除窗口。Esc、点击其他位置或 4 秒超时取消。可访问名称从“删除任务”变为“确认删除任务”，提示同时说明整个分支与 Esc 取消。

原位标题及根任务、同级和子任务草稿在插入文字后，从多行光标附近发出 5 个小星星；输入法中间组合不产生粒子，compositionend 后呈现一次。全局 canvas 最多保留 80 个粒子，每个 480ms 内淡出、缩小和轻微旋转，四角星大小 2–4.5px。canvas 固定覆盖窗口、忽略指针且 aria-hidden，不抖动窗口、不影响输入。光标测量跟随换行和字段滚动；完成后移除 canvas。关闭应用动画或系统减少动态效果时停止并清理，操作区也停用过渡。

### Detailed settings dialog

铅笔打开 `task-settings-dialog`，通过 HTML `showModal()` 显示。宽 `min(460px, calc(100vw - 28px))`、最大高 `calc(100vh - 32px)`、9px 圆角和 1px 边框；主体内边距 14px，最大高 `calc(100vh - 145px)` 并可滚动，头尾内边距 10px 14px。标题字段 16px/24px、6px 8px 内边距，描述域高 70px。

保留描述、日期、提醒、重复、延期、影响范围确认和版本冲突处理。重复规则明确显示开始日期，结束日期可留空并标注“一直重复”。打开时聚焦标题，关闭后恢复任务焦点；Esc/关闭使用同一未保存检查，Cmd/Ctrl+Enter 保存。不再把完整详情展开在任务行下。

### Plan date / branch parent dialog

“调整计划日期”和分支的“移动到其他父任务”共用详细设置的 HTML 模态表面、头尾与按钮密度，背景列表位置保持不变。主体最大高 `calc(100vh - 170px)`，超出时内部滚动；根任务可选择计划日期及不重复/每日/每周/每月规则，间隔 1—99、每周星期与可选结束日期沿用详细设置的字段和反转星期按钮。计划日期作为重复起点，结束日期留空即一直重复，每周规则从起点当日或之后的首个匹配日开始；切回不重复后忽略已隐藏的旧结束日期。改期与重复规则一起提交。子任务使用自定义下拉选择同一根内的新父任务。Esc、取消或关闭经过未保存检查后恢复任务焦点；保存沿用版本冲突提示，并保留当前移动选择。

### Component confirmation

应用确认与提示统一使用 Vue `ConfirmDialog` 和 HTML `showModal()`，与任务详情共用当前窗口的顶层模态栈，确认层显示在详情之上。表面宽 `min(360px, calc(100vw - 32px))`、10px 圆角、20px 18px 0 内边距，阴影 `0 14px 46px #0003`，背板 `#0004`；标题 15px/600，消息 13px/1.7，长消息在 50vh 内滚动。按钮复用主/次按钮与灰色内焦点框。

打开先聚焦取消，Esc 等同取消，关闭恢复原焦点；每次请求独立，不能复用另一次确认结果，同时到达的新请求默认取消。对话框打开期间按窗口登记忙碌状态，避免贴边隐藏。应用确认不调用浏览器 confirm 或原生 Question；系统文件和日期选择器继续使用系统能力。

### Drag and drop

任务清单仅未完成项可以拖动；完成项隐藏并禁用拖动柄。放下前先保存正在编辑的源标题，冲突保留草稿并中止移动；源任务与前/后/移入位置在异步保存之前捕获。

拖动柄固定 `align-self: flex-start` 与 6px 顶部偏移，避免选中行切换为顶部对齐时把拖动起点移出指针。柄禁用文字选择，内部 SVG 不接收指针事件，手势落在可拖动的 HTML 元素上。

指针从拖动柄移动超过 5px 后开始拖动。预览克隆当前行，保持源行宽高，固定定位跟随指针，透明度 .94、缩放 1.015；隐藏拖动柄及完整任务操作区，原行透明度 .22。指针位于目标上四分之一时排前，下四分之一时排后，中央一半移入为子任务，Alt 强制前置排序；跨根落在子任务边缘时也保留同级插入意图。下缘排序定位到目标的后继同级节点，避免误排到同级末尾。

接近列表上下边缘自动滚动。松手后预览在 180ms 内回落并淡出，列表同级重排沿用 240ms 过渡。Esc、指针取消或窗口失焦撤销拖动，不写入数据库；关闭动画或减少动态效果时直接清理预览。

支持整个分支跨根移动；防环、容量、版本冲突和幂等由事务处理。源根存在说明、提醒或重复信息时确认影响，说明保留到目标根、原提醒取消、重复实例保留跳过记录。反馈区分前/后插入细线与移入灰底轮廓，行背景/透明度过渡 140ms、树箭头独立旋转 180ms。

### Reminder popover

根任务提醒图标打开宽 284px 的气泡，12px 内边距、8px 圆角，最大高 `calc(100vh - 24px)`，内容超出时内部滚动；高度变化后重新定位，保留边界空间。包含时间、不重复/每天/每周/每月的自定义下拉、间隔、星期和可留空截止日期。星期按钮 28×28px、4px 圆角、11px 字，选中黑白反转。

新重复从所选提醒日期开始，星期默认跟随所选日期；现有系列修改后续实例先说明影响数量并确认，停止重复通过详细设置。Esc/外部点击关闭，保存后任务下显示时间小字。气泡内的重复下拉保持键盘与焦点处理。重复状态使用双箭头图标，配有 HTML 悬停提示与可访问名称。界面不显示系统时区标签或时区变更 toast，日期和提醒仍按现有本地时区规则计算。

### Day timeline

日计划周日期条使用 text 色、5px 圆角的连续滑块，选中日期文字反转；滑块由 Motion 以 280ms、`cubic-bezier(.16,1,.3,1)` 平移。日期内容按前进/后退方向从 ±14px 位移和 .55 透明度进入，以 280ms 归位。快速切换从当前可见滑块位置接续，窗口缩放后重新测量落点。共用 Motion 的取消与减少动态效果处理，关闭动画时立即呈现最终日期与内容。

### Calendar / day drawer / update feedback

月/周年月和模式共用自定义下拉；年/月展开定位当前选中项，周视图改年/月也跳转到相应日期范围。日格 8px 内边距，全部已加载根任务使用前置日历蓝色 token 与固定白字，11px/22px、3px 圆角、0 5px 内边距，单行省略；完成项降低蓝色强度并加删除线。今天用加粗和下划线，选日同时用内描边；非本月仅弱化日期数字，保持任务条可读。

`CalendarTaskScroll` 将任务区限制在日期格内；溢出时显示 6px 的 DOM 轨道与至少 12px 的滑块，轨道使用 secondary、滑块使用 muted，悬停改为 text，沿用浅深主题中性色。内容预留 10px，滚轮、轨道点击、滑块拖动和方向键/PageUp/PageDown/Home/End 均可滚动；轨道具有 scrollbar 角色、方向、受控区域及数值 ARIA，无溢出则隐藏轨道，不依赖系统覆盖滚动条。

拖选或 Shift 点选形成连续日期范围，Cmd/Ctrl 点选增减不连续日期；多选关闭日期抽屉，保留网格操作空间。选区使用现有灰阶底色和选日内描边。在选区内右键显示仅含“创建重复任务”的菜单；菜单使用 surface、7px 圆角和局部浮层阴影。顶栏以 11px 次要文字说明选日操作及已选天数，760px 以下隐藏提示文字。

日期抽屉标题 15px，数量 11px，右上角关闭。点击日期打开当天清单；点击日历任务条或抽屉具体任务，在抽屉内显示完整 `TaskEditor`，内容可滚动、底部操作保留，无额外边框、圆角和阴影。保存、取消或 Esc 返回当天清单，外层关闭关闭抽屉。保留提醒、重复规则、离开确认、版本冲突和显式重试；普通清单继续单击原位编辑。

普通根任务可从日历条指针拖动到另一日期，复用任务预览、Esc 取消和事务改期，完整子树继承日期并保留提醒校验。目标日期显示日历蓝 12% 混合 surface 的底色与 2px 内框，日期数字上移 2px（180ms）；指针在网格左右 28px 边缘停留约 650ms 后翻月或翻周，继续停留的翻页间隔至少 950ms。Esc 在下拉聚焦时也取消拖动并清理预览、落点与边缘翻页状态。重复实例禁用拖动并提示通过详情调整规则；拖动完成后的点击不打开详情。

“创建重复任务”沿用同一日期抽屉及详细设置的字段、按钮和底部操作密度，不增加原生窗口。表单默认每周重复，星期由所选日期去重得到，开始日期取最早所选日，结束日期留空表示一直重复；可切换每日/每周/每月、设置 1—99 间隔及可选提醒时间。星期选择保持现有黑白反转，内容区独立滚动。名称为空或保存中禁用创建；校验错误在表单内显示，关闭或离开未保存表单先确认。保存成功返回首次任务日期的当天清单。

结构动效通过 Motion 执行器与 Vue 过渡组件驱动，任务操作显隐使用 CSS，光标星星使用短暂 canvas。抽屉从右侧完整宽度进入（translateX 100% → 0 / 320ms），退出滑至右侧完整宽度并淡出（translateX 0 → 100% / 240ms），使用 `cubic-bezier(.16,1,.3,1)`。任务完成画出对钩并轻缩放（260–280ms），新增和子任务展开采用 200ms 横向揭示，删除和折叠采用 120ms 淡出并立即停止接收交互；同级重排由 CSS 过渡执行 240ms，菜单/下拉和详情采用 200ms 进入、120ms 退出。完成统计与“都完成了”、视图内容切换采用 200ms 反馈。按钮按下时仅内部图标缩放回位（200ms），保留按钮命中区域。关闭动画或 prefers-reduced-motion 会立即完成运行中的动画并停用 CSS 过渡。

标题栏更多位于固定按钮右侧，最右侧统一为关闭。关于页明确任务数据仅存放本地，删除使用提示与技术栈说明；使用同一应用图形并提供手动检查更新，显示检查中、最新/可用版本或错误；固定 GitHub Release 来源，不提供自定义发布地址。更新代理支持默认系统代理、直连、自定义 HTTP 代理（用户名和密码可选），只影响更新检查和安装包下载。自定义字段按需展开，沿用现有字段与两列凭据布局；已保存密码不回显，留空保留，可显式清除。可单独保存代理，检查和下载也会先保存当前表单；请求期间禁用相应操作，下载完成显示保存路径。安装包通过应用客户端经所选代理下载到用户选择的位置，不自动安装或后台联网检查。这些信息是普通设置反馈，不增加常驻更新徽标。

## Do's and Don'ts

### Do:
- Do 沿用前置 token 的浅深配对、系统字体和 Lucide 线性图标。
- Do 让普通清单单击原位编辑、光标置于末尾，Enter 保存后继续同级录入，Shift+Enter 保存后添加子任务；日历具体任务在日期抽屉内打开完整编辑器。
- Do 让月/周日期格内的全部已加载根任务独立滚动，溢出时显示持久轨道；蓝白任务条只用于日历。
- Do 让新建、同级和子任务复用原位草稿，保护输入法组合、版本冲突草稿、空白取消与保存期间 Esc 停止续录。
- Do 使用同窗组件确认、默认取消与焦点返回；键盘焦点使用灰色细内框，保留选中填色。
- Do 使用自定义 combobox/listbox，展开定位当前选中项，保留键盘选择、禁用状态与焦点返回。
- Do 让日历多选保留网格空间，并在同一抽屉创建默认一直重复的任务；日历蓝色落点仅在任务拖动期间出现。
- Do 同时检查 400px 紧凑布局、固定 footer、日历抽屉、深色与减少动画。

### Don't:
- Don't 把应用图标蓝色推广为主题强调色，或恢复品牌对钩、本地徽标与底部存储说明。
- Don't 恢复任务更多菜单、分支聚焦、面包屑或移到收集箱入口；删除使用行内两次确认。
- Don't 新建今日待办、总览、快速记录或任务详情原生窗口。
- Don't 恢复永久新增输入框或专用常驻子输入，或把所有任务强制多行排版；仅选中项完整换行。
- Don't 恢复黑色外焦点框或应用原生确认窗口；系统文件和日期选择器保留。
- Don't 把日期任务再次排在日历下面，或把详细设置恢复为行下展开。
- Don't 把浏览器截图、detector、窗口选项检查或跨平台编译写成原生 GUI 验收通过。
