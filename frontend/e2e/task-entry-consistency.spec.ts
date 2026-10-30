import { test, expect, type APIRequestContext } from "@playwright/test";
import { call, root } from "./helpers";

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
    animation: "false",
    theme: "light",
    expandDepth: "5",
    completedBottom: "false",
    autoPostpone: "false",
    dayShowCompleted: "true",
    tabEntry: "false",
  });
});
async function child(request: APIRequestContext, parent: any, title: string) {
  const tree = await call(request, "ReadTree", parent.rootId);
  const rows = await call(request, "Create", {
    commandId: crypto.randomUUID(),
    title,
    parentId: parent.id,
    rootId: parent.rootId,
    expectedVersion: tree.find((t: any) => t.id === parent.rootId).version,
  });
  return rows.find((t: any) => t.title === title);
}

test("已完成主任务可就地编辑，添加子任务自动恢复未完成且保留已有完成项", async ({
  page,
  request,
}) => {
  const r = await root(request, "已完成项目");
  const old = await child(request, r, "已做完的步骤");
  let tree = await call(request, "ReadTree", r.id);
  await call(request, "Complete", {
    commandId: crypto.randomUUID(),
    id: r.id,
    rootId: r.id,
    expectedVersion: tree.find((t: any) => t.id === r.id).version,
    completed: true,
  });
  await page.goto("/?preview=1");
  const row = page.locator(`[data-node="${r.id}"]`);
  await row.getByRole("button", { name: r.title, exact: true }).click();
  await page
    .getByLabel("编辑任务标题", { exact: true })
    .fill("修改完成项目标题");
  await page.getByLabel("搜索当前视图").click();
  await expect(row).toHaveAttribute("data-status", "completed");
  await row
    .getByRole("button", { name: "修改完成项目标题", exact: true })
    .click();
  await page.getByLabel("编辑任务标题", { exact: true }).press("Shift+Enter");
  const input = page.getByLabel("新待办标题", { exact: true });
  await input.fill("追加的待办步骤");
  await input.press("Enter");
  await input.press("Escape");
  await expect(row).toHaveAttribute("data-status", "pending");
  tree = await call(request, "ReadTree", r.id);
  expect(tree.find((t: any) => t.id === old.id).status).toBe("completed");
  expect(tree.find((t: any) => t.title === "追加的待办步骤").parentId).toBe(
    r.id,
  );
});

test("同级草稿位于 A 整个展开分支之后，连续录入保持实际同级顺序", async ({
  page,
  request,
}) => {
  const r = await root(request, "根任务");
  const a = await child(request, r, "父任务 A");
  const deep = await child(request, a, "A 的子任务");
  const leaf = await child(request, deep, "A 的最深后代");
  const tail = await child(request, r, "后续同级任务");
  await page.goto("/?preview=1");
  await page.getByRole("button", { name: a.title, exact: true }).click();
  await page.getByLabel("编辑任务标题", { exact: true }).press("Enter");
  const input = page.getByLabel("新待办标题", { exact: true });
  await expect(input).toBeFocused();
  const bounds = (await input.boundingBox())!;
  const leafBounds = (await page
    .locator(`[data-node="${leaf.id}"]`)
    .boundingBox())!;
  const tailBounds = (await page
    .locator(`[data-node="${tail.id}"]`)
    .boundingBox())!;
  expect(bounds.y).toBeGreaterThanOrEqual(leafBounds.y + leafBounds.height);
  expect(bounds.y + bounds.height).toBeLessThanOrEqual(tailBounds.y);
  await input.fill("同级任务 B");
  await input.press("Enter");
  await expect(input).toBeFocused();
  await input.fill("同级任务 C");
  await input.press("Enter");
  await input.press("Escape");
  const tree = await call(request, "ReadTree", r.id);
  expect(
    tree
      .filter((t: any) => t.parentId === r.id)
      .sort((a: any, b: any) => a.order - b.order)
      .map((t: any) => t.title),
  ).toEqual([a.title, "同级任务 B", "同级任务 C", tail.title]);
  await page.screenshot({
    path: "../.impeccable/review/task-entry-branch-order.png",
  });
});

