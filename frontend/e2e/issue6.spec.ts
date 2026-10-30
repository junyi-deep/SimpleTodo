import { test, expect, type APIRequestContext } from "@playwright/test";
import { call, root, select } from "./helpers";
const shift = (day: string, n: number) => {
  const d = new Date(day + "T12:00:00");
  d.setDate(d.getDate() + n);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
};
async function child(request: APIRequestContext, parent: any, title: string) {
  const tree = await call(request, "ReadTree", parent.rootId);
  const rows = await call(request, "Create", {
    commandId: crypto.randomUUID(),
    rootId: parent.rootId,
    parentId: parent.id,
    expectedVersion: tree.find((n: any) => n.id === parent.rootId).version,
    title,
  });
  return rows.find((n: any) => n.title === title);
}
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
    expandDepth: "3",
    completedBottom: "true",
    dayShowCompleted: "true",
    animation: "false",
    theme: "light",
  });
});

test("已完成子任务前显示新草稿，提交前后位置一致", async ({
  page,
  request,
}) => {
  const r = await root(request, "完成的主任务"),
    c = await child(request, r, "已完成子任务");
  const tree = await call(request, "ReadTree", r.id);
  await call(request, "Complete", {
    commandId: crypto.randomUUID(),
    id: r.id,
    rootId: r.id,
    expectedVersion: tree.find((n: any) => n.id === r.id).version,
    completed: true,
  });
  await page.goto("/?preview=1");
  await page.locator(`[data-node="${r.id}"]`).focus();
  await page.keyboard.press("Shift+Enter");
  const input = page.getByLabel("新待办标题", { exact: true });
  await expect(input).toBeFocused();
  const box = (await input.boundingBox())!,
    old = (await page.locator(`[data-node="${c.id}"]`).boundingBox())!;
  expect(box.y).toBeLessThan(old.y);
  await input.fill("新未完成子任务");
  await input.press("Enter");
  await input.press("Escape");
  expect(
    (await page
      .getByRole("button", { name: "新未完成子任务", exact: true })
      .boundingBox())!.y,
  ).toBeLessThan(
    (await page.locator(`[data-node="${c.id}"]`).boundingBox())!.y,
  );
  await page.screenshot({
    path: "../.impeccable/review/issue6-draft-order.png",
  });
});

test("子任务可多选并只处理所选分支，Esc 从复选框或搜索退出选择", async ({
  page,
  request,
}) => {
  const r = await root(request, "批量父任务"),
    a = await child(request, r, "选中的分支"),
    leaf = await child(request, a, "分支后代"),
    b = await child(request, r, "不选的兄弟");
  await page.goto("/?preview=1");
  await page.getByLabel("选择", { exact: true }).click();
  await page
    .getByRole("checkbox", { name: "选择 " + a.title, exact: true })
    .check();
  await page
    .getByRole("checkbox", { name: "选择 " + leaf.title, exact: true })
    .check();
  await page.getByLabel("批量完成", { exact: true }).click();
  await page
    .getByRole("button", { name: "完成 1 项任务及其后代", exact: true })
    .click();
  await expect(page.locator(`[data-node="${a.id}"]`)).toHaveAttribute(
    "data-status",
    "completed",
  );
  await expect(page.locator(`[data-node="${b.id}"]`)).toHaveAttribute(
    "data-status",
    "pending",
  );
  const tree = await call(request, "ReadTree", r.id);
  expect(tree.find((n: any) => n.id === r.id).status).toBe("pending");
  await page
    .getByRole("checkbox", { name: "选择 " + b.title, exact: true })
    .focus();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("checkbox", { name: /^选择 / })).toHaveCount(0);
  await page.getByLabel("选择", { exact: true }).click();
  await page.getByLabel("搜索当前视图").focus();
  await page.keyboard.press("Escape");
  await expect(page.getByLabel("批量完成", { exact: true })).toHaveCount(0);
});

test("节点展开状态跨视图与重启保留，层级设置按当前层数重置", async ({
  page,
  request,
}) => {
  await call(request, "Settings", { expandDepth: "2" });
  const r = await root(request, "展开记忆根"),
    a = await child(request, r, "手动展开父"),
    b = await child(request, a, "记忆后代");
  await page.goto("/?preview=1");
  await page.locator(`[data-node="${a.id}"]`).getByLabel("展开子任务").click();
  await expect(page.locator(`[data-node="${b.id}"]`)).toBeVisible();
  await page.locator(`[data-node="${r.id}"]`).getByLabel("折叠子任务").click();
  await page.getByLabel("日历视图", { exact: true }).click();
  await page.getByLabel("任务列表", { exact: true }).click();
  await expect(page.locator(`[data-node="${a.id}"]`)).toHaveCount(0);
  await page.reload();
  await expect(page.locator(`[data-node="${a.id}"]`)).toHaveCount(0);
  await page.locator(`[data-node="${r.id}"]`).getByLabel("展开子任务").click();
  await expect(page.locator(`[data-node="${b.id}"]`)).toBeVisible();
  await page.getByLabel("展开层级", { exact: true }).click();
  await expect(page.locator(`[data-node="${b.id}"]`)).toHaveCount(0);
  await expect(page.locator(`[data-node="${a.id}"]`)).toBeVisible();
  await page.reload();
  await expect(page.locator(`[data-node="${b.id}"]`)).toHaveCount(0);
});

