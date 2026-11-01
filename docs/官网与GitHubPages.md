# 官网与 GitHub Pages

公开地址：<https://junyi-deep.github.io/SimpleTodo/>

官网独立放在 `frontend/site/`，介绍核心功能、本地数据优势和平台下载。整体参考 Raycast 的深色视觉、光影与产品展示方式，使用 SimpleTodo 自己的标识、文案和实际界面。

## 本地维护

在仓库根目录执行：

```sh
npm ci --prefix frontend
npm run dev:site --prefix frontend
```

访问 <http://127.0.0.1:9255/SimpleTodo/>。生产构建与验证：

```sh
npm run build:site --prefix frontend
npm run test:site --prefix frontend
npm run preview:site --prefix frontend
```

网站构建输出为 `frontend/site-dist/`，不会打包进桌面应用。站点使用 `/SimpleTodo/` 基路径；若更换仓库名或域名，同时调整 Vite 配置、规范链接、分享信息、robots 和 sitemap。

## 发布

`.github/workflows/pages.yml` 在 `main` 分支的站点文件、相关配置或依赖变化时自动运行，也可以手动触发。流程先构建并检查桌面／手机交互，再上传静态产物并部署到 GitHub Pages。仓库 Pages 来源使用 GitHub Actions。

网站更新不需要桌面版本标签。两个下载入口始终指向 GitHub 最新 Release 的 macOS ARM64 和 Windows AMD64 压缩包；桌面应用的版本发布流程独立维护。

## 图片与字体

- 产品图片全部来自实际应用，任务和笔记为隔离数据库中的演示数据。重新生成时执行 `CAPTURE_SITE_ASSETS=1 npm run test:e2e --prefix frontend -- site-assets.spec.ts`，不要使用个人数据库截图。
- 蓝色光影是生成的装饰背景，不是生成的应用界面。完整提示词与图片来源记录在 `frontend/site/assets-origin.json`。
- 图片内嵌或附带来源信息。重新生成后应更新对应的来源元数据。
- 标识来自现有 `build/appicon.svg`；Manrope 字体自行托管，许可保存在 `frontend/site/public/licenses/Manrope-OFL.txt`。

网站不包含分析追踪，也不在浏览器中请求用户的本机任务数据。功能页切换、FAQ 和移动导航均在前端完成；下载按钮跳转 GitHub Release。