test("Shift+Tab 在编辑时保存标题并提升一级，直属子任务连同后代成为主任务", async ({
  page,
  request,
}) => {
  const r = await root(request, "原始主任务");
  const a = await child(request, r, "原父任务");
  const b = await child(request, a, "将要提升的分支");
  const leaf = await child(request, b, "保留的后代");
  const other = await root(request, "后面的主任务");
  await page.goto("/?preview=1");
  const source = page.locator(`[data-node="${b.id}"]`);
  await source.getByLabel("任务设置", { exact: true }).focus();
  await page.keyboard.press("Shift+Tab");
  await expect(source.locator(".task-title")).toBeFocused();
  await expect(source).toHaveAttribute("data-parent", a.id);
  await page.getByRole("button", { name: b.title, exact: true }).click();
  const input = page.getByLabel("编辑任务标题", { exact: true });
  await input.fill("提升并保存标题");
  await input.press("Shift+Tab");
  const row = page.locator(`[data-node="${b.id}"]`);
  await expect(row).toHaveAttribute("data-parent", r.id);
  await expect(row).toBeFocused();
  await expect(
    row.getByRole("button", { name: "提升并保存标题", exact: true }),
  ).toBeVisible();
  await row.press("Shift+Tab");
  await expect(row).not.toHaveAttribute("data-parent");
  await expect(row).toBeFocused();
  const promoted = await call(request, "ReadTree", b.id);
  expect(promoted.find((t: any) => t.id === leaf.id).rootId).toBe(b.id);
  const roots = (await call(request, "Query", { view: "all" })).tasks
    .filter((t: any) => !t.parentId)
    .sort((a: any, b: any) => a.order - b.order);
  expect(roots.map((t: any) => t.id)).toEqual([r.id, b.id, other.id]);
  await row.press("Shift+Tab");
  expect(
    (await call(request, "ReadTree", b.id)).find((t: any) => t.id === b.id)
      .parentId,
  ).toBeNull();
});

test("日历日期抽屉与列表共用新增、就地编辑、子任务和详细设置操作", async ({
  page,
  request,
}) => {
  const boot = await call(request, "Boot");
  const date =
    boot.today.slice(0, 8) + (Number(boot.today.slice(-2)) === 1 ? "02" : "01");
  await page.setViewportSize({ width: 1080, height: 720 });
  await page.goto("/?preview=1");
  await page.getByLabel("日历视图", { exact: true }).click();
  await page.getByRole("button", { name: date, exact: true }).click();
  const drawer = page.getByLabel("日期任务抽屉", { exact: true });
  await drawer.getByLabel("新增待办", { exact: true }).click();
  let input = drawer.getByLabel("新待办标题", { exact: true });
  await input.fill("抽屉新增事项");
  await input.press("Enter");
  await input.press("Escape");
  await drawer
    .getByRole("button", { name: "抽屉新增事项", exact: true })
    .click();
  const edit = drawer.getByLabel("编辑任务标题", { exact: true });
  await expect(edit).toBeFocused();
  await expect(
    drawer.getByRole("region", { name: "任务详细设置", exact: true }),
  ).toHaveCount(0);
  await edit.fill("抽屉内直接修改");
  await edit.press("Shift+Enter");
  input = drawer.getByLabel("新待办标题", { exact: true });
  await input.fill("抽屉内子任务");
  await input.press("Enter");
  await input.press("Escape");
  const q = await call(request, "Query", { view: "day", date });
  const r = q.tasks.find((t: any) => !t.parentId);
  expect(r.title).toBe("抽屉内直接修改");
  expect(r.date).toBe(date);
  expect(q.tasks.find((t: any) => t.parentId === r.id).title).toBe(
    "抽屉内子任务",
  );
  const row = drawer.locator(`[data-node="${r.id}"]`);
  await row.hover();
  await row.getByLabel("任务设置", { exact: true }).click();
  await expect(
    drawer.getByRole("region", { name: "任务详细设置", exact: true }),
  ).toBeVisible();
  await expect(page.getByRole("dialog")).toHaveCount(0);
  await page.screenshot({
    path: "../.impeccable/review/task-entry-drawer-settings.png",
  });
});
