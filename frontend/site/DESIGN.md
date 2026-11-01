---
name: SimpleTodo 官网
description: 深色聚光中的本地桌面工具，使用真实界面讲清产品。
colors:
  background: "#08090b"
  surface: "#111318"
  text: "#f5f5f7"
  muted: "#a0a3ae"
  quiet: "#8e929f"
  line: "#262930"
  accent: "#77a3ff"
  focus: "#a3c0ff"
  action: "#f4f5f8"
  action-text: "#13151a"
typography:
  display:
    fontFamily: '"Manrope Variable", "PingFang SC", "Microsoft YaHei", sans-serif'
    fontSize: "clamp(50px, 5.3vw, 78px)"
    fontWeight: 750
    lineHeight: 1.2
    letterSpacing: "-0.04em"
  headline:
    fontFamily: '"Manrope Variable", "PingFang SC", "Microsoft YaHei", sans-serif'
    fontSize: "clamp(35px, 4vw, 52px)"
    fontWeight: 650
    lineHeight: 1.3
    letterSpacing: "-0.04em"
  title:
    fontFamily: '"Manrope Variable", "PingFang SC", "Microsoft YaHei", sans-serif'
    fontSize: "24px"
    fontWeight: 550
    letterSpacing: "-0.03em"
  body:
    fontFamily: '"Manrope Variable", "PingFang SC", "Microsoft YaHei", sans-serif'
    fontSize: "14px"
    lineHeight: 1.9
  label:
    fontFamily: '"Manrope Variable", "PingFang SC", "Microsoft YaHei", sans-serif'
    fontSize: "14px"
    fontWeight: 650
rounded:
  control: "8px"
  screenshot: "12px"
  showcase: "16px"
  panel: "20px"
spacing:
  control-gap: "8px"
  icon-gap: "12px"
  panel-gap: "24px"
  section: "110px"
components:
  button-primary:
    backgroundColor: "{colors.action}"
    textColor: "{colors.action-text}"
    typography: "{typography.label}"
    rounded: "{rounded.control}"
    padding: "12px 20px"
  button-primary-hover:
    backgroundColor: "#fff"
  button-secondary:
    backgroundColor: "rgb(17 19 24 / 0.7)"
    textColor: "#d5d7df"
    typography: "{typography.label}"
    rounded: "{rounded.control}"
    padding: "12px 20px"
  button-compact:
    backgroundColor: "{colors.action}"
    textColor: "{colors.action-text}"
    rounded: "{rounded.control}"
    padding: "8px 13px"
  feature-panel:
    backgroundColor: "{colors.surface}"
    rounded: "{rounded.panel}"
---

# Design System: SimpleTodo 官网

## Overview

**Creative North Star: "聚光下的本地桌面"**

以“聚光下的本地桌面”为视觉隐喻：近黑画布、宽松留白与蓝色光束托起真实应用窗口。中文大标题负责传达价值，灰阶正文负责解释，白色下载按钮提供清楚的行动入口。

这是独立于桌面操作界面的官网系统。Raycast 是用户指定的视觉参考；SimpleTodo 的蓝色标志、实际应用截图与本地产品事实构成自身身份。拉丁字符使用随站点打包的 Manrope，中文使用本机字体回退。

**Key Characteristics:**

- 近黑底色与单一蓝色强调。
- 真实截图、柔和阴影与克制的窗口倾斜。
- 大标题、低噪声正文与清楚的白色行动按钮。
- 响应式布局、原生语义与减少动态效果支持。

## Colors

近黑、中性灰与冷蓝构成官网；机器可读值以前言为准，零散插图与键帽色保留在实现中。

### Primary

- **聚光蓝（accent）**：品牌光线、功能选中下划线与少量功能图标。
- **焦点蓝（focus）**：可交互元素的键盘焦点描边。

### Neutral

- **近黑画布（background）**：页面连续底色。
- **石墨面板（surface）**：键盘与独立窗口功能容器。
- **正文白（text）**：主要文字。
- **说明灰（muted）**：正文、次级标题行与导航。
- **注释灰（quiet）**：截图说明及补充文字。
- **分隔线（line）**：页面边界、FAQ 与功能标签栏。
- **行动白／行动墨（action / action-text）**：主要下载按钮的高对比组合。

**The Blue Accent Rule.** 蓝色用于品牌、光线、选中指示与键盘焦点；主要下载按钮保留明亮中性色。

## Typography

**Display Font / Body Font:** Manrope Variable，回退 PingFang SC、Microsoft YaHei、sans-serif。拉丁字形随站点打包；中文由系统字体提供。

**Character:** 厚实但克制的标题与轻一些的中文解释文字共用一个字体栈。大标题使用紧凑字距；正文保留宽松行距。

### Hierarchy

- **Display**：首屏标题使用前言中的 display。平板宽度（≤800px）改为（58px）；手机（≤560px）使用（clamp(38px, 10.8vw, 54px)），行高（1.28）。
- **Headline**：章节标题使用 headline；手机为（34px），行高（1.35）。FAQ 标题桌面（40px）、平板（36px），手机采用章节标题规格。
- **Title**：双功能面板使用 title；手机为（21px）。截图说明标题桌面为（20px / 600 / -0.02em），手机（18px）。
- **Body**：面板及 FAQ 解释为 body。章节引导正文（16px / 1.8），首屏引导（17px / 1.8）；手机正文依区域缩为（12–14px）。截图解释限制宽度（620px）。
- **Label**：主要按钮使用 label；导航为（13px），紧凑下载按钮为（12px）。注释与键帽是小字号辅助信息，不承担主要正文。

