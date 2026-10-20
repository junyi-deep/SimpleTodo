import { test, expect } from "@playwright/test";
import { call, root, row } from "./helpers";
test("README 实际界面截图与紧凑窗口布局", async ({ page, request }) => {
  const current = await call(request, "Query", { view: "all", limit: 10000 });
  const ids = current.tasks
    .filter((t: any) => !t.parentId)
    .map((t: any) => t.id);
  if (ids.length) {
    const p = await call(
      request,
      "PreviewBatch",
      ids,
      "delete",
      current.revision,
    );
    await call(request, "CommitBatch", p.token);
  }
  await call(request, "Settings", {
    theme: "light",
    notificationPrompted: "true",
    animation: "false",
  });
  const boot = await call(request, "Boot");
  const first = await root(request, "完成 SimpleTodo 界面整理"),
    parent = await root(request, "准备周五的产品讨论");
  for (const title of ["整理需要确认的问题", "补充关键页面截图"]) {
    const now = await call(request, "FindTask", parent.id);
    await call(request, "Create", {
      commandId: crypto.randomUUID(),
      title,
      parentId: parent.id,
      rootId: parent.id,
      expectedVersion: now.version,
    });
  }
  await root(request, "阅读 20 分钟");
  await root(request, "把周末的安排记下来");
  const done = await root(request, "归档本周资料");
  await call(request, "Complete", {
    commandId: crypto.randomUUID(),
    id: done.id,
    rootId: done.id,
    expectedVersion: done.version,
    completed: true,
  });
  const r = await call(request, "FindTask", first.id),
    when = new Date(Date.now() + 3 * 3600000);
  await call(request, "Update", {
    commandId: crypto.randomUUID(),
    id: r.id,
    rootId: r.id,
    expectedVersion: r.version,
    title: r.title,
    date: r.date,
    description: "精简导航，任务保持在当前窗口。",
    color: "none",
    postpone: "off",
    remindAt: when.getTime(),
    reminderTime: `${String(when.getHours()).padStart(2, "0")}:${String(when.getMinutes()).padStart(2, "0")}`,
  });
  await page.setViewportSize({ width: 600, height: 580 });
  await page.goto("/?preview=1");
  await row(page, parent.title).getByLabel("展开子任务").click();
  await expect(
    page.getByRole("button", { name: "补充关键页面截图", exact: true }),
  ).toBeVisible();
  await page.locator(".brand").hover();
  await page.screenshot({
    path: "../docs/images/tasks.png",
    animations: "disabled",
  });
  await page.setViewportSize({ width: 440, height: 480 });
  await expect(page.getByLabel("新待办标题")).toBeVisible();
  expect(
    await page.locator("body").evaluate((el) => el.scrollWidth <= innerWidth),
  ).toBe(true);
  await page.screenshot({
    path: "../.impeccable/review/redesign-list-440.png",
    animations: "disabled",
  });
  await page.setViewportSize({ width: 1080, height: 720 });
  await page.getByLabel("日历视图", { exact: true }).click();
  await page.getByRole("button", { name: boot.today, exact: true }).click();
  const drawer = page.getByLabel("日期任务抽屉");
  await expect(drawer).toBeVisible();
  await drawer
    .getByRole("treeitem")
    .filter({
      has: page.getByRole("button", { name: parent.title, exact: true }),
    })
    .getByLabel("展开子任务")
    .click();
  await page.locator(".brand").hover();
  await page.screenshot({
    path: "../docs/images/calendar.png",
    animations: "disabled",
  });
  await call(request, "Settings", { theme: "dark" });
  await page.reload();
  await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
  await page.screenshot({
    path: "../.impeccable/review/redesign-dark-1080.png",
    animations: "disabled",
  });
  await call(request, "Settings", { theme: "light" });
});
