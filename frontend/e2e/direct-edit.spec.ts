import { test, expect } from "@playwright/test";
import { call, root } from "./helpers";
test.beforeEach(async ({ request }) => {
  await call(request, "Settings", { expandDepth: "1" });
});
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
    animation: "true",
    theme: "light",
    completedBottom: "false",
  });
});

test("截图问题：稳定父子缩进、数字图标对齐、搜索铺满和子任务输入无边框", async ({
  page,
  request,
}) => {
  const parent = await root(request, "准备下一次产品讨论");
  let prev = parent;
  for (const title of ["整理待确认的问题", "补充讨论材料"]) {
    const version = await call(request, "FindTask", parent.id);
    const tree = await call(request, "Create", {
      commandId: crypto.randomUUID(),
      title,
      parentId: prev.id,
      rootId: parent.id,
      expectedVersion: version.version,
    });
    prev = tree.find((t: any) => t.title === title);
  }
  await page.goto("/?preview=1");
  await page.getByLabel("展开子任务").click();
  await page.getByLabel("展开子任务").click();
  const rows = page.locator("[data-node]");
  const checks = await rows.locator(".task-check").all();
  const boxes = await Promise.all(checks.map((c) => c.boundingBox()));
  expect(boxes[1]!.x - boxes[0]!.x).toBeCloseTo(22, 0);
  expect(boxes[2]!.x - boxes[1]!.x).toBeCloseTo(22, 0);
  const first = rows.first();
  await first.hover();
  const progress = (await first.locator(".task-progress").boundingBox())!,
    icon = (await first.getByLabel("设置提醒", { exact: true }).boundingBox())!;
  expect(
    Math.abs(progress.y + progress.height / 2 - icon.y - icon.height / 2),
  ).toBeLessThan(1);
  const toolbar = (await page.locator(".list-toolbar").boundingBox())!,
    search = (await page.getByLabel("搜索当前视图").boundingBox())!;
  expect(search.width).toBeGreaterThan(toolbar.width - 145);
  await page.getByLabel("搜索当前视图").focus();
  await expect(page.getByLabel("搜索当前视图")).toHaveCSS(
    "outline-style",
    "none",
  );
  await first.focus();
  await page.keyboard.press("Shift+Enter");
  const child = page.getByLabel("新待办标题", { exact: true });
  await child.fill("补充新的事项");
  await expect(child).toHaveCSS("outline-style", "none");
  await expect(child).toHaveCSS("border-top-width", "0px");
  await page.screenshot({
    path: "../.impeccable/review/direct-child-input.png",
    animations: "disabled",
  });
  await page.keyboard.press("Escape");
  await page.setViewportSize({ width: 400, height: 720 });
  await first.hover();
  await expect(first.getByLabel("删除任务", { exact: true })).toBeInViewport();
  await page.screenshot({
    path: "../.impeccable/review/direct-400.png",
    animations: "disabled",
  });
});

test("单击编辑与输入星星，操作只在 hover 或键盘操作时滑出", async ({
  page,
  request,
}) => {
  const t = await root(request, "单击从这里继续");
  await page.goto("/?preview=1");
  const r = page.locator(`[data-node="${t.id}"]`),
    actions = r.locator(".task-actions");
  await page.locator(".brand").hover();
  await expect(actions).toHaveCSS("opacity", "0");
  await r.hover();
  await expect(actions).toHaveCSS("opacity", "1");
  await expect(r.getByLabel("更多任务操作")).toHaveCount(0);
  await r.getByRole("button", { name: t.title, exact: true }).click();
  const input = page.getByLabel("编辑任务标题", { exact: true });
  await expect(input).toBeFocused();
  await input.press("End");
  await input.pressSequentially("ABC", { delay: 30 });
  await expect(page.locator(".typing-stars")).toHaveCount(1);
  await page.screenshot({ path: "../.impeccable/review/direct-stars.png" });
  await page.keyboard.press("Enter");
  await expect
    .poll(async () => (await call(request, "FindTask", t.id)).title)
    .toBe(t.title + "ABC");
  await expect(page.locator(".typing-stars")).toHaveCount(0);
  await page.emulateMedia({ reducedMotion: "reduce" });
  await r.getByRole("button", { name: t.title + "ABC", exact: true }).click();
  await input.pressSequentially("D");
  await expect(page.locator(".typing-stars")).toHaveCount(0);
  await page.keyboard.press("Escape");
});

test("删除两次确认，取消和超时不删除，确认删除完整分支", async ({
  page,
  request,
}) => {
  const t = await root(request, "谨慎删除整个分支");
  await call(request, "Create", {
    commandId: crypto.randomUUID(),
    rootId: t.id,
    parentId: t.id,
    title: "保留到确认后",
    expectedVersion: t.version,
  });
  await page.goto("/?preview=1");
  const r = page.locator(`[data-node="${t.id}"]`);
  await r.hover();
  await r.getByLabel("删除任务", { exact: true }).click();
  const confirm = r.getByLabel("确认删除任务", { exact: true });
  await expect(confirm).toHaveCSS("color", "rgb(36, 128, 70)");
  expect((await call(request, "ReadTree", t.id)).length).toBe(2);
  await page.keyboard.press("Escape");
  await expect(confirm).toHaveCount(0);
  await r.hover();
  await r.getByLabel("删除任务", { exact: true }).click();
  await page.getByLabel("搜索当前视图").click();
  await expect(confirm).toHaveCount(0);
  await r.hover();
  await r.getByLabel("删除任务", { exact: true }).click();
  await expect(confirm).toHaveCount(0, { timeout: 6000 });
  await r.getByLabel("删除任务", { exact: true }).click();
  await confirm.click();
  await expect(r).toHaveCount(0);
  const q = await call(request, "Query", { view: "all" });
  expect(q.tasks).toHaveLength(0);
});

test("单击编辑后 Shift+Enter 保存并直接添加子任务", async ({
  page,
  request,
}) => {
  const t = await root(request, "直接拆分");
  await page.goto("/?preview=1");
  await page.getByRole("button", { name: t.title, exact: true }).click();
  await page.getByLabel("编辑任务标题", { exact: true }).press("Shift+Enter");
  await expect(
    page.getByLabel("新待办标题", { exact: true }),
  ).toBeFocused();
});