## Layout

内容容器最大（1120px），桌面左右各保留（32px）；导航容器最大（1200px），高度（70px）。宽度（≤800px）时左右留白改为（20px），导航高度（64px）。主段落通常上下留白（110px），平板（80px），手机（64px）；各段起止按内容调整，不强制等距。

首屏与主要章节引导居中。解释内容使用双列，功能面板为不等宽双列；在（≤800px）时功能面板与 FAQ 叠为单列，本地数据解释在（≤560px）叠为单列。手机下载区仍保留两个并排的平台选项。

桌面导航固定在顶部。宽度（≤800px）隐藏导航下载按钮，（≤560px）由展开式移动导航接管主要链接。锚点滚动保留顶部空间（92px）。截图可在专属裁剪舞台里重叠，但内容容器和文字不应溢出视口。

## Elevation & Depth

页面用近黑画布、略亮面板、细分隔与柔和环境阴影组织层次。蓝色光束是首屏装饰位图，以 screen 混合和渐隐遮罩进入背景；文字与控制项保留独立前景层。窗口倾斜与透视只服务产品展示。

### Shadow Vocabulary

- **首屏窗口**：深而弥散的阴影（`0 32px 90px rgb(0 0 0 / 0.6)`），叠加一像素半透明浅边。
- **完整截图**：环境阴影（`0 28px 80px rgb(0 0 0 / 0.35)`），叠加石墨色一像素边。
- **小窗口舞台**：较短的阴影（`0 16px 36px #0008`），叠加细边。
- **白色按钮**：轻微顶光（`0 1px 1px rgb(255 255 255 / 0.12)`）。

**The Real Window Rule.** 产品展示使用真实应用截图，并标明演示数据；装饰光线只托起产品，不替代产品证据。

## Shapes

控制项使用克制圆角，功能面板与本机设备示意使用更宽的圆角；完整截图框与浮动窗口分别沿用前言中的 showcase、screenshot 角色。实际完整截图框桌面为（16px），手机为（9px）；浮动小窗口也为（9px）。功能面板手机缩为（16px）。键帽使用较小圆角（5px），保留细边与轻微明度差。

图标来自 Lucide，描边统一为（1.7）；按钮中通常（14–17px），主要功能图标较大。标志继续使用现有 SimpleTodo 图形。

## Components

### Buttons

明亮、直接，下载动作拥有最清楚的前景对比。

- **Primary**：前言中的 button-primary；最小高度（48px），图文间距（12px）。悬停变纯白并上移（2px），按下缩放（0.98）。
- **Secondary**：前言中的 button-secondary，带细描边；悬停略提亮背景与边框。
- **Compact**：导航使用 button-compact，最小高度（36px）；下载区按钮使用（44px）高度。手机首屏按钮为（44px），平台按钮为（42px）。
- **Focus**：链接、按钮与 FAQ 标题使用焦点蓝描边（2px），偏移（5px）。

### Cards / Containers

功能面板是内容舞台：石墨底、柔和圆角，无常驻浮起阴影。键盘面板内边距（34px 28px 26px），手机（28px 23px 22px）；独立窗口面板把解释放在顶部，让截图在下方裁剪区叠放。

### Navigation

导航为灰色文字，悬停变白。移动菜单使用同一深色材料，逐行链接具有宽松点击范围；Escape 关闭并把焦点还给菜单按钮，选择链接或点击菜单外部也关闭。

### Product View Tabs

四个功能标签共用底部分隔线。选中时文字变亮，蓝色细下划线展开（220ms）；键盘支持左右箭头与 Home / End。手机标签的图标排到文字上方。切换后保持截图容器比例（3:2），截图使用 contain，避免截断应用内容。

### Shortcut Selector

按键组合使用真实 kbd 元素。当前选项以较亮石墨底标记，按钮使用 aria-pressed，解释通过礼貌播报区域更新；它展示快捷键说明，不模拟桌面任务操作。

### FAQ

使用原生 details / summary 与底部分隔线。展开箭头旋转（200ms），解释沿正文灰阶展开；首条默认展开。

动效以短距离入场为主：首屏文字（0.8s）、产品舞台（1.1s）、章节进入（0.7s），使用同一个（0.16, 1, 0.3, 1）缓动。交互反馈通常（180ms）。内容默认可见；减少动态效果偏好关闭脚本入场，并缩短 CSS 过渡、关闭平滑滚动。

## Do's and Don'ts

### Do:

- Do 使用这套官网规则扩展营销页面；桌面工作区继续遵循仓库根目录的设计系统。
- Do 保留真实截图比例、清楚的下载入口和可见的键盘焦点。
- Do 让动态效果只增强已可见内容，并响应减少动态效果偏好。
- Do 沿用现有蓝色品牌、随站点打包的字体与本机中文回退。

### Don't:

- Don't 把蓝色扩展成多个竞争的强调色。
- Don't 将营销页面的光效、大字号与宽松留白移植到紧凑的桌面任务界面。
- Don't 用生成的假界面替代实际应用截图。
- Don't 添加虚构客户、用户量、评价或未经核实的功能与授权宣称。
