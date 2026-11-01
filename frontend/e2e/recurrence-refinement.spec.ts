import { test, expect } from "@playwright/test";
import { call, root, select } from "./helpers";
test.beforeEach(async ({ request }) => {
  await call(request, "Settings", { expandDepth: "1" });
});
const shift = (day: string, n: number) => {
  const d = new Date(day + "T12:00:00");
  d.setDate(d.getDate() + n);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
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
test("年月下拉默认定位选中项，Esc 新增按钮没有轮廓", async ({ page }) => {
  await page.goto("/?preview=1");
  await page.getByLabel("新增待办", { exact: true }).click();
  await page.getByLabel("新待办标题", { exact: true }).press("Escape");
  await expect(page.getByLabel("新增待办", { exact: true })).toBeFocused();
  await expect(page.getByLabel("新增待办", { exact: true })).toHaveCSS(
    "outline-style",
    "none",
  );
  await page.setViewportSize({ width: 1080, height: 720 });
  await page.getByLabel("日历视图", { exact: true }).click();
  for (const label of ["年份", "月份"]) {
    await page.getByRole("combobox", { name: label, exact: true }).click();
    await expect(page.getByRole("option", { selected: true })).toBeInViewport();
    await page.keyboard.press("Escape");
  }
  await select(page, "年份", "2099年");
  await page.getByRole("combobox", { name: "年份", exact: true }).click();
  await expect(
    page.getByRole("option", { name: "2099年", exact: true }),
  ).toBeInViewport();
});
test("Enter 连续插入根任务和同级子任务，Esc 取消空草稿", async ({
  page,
  request,
}) => {
  const a = await root(request, "第一项"),
    b = await root(request, "下一项");
  const snap = await call(request, "Query", { view: "all" });
  await call(request, "Reorder", [a.id, b.id], snap.revision);
  await page.goto("/?preview=1");
  await page.getByRole("button", { name: a.title, exact: true }).click();
  await page.getByLabel("编辑任务标题", { exact: true }).press("Enter");
  let input = page.getByLabel("新待办标题", { exact: true });
  await expect(input).toBeFocused();
  await input.fill("插在中间");
  await input.press("Enter");
  await expect(input).toBeFocused();
  await input.press("Escape");
  let q = await call(request, "Query", { view: "all" });
  const middle = q.tasks.find((t: any) => t.title === "插在中间");
  expect(middle.parentId).toBeNull();
  expect(q.tasks.find((t: any) => t.id === a.id).order).toBeLessThan(
    middle.order,
  );
  expect(middle.order).toBeLessThan(
    q.tasks.find((t: any) => t.id === b.id).order,
  );
  const latest = await call(request, "FindTask", a.id);
  const tree = await call(request, "Create", {
    commandId: crypto.randomUUID(),
    parentId: a.id,
    rootId: a.id,
    expectedVersion: latest.version,
    title: "已有子项",
  });
  const child = tree.find((t: any) => t.parentId);
  await page.reload();
  await page.locator(`[data-node="${a.id}"]`).getByLabel("展开子任务").click();
  await page.getByRole("button", { name: child.title, exact: true }).click();
  await page.getByLabel("编辑任务标题", { exact: true }).press("Enter");
  await input.fill("同级子项");
  await input.press("Enter");
  await input.press("Escape");
  q = await call(request, "ReadTree", a.id);
  expect(q.find((t: any) => t.title === "同级子项").parentId).toBe(a.id);
  expect(q).toHaveLength(3);
});
test("多选日期右键创建持续每日任务，跨月与刷新保留", async ({
  page,
  request,
}) => {
  const boot = await call(request, "Boot");
  await page.setViewportSize({ width: 1080, height: 720 });
  await page.goto("/?preview=1");
  await page.getByLabel("日历视图", { exact: true }).click();
  const first = boot.today,
    last = shift(first, 2);
  await page
    .getByRole("button", { name: first, exact: true })
    .click({ modifiers: ["Control"] });
  await page
    .getByRole("button", { name: last, exact: true })
    .click({ modifiers: ["Control"] });
  await page.locator(`[data-day="${first}"]`).click({ button: "right" });
  await page.getByRole("menuitem", { name: "创建范围任务", exact: true }).click();
  const c = page.getByRole("region", { name: "创建范围任务", exact: true });
  await c.getByLabel("任务标题", { exact: true }).fill("持续每日记录");
  await c.getByLabel("结束日期", { exact: true }).fill("");
  await c.getByRole("button", { name: "创建任务", exact: true }).click();
  const later = shift(first, 33);
  const q = await call(request, "Query", { view: "day", date: later });
  expect(q.tasks.some((n: any) => n.title === "持续每日记录")).toBe(true);
  await page.reload();
  const rules = await call(request, "Rules");
  expect(
    rules.every((r: any) => r.frequency === "daily" && r.interval === 1),
  ).toBe(true);
});

test("范围任务可指定开始结束日期，范围外不生成", async ({ page, request }) => {
  const boot = await call(request, "Boot"),
    start = shift(boot.today, 1),
    end = shift(boot.today, 2);
  await page.setViewportSize({ width: 1080, height: 720 });
  await page.goto("/?preview=1");
  await page.getByLabel("日历视图", { exact: true }).click();
  await page.locator(`[data-day="${boot.today}"]`).click({ button: "right" });
  await page.getByRole("menuitem", { name: "创建范围任务", exact: true }).click();
  const c = page.getByRole("region", { name: "创建范围任务", exact: true });
  await c.getByLabel("任务标题", { exact: true }).fill("有限日期范围");
  await c.getByLabel("开始日期", { exact: true }).fill(start);
  await c.getByLabel("结束日期", { exact: true }).fill(end);
  await c.getByRole("button", { name: "创建任务", exact: true }).click();
  for (const [day, present] of [
    [boot.today, false],
    [start, true],
    [end, true],
    [shift(end, 1), false],
  ] as const) {
    const q = await call(request, "Query", { view: "day", date: day });
    expect(q.tasks.some((n: any) => n.title === "有限日期范围")).toBe(present);
  }
});

test("拖动目标有反馈，停留右边缘自动翻月并可以跨月落下", async ({
  page,
  request,
}) => {
  const t = await root(request, "跨月安排");
  await page.setViewportSize({ width: 1080, height: 720 });
  await page.goto("/?preview=1");
  await page.getByLabel("日历视图", { exact: true }).click();
  const source = page.locator(`[data-calendar-task="${t.id}"]`),
    a = (await source.boundingBox())!,
    grid = page.locator(".calendar-grid"),
    b = (await grid.boundingBox())!;
  const old = await page
    .locator(".calendar-day")
    .first()
    .getAttribute("data-day");
  await page.mouse.move(a.x + 20, a.y + a.height / 2);
  await page.mouse.down();
  await page.mouse.move(b.x + b.width - 12, b.y + b.height / 2, { steps: 15 });
  await expect(page.locator("[data-calendar-drop]")).toHaveCount(1);
  await expect
    .poll(() => page.locator(".calendar-day").first().getAttribute("data-day"))
    .not.toBe(old);
  const dest = page.locator(".calendar-day").nth(15),
    date = (await dest.getAttribute("data-day"))!,
    r = (await dest.boundingBox())!;
  await page.mouse.move(r.x + r.width / 2, r.y + 15, { steps: 8 });
  await expect(dest).toHaveAttribute("data-calendar-drop", "");
  await page.screenshot({
    path: "../.impeccable/review/calendar-drop-feedback.png",
  });
  await page.mouse.up();
  await expect
    .poll(async () => (await call(request, "FindTask", t.id)).date)
    .toBe(date);
  await expect(page.locator("[data-calendar-drop]")).toHaveCount(0);
  await expect(page.locator(".calendar-drag-preview")).toHaveCount(0);
});

test("周历支持边缘翻周、取消拖动清理反馈，抽屉滑入且尊重减少动态效果", async ({
  page,
  request,
}) => {
  const t = await root(request, "翻周与动效");
  await call(request, "Settings", { animation: "true" });
  await page.setViewportSize({ width: 1080, height: 720 });
  await page.goto("/?preview=1");
  await page.getByLabel("日历视图", { exact: true }).click();
  await select(page, "日历范围", "周视图");
  const source = page.locator(`[data-calendar-task="${t.id}"]`),
    a = (await source.boundingBox())!,
    grid = page.locator(".calendar-grid"),
    b = (await grid.boundingBox())!,
    old = await page.locator(".calendar-day").first().getAttribute("data-day");
  await page.mouse.move(a.x + 20, a.y + a.height / 2);
  await page.mouse.down();
  await page.mouse.move(b.x + 8, b.y + 100, { steps: 12 });
  await expect
    .poll(() => page.locator(".calendar-day").first().getAttribute("data-day"))
    .toBe(shift(old!, -7));
  await page.keyboard.press("Escape");
  await page.mouse.up();
  await expect(page.locator(".calendar-drag-preview")).toHaveCount(0);
  await expect(page.locator("[data-calendar-drop]")).toHaveCount(0);
  expect((await call(request, "FindTask", t.id)).date).toBe(t.date);
  await page.getByRole("button", { name: "今天", exact: true }).click();
  await page.getByRole("button", { name: t.date, exact: true }).click();
  const drawer = page.getByLabel("日期任务抽屉", { exact: true });
  await expect(drawer).toBeVisible();
  await expect
    .poll(() =>
      drawer.evaluate((el) => {
        const t = getComputedStyle(el).transform;
        return t === "none" || t === "matrix(1, 0, 0, 1, 0, 0)";
      }),
    )
    .toBe(true);
  await page.getByLabel("关闭日期任务", { exact: true }).click();
  await expect(drawer).toHaveCount(0);
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.getByRole("button", { name: t.date, exact: true }).click();
  await expect(drawer).toBeVisible();
  expect(
    await drawer.evaluate(
      (el) =>
        el.getAnimations().filter((a) => a.playState === "running").length,
    ),
  ).toBe(0);
  await page.screenshot({
    path: "../.impeccable/review/repeat-week-drawer.png",
  });
  await page.getByLabel("关闭日期任务", { exact: true }).click();
  await select(page, "年份", "2028年");
  await expect(page.locator(".calendar-day").first()).toHaveAttribute(
    "data-day",
    /^2028-/,
  );
});
