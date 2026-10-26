import { test, expect } from "@playwright/test";
import { call, root, row, select } from "./helpers";
test.beforeEach(async ({ request }) => {
  await call(request, "Settings", { expandDepth: "1" });
});
const nextDay = (day: string, n = 1) => {
  const d = new Date(day + "T12:00:00Z");
  d.setUTCDate(d.getUTCDate() + n);
  return d.toISOString().slice(0, 10);
};
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
    completedBottom: "false",
  });
});
test("批量选择框、展开图标、完成框和标题共享首行中心", async ({
  page,
  request,
}) => {
  const p = await root(request, "需要拆分的任务"),
    leaf = await root(request, "独立事项");
  await call(request, "Create", {
    commandId: crypto.randomUUID(),
    rootId: p.id,
    parentId: p.id,
    title: "分解步骤",
    expectedVersion: p.version,
  });
  await page.goto("/?preview=1");
  await row(page, p.title).getByLabel("展开子任务").click();
  await page.getByRole("button", { name: "选择", exact: true }).click();
  for (const t of [p, leaf]) {
    const group = page
      .locator(".selectable-root")
      .filter({ has: page.locator(`[data-node="${t.id}"]`) });
    const targets = [
      group.locator(".root-select"),
      group.locator("[data-node]").first().locator(".tree-toggle"),
      group.locator("[data-node]").first().locator(".task-check"),
      group.locator("[data-node]").first().locator(".task-title"),
    ];
    const centers = await Promise.all(
      targets.map(async (el) => {
        const b = (await el.boundingBox())!;
        return b.y + b.height / 2;
      }),
    );
    expect(Math.max(...centers) - Math.min(...centers)).toBeLessThanOrEqual(1);
  }
  await page.getByLabel("选择 " + leaf.title, { exact: true }).check();
  await expect(page.getByLabel("批量完成", { exact: true })).toBeEnabled();
  await page.screenshot({
    path: "../.impeccable/review/calendar-batch-alignment.png",
    animations: "disabled",
  });
});

test("日历显示全部任务并内部滚动，蓝底白字，日期清单和任务详情在右侧切换", async ({
  page,
  request,
}) => {
  let first: any;
  for (let i = 1; i <= 40; i++) {
    const t = await root(request, `日历事项 ${String(i).padStart(2, "0")}`);
    first ||= t;
  }
  await page.setViewportSize({ width: 1080, height: 720 });
  await page.goto("/?preview=1");
  await page.getByLabel("日历视图", { exact: true }).click();
  const day = page.locator(`[data-day="${first.date}"]`),
    items = day.locator(".calendar-task-list");
  await expect(items.getByRole("button")).toHaveCount(40);
  expect(await items.evaluate((el) => el.scrollHeight > el.clientHeight)).toBe(
    true,
  );
  await expect(items).toHaveCSS("overflow-y", "auto");
  const rail = day.getByRole("scrollbar");
  await expect(rail).toBeVisible();
  await rail.focus();
  await rail.press("End");
  await expect(items.getByRole("button").last()).toBeInViewport();
  await rail.press("Home");
  const task = items.locator(".calendar-task").first();
  await expect(task).toHaveCSS("color", "rgb(255, 255, 255)");
  await expect(task).toHaveCSS("background-color", "rgb(37, 99, 235)");
  await page.screenshot({
    path: "../.impeccable/review/calendar-scroll-month.png",
    animations: "disabled",
  });
  await page.getByRole("button", { name: first.date, exact: true }).click();
  const drawer = page.getByLabel("日期任务抽屉", { exact: true });
  await expect(drawer.locator("[data-node]")).toHaveCount(40);
  await drawer.getByRole("button", { name: first.title, exact: true }).click();
  const details = drawer.getByRole("region", {
    name: "任务详细设置",
    exact: true,
  });
  await expect(details).toBeVisible();
  await expect(page.getByRole("dialog")).toHaveCount(0);
  await details.getByLabel("任务标题", { exact: true }).fill("已在抽屉中修改");
  await details.getByRole("button", { name: "保存修改", exact: true }).click();
  await expect(
    drawer.getByRole("button", { name: "已在抽屉中修改", exact: true }),
  ).toBeVisible();
  expect((await call(request, "FindTask", first.id)).title).toBe(
    "已在抽屉中修改",
  );
  await page.getByLabel("关闭日期任务", { exact: true }).click();
  await day
    .getByRole("button", { name: "编辑任务：已在抽屉中修改", exact: true })
    .click();
  await expect(details).toBeVisible();
  await page.screenshot({
    path: "../.impeccable/review/calendar-drawer-editor.png",
    animations: "disabled",
  });
  await page.keyboard.press("Escape");
  await expect(details).toHaveCount(0);
  await expect(drawer).toBeVisible();
  await page.getByLabel("关闭日期任务", { exact: true }).click();
  await select(page, "日历范围", "周视图");
  await expect(day.locator(".calendar-task")).toHaveCount(40);
  await expect(day.getByRole("scrollbar")).toBeVisible();
  await page.screenshot({
    path: "../.impeccable/review/calendar-scroll-week.png",
    animations: "disabled",
  });
});

