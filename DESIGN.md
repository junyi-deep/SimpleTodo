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
  app-icon-blue: "#2563eb"
  fixed-white: "#fff"
  window-close-hover: "#c33"
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
  task-row-selected:
    backgroundColor: "{colors.secondary}"
  add-task-button:
    textColor: "{colors.muted}"
    rounded: "{rounded.control}"
  draft-task-row:
    backgroundColor: "{colors.secondary}"
    rounded: "{rounded.control}"
    padding: "6px 6px 6px 21px"
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

SimpleTodo 的界面把频繁勾选、拆分、改期与快速记录放在同一工作现场。白色与黑色表面、层次灰文字、系统字体和细分隔线保持任务密度；顶栏品牌只保留 SimpleTodo 文字，不显示对钩或本地徽标。应用图标与关于页共用蓝底白色交叉双对钩，蓝色不扩展为界面强调色。

最新清单交互在现有紧凑单窗口上扩展：单击选中后长标题完整换行，双击在原位置编辑标题，铅笔打开主窗口内的 HTML 详细设置对话框。本文依据最终 CSS、组件和实际浏览器截图；本轮 macOS 原生补验遇锁屏，Windows 浏览器布局和窗口选项检查也不构成 Windows 原生 GUI 实测。

**Key Characteristics:**
- 黑白灰浅深主题配对，系统字体与 Lucide 线性图标。
- 唯一原生主窗口；列表、月/周日历、专注、设置和数据工具原位切换。
- 35px 最小任务行，选中后换行；标题原位编辑、加号草稿与独立详细设置。
- 左侧搜索、固定底部统计、自定义下拉、宽日期抽屉和拖动预览。

## Colors

工作区主题使用无彩色灰阶。浅色是无后缀 token，深色使用对应 `-dark` token；精确值以前置 YAML 与 `frontend/src/style.css` 为准。

### Primary

`accent` 用于主动作、勾选、可见焦点与关键状态，`accent-strong` 是主动作悬停色，`accent-soft` 是柔和选中面。选中周日期及重复星期按钮使用 `text` 背景与 `surface` 文字，形成浅深主题中的反转。

### Neutral

`surface` 承载窗口、字段和浮层；`secondary` 承载悬停/选中任务、草稿行和日历条目；`hover` 用于按钮、菜单与拖动目标反馈。`text`、`muted`、`quiet` 分配给正文、辅助信息与安静标签；浅色后两者同值，深色保留各自配对。`border` 用于细边界与网格。业务危险操作使用灰色 `danger`，由文字、图标、影响范围和确认补充含义。历史颜色字段不形成任务圆点或配置入口。

### Application asset / platform control

`app-icon-blue` 与 `fixed-white` 来自 `build/appicon.svg`：蓝色圆角底面、白色短笔画相交的两个对钩。只用于应用图标及关于页共用图形，不用于顶栏标识或日常工作区。PNG、ICNS、ICO、Icon Composer 资产与托盘模板从这一几何源派生，模板遵循系统单色渲染。

Windows 关闭按钮悬停是当前 CSS 中明确的局部平台反馈：`window-close-hover` / `fixed-white`。它不改变黑白灰主题与业务 `danger` 的定义。

## Typography

所有角色使用前置 token 中的系统字体栈，没有远程字体或额外展示字体。品牌 13px/650，窄窗口降至 12px；日计划/收集箱标题 16px/650；日历年月 17px/600；抽屉标题 15px/600。任务默认 13px/21px，按钮 12px，标签与 footer 11px，提醒 10px/16px。详细设置标题字段 16px/24px，专注时钟 62px/300。

未选任务标题使用单行省略与完整 `title` 悬停内容。选中项改为 `white-space: normal`、`overflow-wrap: anywhere`，行高随完整内容扩展。`InlineTitle` 文本域继承当前文字，行高 21px，最小高 24px、最大高 210px，随内容调整高度。提醒在标题下另占小字行，计数与时间使用 tabular-nums；完成项保留勾号与划线。

## Layout

唯一主窗口中的 `.compact-app` 占满可用宽度和 100dvh，任务列表独立滚动，固定 footer 不随任务滚走。主窗口默认 480×720，最小 400×440；保存过的用户尺寸保留。日历仍在原窗口调整到 1080×720，并限制在可用屏幕区域，返回时恢复列表尺寸，缺省 480×720。截图覆盖 480×720 清单、400×720 Windows 布局和 1080×720 日历抽屉；截图尺寸不另定义平台下限。

