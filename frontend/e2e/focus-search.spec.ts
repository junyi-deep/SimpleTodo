import { test, expect } from "@playwright/test";
import { call, root } from "./helpers";

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
    autoPostpone: "false",
    dayShowCompleted: "true",
    notificationPrompted: "true",
  });
});

test("专注任务可搜索选择，清空重搜与无匹配保持原选择，启动实际关联所选任务", async ({
  page,
  request,
}) => {
  const a = await root(request, "整理项目文档"),
    b = await root(request, "Review 设计稿");
  await page.goto("/?preview=1&window=focus");
  const trigger = page.getByRole("combobox", { name: "关联任务", exact: true });
  await trigger.click();
  const search = page.getByRole("combobox", {
    name: "搜索专注任务",
    exact: true,
  });
  await expect(search).toBeFocused();
  await search.fill("项目");
  await expect(
    page.getByRole("option", { name: b.title, exact: true }),
  ).toHaveCount(0);
  await page.getByRole("option", { name: a.title, exact: true }).click();
  await expect(trigger).toContainText(a.title);
  await expect(trigger).toBeFocused();
  await trigger.click();
  await expect(search).toHaveValue("");
  await search.fill("不存在");
  await expect(
    page.getByRole("status").filter({ hasText: "没有匹配的任务" }),
  ).toBeVisible();
  await search.press("Enter");
  await expect(trigger).toContainText(a.title);
  await search.fill("review");
  await expect(
    page.getByRole("option", { name: b.title, exact: true }),
  ).toBeVisible();
  await search.dispatchEvent("compositionstart");
  await search.dispatchEvent("keydown", { key: "Enter", isComposing: true });
  await expect(search).toBeVisible();
  await expect(trigger).toContainText(a.title);
  await search.dispatchEvent("compositionend");
  await search.press("Enter");
  await expect(trigger).toContainText(b.title);
  await page.getByRole("button", { name: "开始专注", exact: true }).click();
  await expect(
    page.getByRole("button", { name: "暂停", exact: true }),
  ).toBeVisible();
  expect((await call(request, "Boot")).focus.taskId).toBe(b.id);
  await call(request, "Focus", "stop", 25, null);
});

test("专注搜索支持空格、方向键和 Escape，长列表滚动与浅深窄窗口浮层不溢出", async ({
  page,
  request,
}) => {
  for (let i = 0; i < 12; i++)
    await root(request, `项目计划 ${String(i + 1).padStart(2, "0")}`);
  await page.setViewportSize({ width: 400, height: 740 });
  await page.goto("/?preview=1&window=focus");
  const trigger = page.getByRole("combobox", { name: "关联任务", exact: true });
  await trigger.focus();
  await trigger.press("Enter");
  const search = page.getByRole("combobox", {
    name: "搜索专注任务",
    exact: true,
  });
  await search.fill("项目");
  await search.press("Space");
  await search.press("1");
  await expect(search).toHaveValue("项目 1");
  // Use the complete prefix to match the task titles, then select the next result.
  await search.fill("项目计划 1");
  await search.press("ArrowDown");
  await search.press("Enter");
  await expect(trigger).toContainText("项目计划 11");
  await trigger.click();
  await expect(page.locator('.select-results [data-active="true"]')).toHaveText(
    "项目计划 11",
  );
  await search.press("Enter");
  await expect(trigger).toContainText("项目计划 11");
  await trigger.click();
  const selected = page.getByRole("option", {
    name: "项目计划 11",
    exact: true,
  });
  const viewport = await page.locator(".select-results").boundingBox(),
    bounds = await selected.boundingBox();
  expect(bounds!.y).toBeGreaterThanOrEqual(viewport!.y);
  expect(bounds!.y + bounds!.height).toBeLessThanOrEqual(
    viewport!.y + viewport!.height + 1,
  );
  await page.screenshot({
    path: "../.impeccable/review/focus-search-light.png",
    animations: "disabled",
  });
  await search.press("Escape");
  await expect(trigger).toBeFocused();
  await expect(search).toBeHidden();
  await trigger.click();
  await search.fill("自由");
  await search.press("Enter");
  await expect(trigger).toContainText("自由专注");
  await call(request, "Settings", { theme: "dark" });
  await page.reload();
  await trigger.click();
  await search.fill("项目计划");
  await page.screenshot({
    path: "../.impeccable/review/focus-search-dark.png",
    animations: "disabled",
  });
  const popup = await page.locator(".select-options.searchable").boundingBox();
  expect(popup!.x).toBeGreaterThanOrEqual(0);
  expect(popup!.x + popup!.width).toBeLessThanOrEqual(400);
  expect(popup!.y + popup!.height).toBeLessThanOrEqual(740);
  await page.locator(".focus-page .quiet-note").click();
  await expect(search).toBeHidden();
  await trigger.click();
  await page.setViewportSize({ width: 480, height: 720 });
  await expect(search).toBeHidden();
});