test("真实指针拖动日历任务改期并保留子树，拖后不误开详情", async ({
  page,
  request,
}) => {
  const t = await root(request, "拖动改期的任务");
  await call(request, "Create", {
    commandId: crypto.randomUUID(),
    rootId: t.id,
    parentId: t.id,
    title: "一起移动的步骤",
    expectedVersion: t.version,
  });
  await page.setViewportSize({ width: 1080, height: 720 });
  await page.goto("/?preview=1");
  await page.getByLabel("日历视图", { exact: true }).click();
  const targetDate = nextDay(t.date),
    source = page.locator(`[data-calendar-task="${t.id}"]`);
  const a = (await source.boundingBox())!,
    b = (await page
      .getByRole("button", { name: targetDate, exact: true })
      .boundingBox())!;
  await page.mouse.move(a.x + 20, a.y + a.height / 2);
  await page.mouse.down();
  await page.mouse.move(a.x + 30, a.y + a.height / 2, { steps: 3 });
  await expect(page.locator(".calendar-drag-preview")).toHaveCount(1);
  await page.mouse.move(b.x + b.width / 2, b.y + b.height / 2, { steps: 12 });
  await page.mouse.up();
  await expect
    .poll(async () => (await call(request, "FindTask", t.id)).date)
    .toBe(targetDate);
  await expect(page.getByLabel("日期任务抽屉", { exact: true })).toHaveCount(0);
  await expect(
    page.locator(`[data-day="${targetDate}"] [data-calendar-task="${t.id}"]`),
  ).toBeVisible();
  expect((await call(request, "ReadTree", t.id)).length).toBe(2);
  await page.reload();
  await page.getByLabel("日历视图", { exact: true }).click();
  await expect(
    page.locator(`[data-day="${targetDate}"] [data-calendar-task="${t.id}"]`),
  ).toBeVisible();
});

test("抽屉详情保留冲突草稿，取消离开不会关闭编辑", async ({
  page,
  request,
}) => {
  const t = await root(request, "详情保护");
  await page.setViewportSize({ width: 1080, height: 720 });
  await page.goto("/?preview=1");
  await page.getByLabel("日历视图", { exact: true }).click();
  await page
    .getByRole("button", { name: "编辑任务：" + t.title, exact: true })
    .click();
  const details = page.getByRole("region", {
    name: "任务详细设置",
    exact: true,
  });
  await details.getByLabel("任务标题", { exact: true }).fill("保留我的草稿");
  await page.getByRole("button", { name: "关闭日期任务", exact: true }).click();
  await page.locator(".component-confirmation").getByRole("button", { name: "取消", exact: true }).click();
  await expect(details).toBeVisible();
  await expect(details.getByLabel("任务标题", { exact: true })).toHaveValue(
    "保留我的草稿",
  );
  const now = await call(request, "FindTask", t.id);
  await call(request, "Update", {
    commandId: crypto.randomUUID(),
    id: now.id,
    rootId: now.id,
    expectedVersion: now.version,
    title: "其他窗口的修改",
    date: now.date,
    color: "none",
    postpone: "off",
  });
  await details.getByRole("button", { name: "保存修改", exact: true }).click();
  await expect(details.getByRole("alert")).toContainText("CONFLICT");
  await expect(details.getByLabel("任务标题", { exact: true })).toHaveValue(
    "保留我的草稿",
  );
});
