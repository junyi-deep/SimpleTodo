import { test, expect } from "@playwright/test";
import { call, root, row, select } from "./helpers";

test.beforeEach(async ({ request }) => {
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
    showTags: "false",
    dayShowCompleted: "true",
    dayFilterTags: "[]",
    inboxFilterTags: "[]",
  });
  await call(request, "SaveTags", [
    { id: "important", name: "重要", order: 0 },
    { id: "urgent", name: "紧急", order: 1 },
  ]);
});

async function seed(request: any) {
  const pending = await root(request, "整理本周工作计划"),
    done = await root(request, "完成项目文档"),
    other = await root(request, "安排设计评审");
  await call(
    request,
    "SetTaskTags",
    {
      commandId: crypto.randomUUID(),
      id: pending.id,
      rootId: pending.id,
      expectedVersion: pending.version,
    },
    ["important"],
  );
  await call(request, "Complete", {
    commandId: crypto.randomUUID(),
    id: done.id,
    rootId: done.id,
    expectedVersion: done.version,
    completed: true,
  });
  await call(
    request,
    "SetTaskTags",
    {
      commandId: crypto.randomUUID(),
      id: other.id,
      rootId: other.id,
      expectedVersion: other.version,
    },
    ["urgent"],
  );
  const latest = await call(request, "FindTask", pending.id);
  await call(request, "Create", {
    commandId: crypto.randomUUID(),
    parentId: pending.id,
    rootId: pending.id,
    expectedVersion: latest.version,
    title: "梳理优先级",
  });
  return { pending, done, other };
}

test("关闭全局标签显示时，选中任务和标签选择立即显示标签，离开后遵循全局设置", async ({
  page,
  request,
}) => {
  const { pending, other } = await seed(request);
  await page.setViewportSize({ width: 400, height: 740 });
  await page.goto("/?preview=1");
  const item = page.locator(`[data-node="${pending.id}"]`);
  await expect(item.locator(".task-tags")).toHaveCount(0);
  await row(page, pending.title)
    .getByRole("button", { name: pending.title, exact: true })
    .click();
  await expect(item.locator(".task-tags")).toContainText("重要");
  await item.getByLabel(`设置标签 ${pending.title}`, { exact: true }).click();
  const picker = page.getByRole("dialog", {
    name: "选择任务标签",
    exact: true,
  });
  await picker.getByRole("checkbox", { name: "紧急", exact: true }).check();
  await expect(item.locator(".task-tags")).toContainText("紧急");
  await page.screenshot({
    path: "../.impeccable/review/review-selected-tags.png",
    animations: "disabled",
  });
  await page.keyboard.press("Escape");
  await expect(picker).toBeHidden();
  await expect(item.locator(".task-tags")).toContainText("紧急");
  await row(page, other.title)
    .getByRole("button", { name: other.title, exact: true })
    .click();
  await expect(item.locator(".task-tags")).toHaveCount(0);
  await page.reload();
  expect((await call(request, "FindTask", pending.id)).tagIds).toEqual([
    "important",
    "urgent",
  ]);
  await row(page, pending.title)
    .getByRole("button", { name: pending.title, exact: true })
    .click();
  await expect(item.locator(".task-tags")).toContainText("重要");
});

test("首页饼图反映真实状态，日期状态及子任务筛选共同生效", async ({
  page,
  request,
}) => {
  const { pending } = await seed(request);
  await page.setViewportSize({ width: 1072, height: 712 });
  await page.goto("/?preview=1&window=review");
  await page.getByLabel("统计开始日期").fill(pending.date);
  await page.getByLabel("统计结束日期").fill(pending.date);
  await expect(
    page.getByRole("img", {
      name: "任务完成状态饼图：共 3 项，已完成 1 项，未完成 2 项",
      exact: true,
    }),
  ).toBeVisible();
  await page.getByRole("heading", { name: "任务统计", exact: true }).click();
  await page.screenshot({
    path: "../.impeccable/review/review-home-light.png",
    animations: "disabled",
  });
  await page.getByRole("checkbox", { name: "统计子任务", exact: true }).check();
  await expect(
    page.getByRole("img", {
      name: "任务完成状态饼图：共 4 项，已完成 1 项，未完成 3 项",
      exact: true,
    }),
  ).toBeVisible();
  await select(page, "统计任务状态", "已完成");
  await expect(
    page.getByRole("img", {
      name: "任务完成状态饼图：共 1 项，已完成 1 项，未完成 0 项",
      exact: true,
    }),
  ).toBeVisible();
  await expect(page.locator(".review-pie circle.pie-completed")).toHaveCount(1);
  await select(page, "统计任务状态", "未完成");
  await expect(
    page.getByRole("img", {
      name: "任务完成状态饼图：共 3 项，已完成 0 项，未完成 3 项",
      exact: true,
    }),
  ).toBeVisible();
  await select(page, "统计任务状态", "全部状态");
  await page.getByLabel("统计开始日期").fill("2099-01-01");
  await page.getByLabel("统计结束日期").fill("2099-01-02");
  await expect(page.locator(".review-pie .pie-empty")).toHaveCount(1);
  await expect(page.locator(".review-pie path")).toHaveCount(0);
  await page.getByRole("heading", { name: "任务统计", exact: true }).click();
  await page.screenshot({
    path: "../.impeccable/review/review-empty-light.png",
    animations: "disabled",
  });
  await call(request, "Settings", { theme: "dark" });
  await page.setViewportSize({ width: 800, height: 720 });
  await page.reload();
  await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
  await page.screenshot({
    path: "../.impeccable/review/review-home-dark.png",
    animations: "disabled",
  });
  expect(
    await page.locator("body").evaluate((el) => el.scrollWidth <= innerWidth),
  ).toBe(true);
});