头部高 44px，普通左内边距 14px、右内边距 8px；macOS 左侧为交通灯留 82px，520px 以下留 78px。Windows 使用无原生标题栏窗口，顶栏包含列表、日历、计时、固定、最小化、放大/还原和关闭；头部支持窗口拖动，按钮与 summary 显式不拖动。macOS 保留系统交通灯，不额外显示 Windows 关闭按钮。

标题区高 43px、周日期高 32px、工具条最小高 34px。搜索位于工具条左侧，最小宽 90px、输入填满自身容器；右侧是“选择”，进入选择后出现完成、未完成和删除三个图标，无选中任务时禁用。固定 footer 高 28px，内边距 0 13px，统计靠右，仅显示完成数量，不恢复本地存储标签。

任务列表内边距 4px 10px 14px，根任务之间用细线分隔；行最小 35px，选中长标题、提醒或子输入可增加高度。视觉缩进每级 14px、最多显示五级，可访问名称仍说明真实层级；list/listitem/group 语义保留行内表单的原生可访问性。列表末端仅有居中轻量加号，点击后出现草稿，不保留永久输入框。

日计划与收集箱通过标题入口直接切换，不使用标签条；收集箱显示全量任务。专注、设置和数据工具继续在同一原生主窗口内切换。详细任务设置是主窗口内的 HTML `dialog`，属于当前窗口，不是新建原生窗口。

月历七列六行 `minmax(66px, 1fr)`，周历七列一行 `minmax(240px, 1fr)`。日期抽屉覆盖右侧网格，宽 `clamp(360px, 42vw, 480px)`，最大不超过容器宽减 28px；内部复用任务列表和草稿。Esc 或外部点击关闭；离开未保存详情前确认，标题草稿先保存，冲突时保留编辑现场。

520px 以下压缩头部间距与设置内容；440px 以下头部按钮声明宽 24px、分隔间距 2px。按钮仍继承其现有最小宽度，不能用声明宽度推断所有点击区域已缩至 24px。没有手机专用导航或营销布局。

## Elevation & Depth

工作面以灰阶表面与细分隔线分层。阴影只用于临时浮出的内容、拖动对象或覆盖工作面：

- 任务菜单：`0 8px 30px rgb(0 0 0 / 0.15)`。
- 自定义下拉：`0 8px 28px rgb(0 0 0 / .12)`。
- 提醒气泡：`0 8px 28px rgb(0 0 0 / .14)`。
- 日期抽屉：`-12px 0 40px rgb(0 0 0 / .14)`。
- 拖动预览：`0 8px 24px rgb(0 0 0 / .2)`。
- 详细设置：`0 14px 46px #0002`；背板 `#0004`。

选中日期内描边 `inset 0 0 0 1.5px var(--text)`。拖动目标保留顶部内线 `inset 0 2px 0 var(--text)`，最终背景使用 hover，并显示“中间放置为子任务 · 边缘调整顺序”的 10px 提示。普通选中行使用 secondary，无额外卡片阴影。

## Shapes

小圆角服务于紧凑工具：勾选框与重复星期 4px，加号/草稿/周日期/下拉触发器 5px，任务行与拖动预览 6px，字段/下拉列表/日历网格 7px，提醒气泡 8px，任务菜单与详细设置 9px。边界通常 1px；任务勾选框 1.25px，选中日内描边 1.5px。日历单格保持直角，避免七列网格成为卡片墙。

## Components

### Buttons / fields / footer

主/次按钮保持 12px/500、8px 12px 内边距、最小高 34px 和 7px 圆角；主按钮使用 accent/surface。一般图标按钮高 27px、宽度声明 27px，继承基础 29px 最小宽；顶栏图标按钮高 28px。禁用透明度 .45，键盘焦点为 `2px solid var(--accent)`，普通控件 outline-offset 3px。

标准字段使用 surface、border 与 7px 圆角。固定 footer 用 11px 次要文字与顶部细边界表达完成统计，搜索和选择操作保留在列表上方；不把统计与新增输入拼成底部工具条。

### Custom select

统一 `CustomSelect` 使用 combobox 触发器、listbox 浮层与 option。触发器 6px 9px 内边距、最小高 30px、5px 圆角；列表 4px 内边距、最大高 280px、7px 圆角。活动项 hover，选中项勾号。方向键/Home/End 移动，Enter/空格选择，Esc 关闭并返回触发器，Tab 或外部点击关闭。提醒气泡内嵌同一组件，不切回系统 select。

### Task selection / inline title / draft

单击任务选中，完整标题换行；双击或选中行 Enter 将标题替换成 `InlineTitle` 文本域。Enter 或失焦保存，Esc 取消；输入法组合期间不提交。标题更新使用进入编辑时的任务树版本，冲突保留草稿、显示错误并回到输入。完成框高 15px、宽度声明 15px、继承最小宽 16px、4px 圆角；部分完成显示短横线。

