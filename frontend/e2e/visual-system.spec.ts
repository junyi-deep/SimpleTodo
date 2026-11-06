import { test, expect } from "@playwright/test";
import { call, root, row, select } from "./helpers";

test("视觉系统：双主题、控件状态和所有工作视图", async ({ page, request }) => {
  test.setTimeout(90000);
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
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
    theme: "light",
    animation: "false",
    expandDepth: "3",
    autoPostpone: "false",
    completedBottom: "true",
    showTags: "true",
    dayFilterTags: "[]",
    inboxFilterTags: "[]",
    dayShowCompleted: "true",
    inboxShowCompleted: "false",
    notificationPrompted: "true",
  });
  await call(request, "SaveTags", [
    { id: "important", name: "重要", order: 0 },
    { id: "urgent", name: "紧急", order: 1 },
  ]);
  const boot = await call(request, "Boot");
  const plan = await root(request, "整理本周工作计划");
  for (const title of ["梳理本周优先级", "确认里程碑与交付时间"]) {
    const latest = await call(request, "FindTask", plan.id);
    await call(request, "Create", {
      commandId: crypto.randomUUID(),
      rootId: plan.id,
      parentId: plan.id,
      expectedVersion: latest.version,
      title,
    });
  }
  const version = await call(request, "FindTask", plan.id);
  await call(
    request,
    "SetTaskTags",
    {
      id: plan.id,
      rootId: plan.id,
      expectedVersion: version.version,
      commandId: crypto.randomUUID(),
    },
    ["important"],
  );
  const document = await root(request, "整理项目文档与会议记录");
  await call(request, "Update", {
    commandId: crypto.randomUUID(),
    id: document.id,
    rootId: document.id,
    expectedVersion: document.version,
    title: document.title,
    date: document.date,
    description:
      "把分散的记录整理成清晰的下一步。\n\n检查会议结论，补齐负责人和交付时间。",
    postpone: "off",
    color: "none",
    remindAt: Date.now() + 3600000,
  });
  await root(request, "给自己留一段专注时间");
  await root(request, "回顾本周进展，记下下一步要做的事情");
  for (const title of ["完成今天的阅读笔记", "清理收集箱"]) {
    const t = await root(request, title);
    await call(request, "Complete", {
      id: t.id,
      rootId: t.id,
      expectedVersion: t.version,
      commandId: crypto.randomUUID(),
      completed: true,
    });
  }
  const next = new Date(boot.today + "T12:00:00");
  next.setDate(next.getDate() + 2);
  const nextDay = [
    next.getFullYear(),
    String(next.getMonth() + 1).padStart(2, "0"),
    String(next.getDate()).padStart(2, "0"),
  ].join("-");
  await root(request, "预约下周的设计评审", nextDay);
  await call(request, "SaveNote", {
    id: "",
    title: "本周工作回顾",
    content:
      "# 本周小结\n\n## 做得不错\n- 把重要的事情安排在前面\n- 为深度工作留出时间\n\n## 下一步\n把计划拆成可以完成的小步骤。",
    version: 0,
  });
  await page.setViewportSize({ width: 480, height: 720 });
  await page.goto("/?preview=1");
  await expect(row(page, plan.title)).toBeVisible();
  const loadedFonts = await page.evaluate(async () => {
    const weights = [400, 450, 500, 550, 600, 650, 700];
    const loaded = await Promise.all(
      weights.map((weight) =>
        document.fonts.load(
          `${weight} 10px "Source Han Sans SC"`,
          "任务提醒完成 123 ABC",
        ),
      ),
    );
    await document.fonts.ready;
    return loaded.map((faces) =>
      faces.map((face) => ({ family: face.family, status: face.status })),
    );
  });
  for (const faces of loadedFonts) {
    expect(faces).toHaveLength(1);
    expect(faces[0].status).toBe("loaded");
  }
  expect(
    await row(page, plan.title).evaluate(
      (el) => getComputedStyle(el).fontFamily,
    ),
  ).toMatch(/^"Source Han Sans SC"/);
  const edit = row(page, document.title).getByLabel("任务设置", {
    exact: true,
  });
  await row(page, document.title).hover();
  const neutral = await edit.evaluate((el) => getComputedStyle(el).color);
  const date = row(page, document.title).getByLabel("调整任务时间", {
    exact: true,
  });
  expect(await date.evaluate((el) => getComputedStyle(el).color)).toBe(neutral);
  await edit.hover();
  expect(await edit.evaluate((el) => getComputedStyle(el).color)).not.toBe(
    neutral,
  );
  await page.getByLabel("搜索当前视图").hover();
  await page.screenshot({
    path: "../.impeccable/review/refined-list-light.png",
    animations: "disabled",
  });
  await page.getByLabel("过滤显示", { exact: true }).focus();
  await page.screenshot({
    path: "../.impeccable/review/refined-keyboard-focus.png",
    animations: "disabled",
  });
  expect(
    await page.locator("body").evaluate((el) => el.scrollWidth <= innerWidth),
  ).toBe(true);
  await call(request, "Settings", { theme: "dark" });
  await page.reload();
  await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
  await page.screenshot({
    path: "../.impeccable/review/refined-list-dark.png",
    animations: "disabled",
  });
  await page.setViewportSize({ width: 400, height: 720 });
  await row(page, document.title).hover();
  await row(page, document.title)
    .getByLabel("任务设置", { exact: true })
    .click();
  const editor = page.getByRole("dialog", { name: "任务详情", exact: true });
  await expect(editor).toBeVisible();
  expect(await editor.evaluate((el) => el.scrollWidth <= el.clientWidth)).toBe(
    true,
  );
  await editor.getByLabel("描述", { exact: true }).focus();
  const title = editor.getByLabel("任务标题", { exact: true });
  await expect(title).toHaveCSS("border-top-width", "0px");
  await expect(title).toHaveCSS("padding-left", "0px");
  await expect(title).toHaveCSS("background-color", "rgba(0, 0, 0, 0)");
  await page.screenshot({
    path: "../.impeccable/review/refined-editor-dark.png",
    animations: "disabled",
  });
  await editor.getByRole("button", { name: "取消", exact: true }).click();
  await call(request, "Settings", { theme: "light" });
  await page.setViewportSize({ width: 1080, height: 760 });
  await page.reload();
  await page.getByLabel("日历视图", { exact: true }).click();
  await page.getByRole("button", { name: boot.today, exact: true }).click();
  await expect(page.getByLabel("日期任务抽屉", { exact: true })).toBeVisible();
  await page.screenshot({
    path: "../.impeccable/review/refined-calendar-light.png",
    animations: "disabled",
  });
  await page.getByLabel("关闭日期任务", { exact: true }).click();
  await select(page, "日历范围", "甘特图");
  await expect(
    page.getByRole("region", { name: "甘特图", exact: true }),
  ).toBeVisible();
  await page.screenshot({
    path: "../.impeccable/review/refined-gantt-light.png",
    animations: "disabled",
  });
  await page.goto("/?preview=1&window=settings");
  await page.setViewportSize({ width: 760, height: 720 });
  await page.getByRole("button", { name: "外观与窗口", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "轻一点，清楚一点", exact: true }),
  ).toBeVisible();
  await page.screenshot({
    path: "../.impeccable/review/refined-settings-light.png",
    animations: "disabled",
  });
  await call(request, "Settings", { theme: "dark" });
  await page.setViewportSize({ width: 1080, height: 760 });
  await page.goto("/?preview=1&window=review");
  await expect(
    page.getByRole("heading", { name: "完成热力图", exact: true }),
  ).toBeVisible();
  await page.screenshot({
    path: "../.impeccable/review/refined-review-dark.png",
    animations: "disabled",
  });
  await page.getByRole("button", { name: "复盘", exact: true }).click();
  await page.getByRole("button", { name: /本周工作回顾 创建/ }).click();
  await page.getByLabel("切换笔记预览", { exact: true }).click();
  await expect(page.locator(".note-preview h1")).toHaveText("本周小结");
  await page.screenshot({
    path: "../.impeccable/review/refined-notes-dark.png",
    animations: "disabled",
  });
  await page.setViewportSize({ width: 480, height: 720 });
  await page.goto("/?preview=1&window=focus");
  await expect(page.getByRole("timer")).toBeVisible();
  await page.screenshot({
    path: "../.impeccable/review/refined-focus-dark.png",
    animations: "disabled",
  });
  expect(errors).toEqual([]);
});
