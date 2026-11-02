import { test, expect } from "@playwright/test";
import { call, root, row, select } from "./helpers";

test.beforeEach(async ({ request }) => {
  const p = await call(
    request,
    "PreviewImport",
    JSON.stringify({ formatVersion: 1, tasks: [] }),
    "json",
    null,
    true,
  );
  await call(request, "CommitImport", p.token);
  await call(request, "Settings", {
    theme: "light",
    animation: "false",
    dayShowCompleted: "true",
    calendarShowCompleted: "false",
    calendarShowChildren: "true",
    showTags: "true",
    dayFilterTags: "[]",
    inboxFilterTags: "[]",
    expandDepth: "3",
    autoPostpone: "false",
  });
  await call(request, "SaveTags", [
    { id: "important", name: "重要", order: 0, color: "#b93838" },
    { id: "urgent", name: "紧急", order: 1, color: "#b45309" },
  ]);
});

test("标签浮层在窄窗口底部完整显示、内部滚动并能搜索选择", async ({
  page,
  request,
}) => {
  await call(
    request,
    "SaveTags",
    Array.from({ length: 30 }, (_, i) => ({
      id: `tag-${i}`,
      name: `项目标签 ${String(i).padStart(2, "0")}`,
      order: i,
      color: "#3563c1",
    })),
  );
  for (let i = 0; i < 12; i++) await root(request, `浮层任务 ${i}`);
  await page.setViewportSize({ width: 400, height: 620 });
  await page.goto("/?preview=1");
  const task = row(page, "浮层任务 11");
  await task.scrollIntoViewIfNeeded();
  await task.hover();
  const trigger = task.getByLabel("设置标签 浮层任务 11", { exact: true });
  await trigger.click();
  const popup = page.getByRole("dialog", { name: "选择任务标签", exact: true });
  await expect(popup).toBeVisible();
  expect(await popup.evaluate((el) => el.matches(":popover-open"))).toBe(true);
  const b = (await popup.boundingBox())!;
  expect(b.x).toBeGreaterThanOrEqual(7);
  expect(b.y).toBeGreaterThanOrEqual(7);
  expect(b.x + b.width).toBeLessThanOrEqual(393);
  expect(b.y + b.height).toBeLessThanOrEqual(613);
  expect(
    await popup
      .locator(".tag-options")
      .evaluate((el) => el.scrollHeight > el.clientHeight),
  ).toBe(true);
  await popup
    .getByRole("checkbox", { name: "项目标签 29", exact: true })
    .scrollIntoViewIfNeeded();
  await popup
    .getByRole("checkbox", { name: "项目标签 29", exact: true })
    .check();
  await expect(popup).toBeVisible();
  await popup.getByLabel("搜索或新增标签").fill("29");
  await expect(popup.getByRole("checkbox")).toHaveCount(1);
  await expect(popup.getByRole("checkbox")).toBeChecked();
  await page.screenshot({
    path: "../.impeccable/review/issue7-tags-light.png",
  });
  await page.keyboard.press("Escape");
  await expect(popup).toHaveCount(0);
  await expect(trigger).toBeFocused();
});

test("标签颜色设置保存并跨页面生效，浅深主题都有填充色", async ({
  page,
  request,
}) => {
  const t = await root(request, "带颜色的任务");
  await call(
    request,
    "SetTaskTags",
    {
      commandId: crypto.randomUUID(),
      id: t.id,
      rootId: t.id,
      expectedVersion: t.version,
    },
    ["important"],
  );
  await page.goto("/?preview=1");
  await page.getByLabel("打开设置", { exact: true }).click();
  await page.getByRole("button", { name: "标签", exact: true }).click();
  await select(page, "标签颜色 1", "蓝色");
  await page
    .getByRole("button", { name: "保存标签与顺序", exact: true })
    .click();
  await expect
    .poll(async () => (await call(request, "Tags"))[0].color)
    .toBe("#3563c1");
  await expect(
    page.getByRole("button", { name: "保存标签与顺序", exact: true }),
  ).toBeEnabled();
  await page.screenshot({
    path: "../.impeccable/review/issue7-tag-settings.png",
  });
  await page.goto("/?preview=1");
  const chip = row(page, t.title).locator(".task-tags .tag-chip");
  await expect(chip).toHaveCSS("background-color", "rgb(53, 99, 193)");
  await expect(chip).toHaveCSS("color", "rgb(255, 255, 255)");
  await call(request, "Settings", { theme: "dark" });
  await page.reload();
  await expect(chip).toHaveCSS("background-color", "rgb(53, 99, 193)");
  await row(page, t.title).hover();
  await row(page, t.title)
    .getByLabel(`设置标签 ${t.title}`, { exact: true })
    .click();
  await page.screenshot({ path: "../.impeccable/review/issue7-tags-dark.png" });
});

