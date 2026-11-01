import { test, expect } from "@playwright/test";
import { call, root } from "./helpers";

test("生成官网真实应用截图（隔离演示数据）", async ({ browser, request }) => {
  test.skip(process.env.CAPTURE_SITE_ASSETS !== "1", "仅在更新官网截图时运行");
  test.setTimeout(90000);
  const preview = await call(
    request,
    "PreviewImport",
    JSON.stringify({ formatVersion: 1, tasks: [] }),
    "json",
    null,
    true,
  );
  await call(request, "CommitImport", preview.token);
  await call(request, "Settings", {
    theme: "dark",
    animation: "false",
    expandDepth: "3",
    autoPostpone: "false",
    completedBottom: "true",
    showTags: "true",
    dayShowCompleted: "true",
    dayFilterTags: "[]",
    inboxFilterTags: "[]",
    notificationPrompted: "true",
  });
  await call(request, "SaveTags", [
    { id: "important", name: "重要", order: 0 },
    { id: "urgent", name: "紧急", order: 1 },
  ]);
  const plan = await root(request, "让这一周，从容一点");
  for (const title of [
    "梳理最重要的三件事",
    "拆成可以开始的小步骤",
    "为临时的灵感留个位置",
  ]) {
    const latest = await call(request, "FindTask", plan.id);
    await call(request, "Create", {
      commandId: crypto.randomUUID(),
      parentId: plan.id,
      rootId: plan.id,
      expectedVersion: latest.version,
      title,
    });
  }
  const focus = await root(request, "留出 25 分钟，认真做一件事");
  await call(request, "Update", {
    commandId: crypto.randomUUID(),
    id: focus.id,
    rootId: focus.id,
    expectedVersion: focus.version,
    title: focus.title,
    date: focus.date,
    color: "none",
    postpone: "off",
    remindAt: Date.now() + 3600000,
  });
  const latest = await call(request, "FindTask", focus.id);
  await call(
    request,
    "SetTaskTags",
    {
      commandId: crypto.randomUUID(),
      id: focus.id,
      rootId: focus.id,
      expectedVersion: latest.version,
    },
    ["important"],
  );
  await root(request, "整理会议记录与下一步行动");
  for (const title of ["读完那篇收藏很久的文章", "完成今天的散步"]) {
    const task = await root(request, title);
    await call(request, "Complete", {
      commandId: crypto.randomUUID(),
      id: task.id,
      rootId: task.id,
      expectedVersion: task.version,
      completed: true,
    });
  }
  await call(request, "SaveNote", {
    id: "",
    title: "这一周，值得记下的事",
    content:
      "# 本周的小进展\n\n把重要的事情放在前面，也为生活留一点空白。\n\n## 做得不错\n- 认真完成了设计评审\n- 每天留出一段专注时间\n\n## 下一步\n把想做的事拆成小步骤，一件一件完成。",
    version: 0,
  });
  const context = await browser.newContext({
    viewport: { width: 480, height: 720 },
    deviceScaleFactor: 2,
  });
  const page = await context.newPage();
  const capture = async (name: string) =>
    page.screenshot({
      path: `site/public/images/${name}.png`,
      animations: "disabled",
    });
  try {
    await page.goto("http://127.0.0.1:9245/?preview=1");
    await expect(
      page.getByRole("button", { name: plan.title, exact: true }),
    ).toBeVisible();
    await capture("list-portrait");
    await page.setViewportSize({ width: 960, height: 640 });
    // Start a fresh render after resizing, avoiding a stale compositor tile.
    await page.reload();
    await expect(
      page.getByRole("button", { name: plan.title, exact: true }),
    ).toBeVisible();
    await capture("list-dark");
    await page.setViewportSize({ width: 1080, height: 720 });
    await page.goto("http://127.0.0.1:9245/?preview=1&window=calendar");
    await expect(
      page.getByRole("button", { name: focus.date, exact: true }),
    ).toBeVisible();
    await page.getByRole("button", { name: focus.date, exact: true }).click();
    await capture("calendar-dark");
    await page.goto("http://127.0.0.1:9245/?preview=1&window=review");
    await expect(
      page.getByRole("heading", { name: "完成热力图", exact: true }),
    ).toBeVisible();
    await capture("review-dark");
    await page.getByRole("button", { name: "复盘", exact: true }).click();
    await page
      .getByRole("button", { name: /这一周，值得记下的事 创建/ })
      .click();
    await page
      .getByRole("button", { name: "切换笔记预览", exact: true })
      .click();
    await capture("notes-dark");
    await page.setViewportSize({ width: 800, height: 600 });
    await page.goto("http://127.0.0.1:9245/?preview=1&window=focus");
    await expect(page.getByRole("timer")).toBeVisible();
    await capture("focus-dark");
  } finally {
    await context.close();
  }
});