列表末端加号最小高 34px，满可用宽度、图标居中。点击后草稿行最小高 35px、6px 6px 6px 21px 内边距、secondary 背景和 5px 圆角；左侧空框只作界面占位。空白失焦或 Esc 不写入数据库，非空 Enter/失焦保存。新草稿支持多行粘贴进入导入预览，保留输入法保护。Shift+Enter 仍用于添加子任务。

铅笔在悬停、选中或 focus-within 出现，打开详细设置；提醒与更多按钮仍紧邻右侧。详细设置打开或正在保存时阻止开始拖动；离开标题编辑时先尝试保存，冲突保留草稿。

### Detailed settings dialog

铅笔打开 `task-settings-dialog`，通过 HTML `showModal()` 显示。宽 `min(460px, calc(100vw - 28px))`、最大高 `calc(100vh - 32px)`、9px 圆角和 1px 边框；主体内边距 14px，最大高 `calc(100vh - 145px)` 并可滚动，头尾内边距 10px 14px。标题字段 16px/24px、6px 8px 内边距，描述域高 70px。

保留描述、日期、提醒、重复、延期、影响范围确认和版本冲突处理。打开时聚焦标题，关闭后恢复任务焦点；Esc/关闭使用同一未保存检查，Cmd/Ctrl+Enter 保存。不再把完整详情展开在任务行下。

### Drag and drop

预览由当前行克隆，最大宽 420px，加阴影并移除拖动柄/更多/提醒等操作控件；原行透明度 .35。指针位于目标上四分之一时排前，下四分之一时排后，中央为子任务，Alt 强制前置排序。下缘排序定位到目标的后继同级节点，避免误排到同级末尾。

支持整个分支跨根移动；防环、容量、版本冲突和幂等由事务处理。源根存在说明、提醒或重复信息时确认影响，说明保留到目标根、原提醒取消、重复实例保留跳过记录。反馈包含目标灰底、内线和文字提示，行过渡 140ms、树箭头 180ms。

### Reminder popover

根任务提醒图标打开宽 284px 的气泡，12px 内边距、8px 圆角，最大高 `calc(100vh - 24px)`，内容超出时内部滚动；高度变化后重新定位，保留边界空间。包含时间、不重复/每天/每周/每月的自定义下拉、间隔、星期和可留空截止日期。星期按钮 28×28px、4px 圆角、11px 字，选中黑白反转。

新重复从所选提醒日期开始，星期默认跟随所选日期；现有系列修改后续实例先说明影响数量并确认，停止重复通过详细设置。Esc/外部点击关闭，保存后任务下显示时间小字。气泡内的重复下拉保持键盘与焦点处理。

### Calendar / day drawer / update feedback

月/周年月和模式共用自定义下拉。日格 8px 内边距，任务条 11px/20px、灰底、单行省略；今天用加粗和下划线，选日同时用内描边，非本月透明度 .48。日期抽屉标题 15px，数量 11px，右上角关闭；列表复用最新选中/原位编辑/草稿行为。

抽屉进入/退出 200ms、横移 22px 与透明度，位移缓动 `cubic-bezier(.16,1,.3,1)`；关闭动画或 prefers-reduced-motion 会移除所有动画与过渡。

关于页使用同一应用图形并提供手动检查更新，显示检查中、最新/可用版本或错误；固定 GitHub Release 来源，不提供自定义发布地址。平台压缩包下载指向 Release 附件，不自动安装或后台联网检查。这些信息是普通设置反馈，不增加常驻更新徽标。

## Do's and Don'ts

### Do:
- Do 沿用前置 token 的浅深配对、系统字体和 Lucide 线性图标。
- Do 区分单击选中、双击原位编辑与铅笔详细设置，保留键盘焦点和完整标题。
- Do 让加号先创建界面草稿，保护输入法组合、版本冲突草稿和空白取消。
- Do 使用自定义 combobox/listbox，保留键盘选择、禁用状态与焦点返回。
- Do 同时检查 400px 紧凑布局、固定 footer、日历抽屉、深色与减少动画。

### Don't:
- Don't 把应用图标蓝色推广为主题强调色，或恢复品牌对钩、本地徽标与底部存储说明。
- Don't 新建今日待办、总览、快速记录或任务详情原生窗口。
- Don't 恢复永久新增输入框，或把所有任务强制多行排版；仅选中项完整换行。
- Don't 把日期任务再次排在日历下面，或把详细设置恢复为行下展开。
- Don't 把浏览器截图、detector、窗口选项检查或跨平台编译写成原生 GUI 验收通过。