test("日历与甘特图默认隐藏已完成任务并记住独立的显示选择", async ({
  page,
  request,
}) => {
  const a = await root(request, "日历待办");
  const b = await root(request, "日历已完成");
  await call(request, "Create", {
    commandId: crypto.randomUUID(),
    rootId: a.id,
    parentId: a.id,
    title: "已完成子任务",
    expectedVersion: a.version,
  });
  let q = await call(request, "Query", { view: "day", date: a.date });
  const child = q.tasks.find((t: any) => t.title === "已完成子任务");
  const parent = q.tasks.find((t: any) => t.id === a.id);
  await call(request, "Complete", {
    commandId: crypto.randomUUID(),
    id: child.id,
    rootId: a.id,
    expectedVersion: parent.version,
    completed: true,
  });
  await call(request, "Complete", {
    commandId: crypto.randomUUID(),
    id: b.id,
    rootId: b.id,
    expectedVersion: b.version,
    completed: true,
  });
  // Add another pending child to keep the parent pending.
  q = await call(request, "Query", { view: "day", date: a.date });
  const latest = q.tasks.find((t: any) => t.id === a.id);
  await call(request, "Create", {
    commandId: crypto.randomUUID(),
    rootId: a.id,
    parentId: a.id,
    title: "待办子任务",
    expectedVersion: latest.version,
  });
  await page.setViewportSize({ width: 1100, height: 760 });
  await page.goto("/?preview=1&window=calendar");
  const show = page.getByRole("checkbox", {
    name: "显示已完成任务",
    exact: true,
  });
  await expect(show).not.toBeChecked();
  await expect(
    page.getByLabel("编辑任务：日历已完成", { exact: true }),
  ).toHaveCount(0);
  await expect(
    page.getByLabel("编辑任务：已完成子任务", { exact: true }),
  ).toHaveCount(0);
  await show.check();
  await expect(
    page.getByLabel("编辑任务：日历已完成", { exact: true }),
  ).toBeVisible();
  await select(page, "日历范围", "甘特图");
  await expect(
    page.locator(".gantt-row .gantt-label").filter({ hasText: "日历已完成" }),
  ).toBeVisible();
  await expect(
    page.locator(".gantt-row .gantt-label").filter({ hasText: "已完成子任务" }),
  ).toBeVisible();
  await show.uncheck();
  await expect(
    page.locator(".gantt-row .gantt-label").filter({ hasText: "日历已完成" }),
  ).toHaveCount(0);
  await expect(
    page.locator(".gantt-row .gantt-label").filter({ hasText: "已完成子任务" }),
  ).toHaveCount(0);
  await page.screenshot({ path: "../.impeccable/review/issue7-gantt.png" });
  await show.check();
  await page.reload();
  await expect(show).toBeChecked();
  await select(page, "日历范围", "甘特图");
  await expect(
    page.locator(".gantt-row .gantt-label").filter({ hasText: "日历已完成" }),
  ).toBeVisible();
  expect((await call(request, "Boot")).settings.dayShowCompleted).toBe("true");
});

test("日历抽屉中的标签浮层选择和关闭不影响抽屉", async ({ page, request }) => {
  const t = await root(request, "抽屉标签任务");
  await page.setViewportSize({ width: 1000, height: 720 });
  await page.goto("/?preview=1&window=calendar");
  await page.getByRole("button", { name: t.date, exact: true }).click();
  const drawer = page.locator(".day-drawer");
  const task = row(page, t.title);
  await task.hover();
  await task.getByLabel(`设置标签 ${t.title}`, { exact: true }).click();
  const popup = page.getByRole("dialog", { name: "选择任务标签", exact: true });
  await popup.getByRole("checkbox", { name: "重要", exact: true }).check();
  await expect(drawer).toBeVisible();
  await expect(
    popup.getByRole("checkbox", { name: "重要", exact: true }),
  ).toBeChecked();
  await page.keyboard.press("Escape");
  await expect(popup).toHaveCount(0);
  await expect(drawer).toBeVisible();
  await expect(task.locator(".task-tags")).toContainText("重要");
});