test("拖住已选日期滑动，日期改变而选中槽位不变", async ({ page, request }) => {
  const boot = await call(request, "Boot");
  await page.goto("/?preview=1");
  const selected = page.locator(".timeline-days .selected"),
    pill = page.locator(".timeline-pill"),
    box = (await selected.boundingBox())!,
    track = (await page.locator(".timeline-days").boundingBox())!,
    before = (await pill.boundingBox())!;
  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
  await page.mouse.down();
  await page.mouse.move(
    box.x + box.width / 2 + (track.width / 7) * 1.25,
    box.y + box.height / 2,
    { steps: 12 },
  );
  await expect(selected).toHaveAttribute(
    "data-timeline-day",
    shift(boot.today, -1),
  );
  expect(Math.abs((await pill.boundingBox())!.x - before.x)).toBeLessThan(1);
  await page.mouse.up();
  await expect(selected).toHaveAttribute(
    "data-timeline-day",
    shift(boot.today, -1),
  );
  await page.screenshot({
    path: "../.impeccable/review/issue6-date-scrub.png",
  });
});

test("月周星期标题分别居中对齐日期列", async ({ page }) => {
  await page.setViewportSize({ width: 1080, height: 720 });
  await page.goto("/?preview=1");
  await page.getByLabel("日历视图", { exact: true }).click();
  for (const view of ["月视图", "周视图"]) {
    await select(page, "日历范围", view);
    const labels = page.locator(".month-weekdays>span"),
      days = page.locator(".calendar-day");
    for (let i = 0; i < 7; i++) {
      const a = (await labels.nth(i).boundingBox())!,
        b = (await days.nth(i).boundingBox())!;
      expect(Math.abs(a.x + a.width / 2 - b.x - b.width / 2)).toBeLessThan(1);
    }
  }
});

test("本机日程只读显示且不进入待办计数，失败保留缓存并可重试", async ({
  page,
  request,
}) => {
  const boot = await call(request, "Boot"),
    start = new Date(boot.today + "T09:00:00").getTime();
  const state = {
    sources: ["native-outlook"],
    lastSync: Date.now(),
    events: [
      {
        id: "native-event",
        calendarId: "native-outlook",
        title: "本机测试会议",
        location: "会议室",
        start,
        end: start + 3600000,
        allDay: false,
      },
    ],
  };
  let fail = false;
  await page.route("**/api/SyncLocalCalendar", (route) =>
    route.fulfill({
      status: fail ? 400 : 200,
      json: fail ? { error: "测试权限已撤销" } : state,
    }),
  );
  await page.route("**/api/LocalCalendarState", (route) =>
    route.fulfill({ json: state }),
  );
  await page.setViewportSize({ width: 1080, height: 720 });
  await page.goto("/?preview=1");
  await page.getByLabel("日历视图", { exact: true }).click();
  await expect(
    page.getByLabel("查看本机日程：本机测试会议", { exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: boot.today, exact: true }).click();
  const drawer = page.getByLabel("日期任务抽屉", { exact: true });
  await expect(drawer.getByText("0 项任务", { exact: true })).toBeVisible();
  await drawer.getByRole("button", { name: /本机测试会议/ }).click();
  await expect(
    page.getByRole("dialog", { name: "本机日程详情", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "保存修改", exact: true }),
  ).toHaveCount(0);
  await page.keyboard.press("Escape");
  fail = true;
  await page.getByLabel("刷新本机日程", { exact: true }).click();
  await expect(page.getByText("测试权限已撤销", { exact: true })).toBeVisible();
  await expect(
    page.getByLabel("查看本机日程：本机测试会议", { exact: true }),
  ).toBeVisible();
  fail = false;
  await page.getByLabel("刷新本机日程", { exact: true }).click();
  await expect(page.getByText("测试权限已撤销", { exact: true })).toHaveCount(
    0,
  );
  await page.screenshot({
    path: "../.impeccable/review/issue6-local-calendar.png",
  });
});

test("本机日历设置可选择来源、保存与停止显示", async ({ page }) => {
  let selected: string[] = [];
  const state = () => ({
    sources: selected,
    lastSync: selected.length ? Date.now() : 0,
    events: [],
  });
  await page.route("**/api/LocalCalendarState", (route) =>
    route.fulfill({ json: state() }),
  );
  await page.route("**/api/LocalCalendarSources", (route) =>
    route.fulfill({
      json: [{ id: "outlook-work", name: "工作日历", source: "本机测试账户" }],
    }),
  );
  await page.route("**/api/ConfigureLocalCalendar", (route) => {
    selected = route.request().postDataJSON()[0];
    return route.fulfill({ json: state() });
  });
  await page.goto("/?preview=1");
  await page.getByLabel("打开设置", { exact: true }).click();
  await page.getByRole("button", { name: "本地数据", exact: true }).click();
  const settings = page.getByRole("region", {
    name: "本机 Outlook 日程",
    exact: true,
  });
  await settings
    .getByRole("button", { name: "选择本机日历", exact: true })
    .click();
  await settings.getByRole("checkbox", { name: /工作日历/ }).check();
  await settings
    .getByRole("button", { name: "保存并读取日程", exact: true })
    .click();
  await expect(
    settings.getByRole("button", { name: "停止显示", exact: true }),
  ).toBeVisible();
  expect(selected).toEqual(["outlook-work"]);
  await settings.getByRole("button", { name: "停止显示", exact: true }).click();
  await expect(
    settings.getByRole("button", { name: "停止显示", exact: true }),
  ).toHaveCount(0);
  expect(selected).toEqual([]);
});