test("任务总览筛选属于左侧，标签为可搜索多选浮层且不挤动报告，支持 Escape 和外部关闭", async ({
  page,
  request,
}) => {
  const { pending, other } = await seed(request);
  const errors: string[] = [];
  page.on("pageerror", (e) => errors.push(e.message));
  await page.setViewportSize({ width: 1072, height: 712 });
  await page.goto("/?preview=1&window=review");
  await page.getByRole("button", { name: "复盘", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "任务总览", exact: true }),
  ).toBeVisible();
  await expect(
    page.locator(".task-report").getByLabel("统计开始日期"),
  ).toBeVisible();
  const heading = await page.locator(".report-heading").boundingBox();
  const pane = await page.locator(".notes-pane").boundingBox();
  expect(Math.abs(heading!.y - pane!.y)).toBeLessThan(3);
  expect(heading!.y).toBeLessThan(90);
  const before = await page.locator(".report-task").first().boundingBox();
  await page.getByRole("button", { name: "筛选报告标签", exact: true }).click();
  const tags = page.getByRole("dialog", { name: "报告标签筛选", exact: true });
  await expect(tags).toBeVisible();
  expect((await page.locator(".report-task").first().boundingBox())!.y).toBe(
    before!.y,
  );
  await tags.getByLabel("搜索筛选标签").fill("重");
  await expect(
    tags.getByRole("checkbox", { name: "紧急", exact: true }),
  ).toHaveCount(0);
  await tags.getByRole("checkbox", { name: "重要", exact: true }).check();
  await expect(page.locator(".report-task")).toHaveCount(1);
  await tags.getByLabel("搜索筛选标签").fill("");
  await tags.getByRole("checkbox", { name: "紧急", exact: true }).check();
  await expect(page.locator(".report-task")).toHaveCount(2);
  await tags.getByRole("checkbox", { name: "紧急", exact: true }).uncheck();
  await expect(page.locator(".report-task")).toHaveCount(1);
  await page.getByRole("button", { name: "首页", exact: true }).click();
  await expect(
    page.getByRole("img", {
      name: "任务完成状态饼图：共 3 项，已完成 1 项，未完成 2 项",
      exact: true,
    }),
  ).toBeVisible();
  await page.getByRole("button", { name: "复盘", exact: true }).click();
  await expect(page.locator(".report-task")).toContainText(pending.title);
  await page.getByRole("button", { name: "筛选报告标签", exact: true }).click();
  await expect(tags).toBeVisible();
  await page.screenshot({
    path: "../.impeccable/review/review-report-light.png",
    animations: "disabled",
  });
  await page.keyboard.press("Escape");
  await expect(tags).toBeHidden();
  await expect(
    page.getByRole("button", { name: "筛选报告标签", exact: true }),
  ).toBeFocused();
  await page.getByRole("button", { name: "筛选报告标签", exact: true }).click();
  await tags.getByRole("button", { name: "清除标签筛选", exact: true }).click();
  await page.getByLabel("搜索任务报告").fill(other.title);
  await expect(tags).toBeHidden();
  await expect(page.locator(".report-task")).toHaveCount(1);
  await page.getByLabel("搜索任务报告").fill("");
  await page.getByRole("button", { name: "筛选报告标签", exact: true }).click();
  await page.setViewportSize({ width: 900, height: 720 });
  await expect(tags).toBeHidden();
  await call(request, "Settings", { theme: "dark" });
  await page.reload();
  await page.getByRole("button", { name: "复盘", exact: true }).click();
  await page.screenshot({
    path: "../.impeccable/review/review-report-dark.png",
    animations: "disabled",
  });
  expect(
    await page.locator("body").evaluate((el) => el.scrollWidth <= innerWidth),
  ).toBe(true);
  expect(errors).toEqual([]);
});
