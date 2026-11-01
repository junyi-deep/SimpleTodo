import { test, expect } from "@playwright/test";
import { call, root, row, select, add } from "./helpers";
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
    expandDepth: "3",
    autoPostpone: "false",
    showTags: "false",
    calendarShowChildren: "false",
    dayShowCompleted: "true",
    inboxShowCompleted: "false",
    skippedUpdateVersion: "",
    dayFilterTags: "[]",
    inboxFilterTags: "[]",
  });
  await call(request, "SaveTags", [
    { id: "important", name: "重要", order: 0 },
    { id: "urgent", name: "紧急", order: 1 },
  ]);
});
test("标签可多选、搜索、新增、筛选和显示，并可设置顺序", async ({
  page,
  request,
}) => {
  const a = await root(request, "有标签任务"),
    b = await root(request, "无标签任务");
  await page.goto("/?preview=1");
  await row(page, a.title).hover();
  await row(page, a.title)
    .getByLabel("设置标签 " + a.title, { exact: true })
    .click();
  const picker = page.getByRole("dialog", {
    name: "选择任务标签",
    exact: true,
  });
  await picker.getByRole("checkbox", { name: "重要", exact: true }).check();
  await picker.getByRole("checkbox", { name: "紧急", exact: true }).check();
  await picker.getByLabel("搜索或新增标签").fill("项目");
  await picker.getByRole("button", { name: "新增“项目”", exact: true }).click();
  await expect(
    picker.getByRole("checkbox", { name: "项目", exact: true }),
  ).toBeChecked();
  await page.keyboard.press("Escape");
  await page.getByLabel("展开层级", { exact: true }).click();
  await page.getByLabel("显示标签", { exact: true }).check();
  await page.keyboard.press("Escape");
  await expect(row(page, a.title).locator(".task-tags")).toContainText("重要");
  await page.getByLabel("过滤显示", { exact: true }).click();
  await page
    .getByRole("group", { name: "任务显示过滤" })
    .getByRole("checkbox", { name: "重要", exact: true })
    .check();
  await page.keyboard.press("Escape");
  await expect(row(page, b.title)).toHaveCount(0);
  await page.reload();
  await expect(row(page, b.title)).toHaveCount(0);
  await page.getByLabel("打开设置", { exact: true }).click();
  await page.getByRole("button", { name: "标签", exact: true }).click();
  await page.getByLabel("下移标签").first().click();
  await page
    .getByRole("button", { name: "保存标签与顺序", exact: true })
    .click();
  expect((await call(request, "Tags"))[0].name).toBe("紧急");
  await expect(
    page.getByRole("button", { name: "保存标签与顺序", exact: true }),
  ).toBeEnabled();
  await page.screenshot({ path: "../.impeccable/review/workspace-tags.png" });
});
test("新任务按范围每天独立完成，详情没有周期规则并展示完成时间", async ({
  page,
  request,
}) => {
  const boot = await call(request, "Boot");
  await page.goto("/?preview=1");
  await add(page, "每天独立处理");
  await page.getByLabel("新待办标题", { exact: true }).press("Escape");
  const q = await call(request, "Query", { view: "day", date: boot.today });
  const t = q.tasks.find((x: any) => x.title === "每天独立处理");
  expect(t.seriesId).toBeTruthy();
  await row(page, t.title).hover();
  await row(page, t.title).getByLabel("调整任务时间", { exact: true }).click();
  await expect(
    page.getByRole("dialog", { name: "调整任务时间", exact: true }),
  ).toBeVisible();
  await page
    .getByRole("dialog", { name: "调整任务时间", exact: true })
    .getByRole("button", { name: "取消", exact: true })
    .click();
  await row(page, t.title).hover();
  await row(page, t.title).getByLabel("任务设置", { exact: true }).click();
  const editor = page.getByRole("dialog", { name: "任务详情", exact: true });
  await expect(editor.getByLabel("开始时间", { exact: true })).toBeVisible();
  await expect(editor.getByLabel("结束时间", { exact: true })).toBeVisible();
  await expect(editor.getByText("重复", { exact: true })).toHaveCount(0);
  await editor.getByRole("button", { name: "取消", exact: true }).click();
  await row(page, t.title)
    .getByLabel("完成任务及所有后代", { exact: true })
    .click();
  await expect(row(page, t.title).locator(".task-time")).toContainText("~");
  const d = new Date(boot.today + "T12:00:00");
  d.setDate(d.getDate() + 1);
  const tomorrow = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
  const next = await call(request, "Query", { view: "day", date: tomorrow });
  expect(next.tasks.find((x: any) => x.title === t.title).status).toBe(
    "pending",
  );
});
test("日历子任务开关与甘特图调整起止时间", async ({ page, request }) => {
  const t = await root(request, "甘特任务");
  await call(request, "Create", {
    commandId: crypto.randomUUID(),
    rootId: t.id,
    parentId: t.id,
    title: "日历中的子任务",
    expectedVersion: t.version,
  });
  await page.setViewportSize({ width: 1100, height: 760 });
  await page.goto("/?preview=1");
  await page.getByLabel("日历视图", { exact: true }).click();
  await expect(
    page.getByLabel("编辑任务：日历中的子任务", { exact: true }),
  ).toHaveCount(0);
  await page.getByRole("checkbox", { name: "子任务", exact: true }).check();
  await expect(
    page.getByLabel("编辑任务：日历中的子任务", { exact: true }),
  ).toBeVisible();
  await select(page, "日历范围", "甘特图");
  await expect(
    page.getByRole("region", { name: "甘特图", exact: true }),
  ).toBeVisible();
  const handle = page.getByLabel("调整结束时间 甘特任务", { exact: true });
  await handle.scrollIntoViewIfNeeded();
  const b = (await handle.boundingBox())!;
  await page.mouse.move(b.x + b.width / 2, b.y + b.height / 2);
  await page.mouse.down();
  await page.mouse.move(b.x + b.width / 2 - 60, b.y + b.height / 2, {
    steps: 10,
  });
  await page.mouse.up();
  await expect
    .poll(async () => {
      const r = await call(request, "FindTask", t.id);
      return r.endAt;
    })
    .not.toBeNull();
  const start = page.getByLabel("调整开始时间 甘特任务", { exact: true });
  await start.focus();
  await start.press("ArrowRight");
  await expect
    .poll(async () => {
      const r = await call(request, "FindTask", t.id);
      return r.startAt > t.startAt;
    })
    .toBe(true);
  await page.screenshot({ path: "../.impeccable/review/workspace-gantt.png" });
});
test("复盘图表、筛选、Markdown 笔记保存搜索和预览", async ({
  page,
  request,
}) => {
  const r = await root(request, "复盘任务");
  await call(
    request,
    "SetTaskTags",
    {
      commandId: crypto.randomUUID(),
      id: r.id,
      rootId: r.id,
      expectedVersion: r.version,
    },
    ["important"],
  );
  await page.setViewportSize({ width: 1100, height: 760 });
  await page.goto("/?preview=1");
  await page.getByLabel("复盘视图", { exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "完成热力图", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("img", { name: "最近两周完成任务柱状图" }),
  ).toBeVisible();
  await page.screenshot({
    path: "../.impeccable/review/workspace-review-home.png",
  });
  await page.getByRole("button", { name: "复盘", exact: true }).click();
  await expect(page.getByRole("button", { name: /复盘任务/ })).toBeVisible();
  await page.getByLabel("新建笔记", { exact: true }).click();
  await page.getByLabel("笔记标题", { exact: true }).fill("工作复盘");
  await page
    .getByLabel("Markdown 笔记内容", { exact: true })
    .fill("# 回顾\n\n- 完成一件事\n\n<script>window.injection=true</script>");
  await page.getByLabel("切换笔记预览", { exact: true }).click();
  await expect(page.locator(".note-preview h1")).toHaveText("回顾");
  expect(await page.evaluate(() => (window as any).injection)).toBeUndefined();
  await page.getByLabel("保存笔记", { exact: true }).click();
  await expect(
    page.getByLabel("复盘笔记", { exact: true }).getByText(/^已保存/),
  ).toBeVisible();
  await page.getByLabel("返回笔记列表", { exact: true }).click();
  await page.getByLabel("搜索笔记", { exact: true }).fill("工作复盘");
  await expect(
    page.getByRole("button", { name: /工作复盘 创建/ }),
  ).toBeVisible();
  await page.screenshot({
    path: "../.impeccable/review/workspace-review-notes.png",
  });
  await page.reload();
  await page.getByLabel("复盘视图", { exact: true }).click();
  await page.getByRole("button", { name: "复盘", exact: true }).click();
  await page.getByLabel("搜索笔记", { exact: true }).fill("工作复盘");
  await expect(
    page.getByRole("button", { name: /工作复盘 创建/ }),
  ).toBeVisible();
});
test("自动更新弹窗可跳过版本并持久化", async ({ page, request }) => {
  await page.route("**/api/AutomaticUpdate", async (route) => {
    const b = await call(request, "Boot");
    await route.fulfill({
      json:
        b.settings.skippedUpdateVersion === "99.0.0"
          ? null
          : {
              version: "99.0.0",
              available: true,
              releaseURL: "https://github.com/junyi-deep/SimpleTodo/releases",
            },
    });
  });
  await page.goto("/?preview=1");
  const prompt = page.getByRole("dialog", { name: "发现新版本", exact: true });
  await expect(prompt).toBeVisible();
  await expect(
    prompt.getByRole("button", { name: "立即更新", exact: true }),
  ).toBeVisible();
  await prompt.getByRole("button", { name: "跳过此版本", exact: true }).click();
  await expect(prompt).toHaveCount(0);
  expect((await call(request, "Boot")).settings.skippedUpdateVersion).toBe(
    "99.0.0",
  );
  await page.reload();
  await expect(prompt).toHaveCount(0);
});
test("今日边框跟随日期连续滑动，跨日自动进入新的一天", async ({
  page,
  request,
}) => {
  const boot = await call(request, "Boot");
  let today = boot.today;
  await page.route("**/api/Boot", (route) =>
    route.fulfill({ json: { ...boot, today } }),
  );
  await page.goto("/?preview=1");
  const selected = page.locator(".timeline-days .selected"),
    b = (await selected.boundingBox())!,
    outline = page.locator(".timeline-today-outline");
  let before = (await outline.boundingBox())!.x;
  await page.mouse.move(b.x + b.width / 2, b.y + b.height / 2);
  await page.mouse.down();
  for (let i = 1; i <= 9; i++) {
    await page.mouse.move(b.x + b.width / 2 + i * 5, b.y + b.height / 2);
    const x = (await outline.boundingBox())!.x;
    expect(Math.abs(x - before)).toBeLessThan(12);
    before = x;
  }
  await page.mouse.up();
  const d = new Date(today + "T12:00:00");
  d.setDate(d.getDate() + 1);
  today = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
  await expect(selected).toHaveAttribute("data-timeline-day", today, {
    timeout: 10000,
  });
});

test("笔记删除使用组件确认，取消及失败保留笔记，确认成功返回列表", async ({
  page,
  request,
}) => {
  const note = await call(request, "SaveNote", {
    id: "",
    title: "待删除笔记",
    content: "必须保留直到确认",
    version: 0,
  });
  await page.setViewportSize({ width: 1100, height: 760 });
  await page.goto("/?preview=1");
  await page.getByLabel("复盘视图", { exact: true }).click();
  await page.getByRole("button", { name: "复盘", exact: true }).click();
  await page.getByRole("button", { name: /待删除笔记 创建/ }).click();
  await page.screenshot({
    path: "../.impeccable/review/workspace-note-editor.png",
  });
  await page.getByLabel("删除笔记", { exact: true }).click();
  const confirm = page.getByRole("dialog", { name: "请确认", exact: true });
  await expect(confirm).toBeVisible();
  await page.screenshot({
    path: "../.impeccable/review/workspace-note-delete.png",
  });
  await confirm.getByRole("button", { name: "取消", exact: true }).click();
  expect(
    (await call(request, "Notes", "")).some((n: any) => n.id === note.id),
  ).toBe(true);
  await page.route("**/api/DeleteNote", (route) =>
    route.fulfill({
      status: 400,
      json: { error: "CONFLICT: 笔记已变化，请重新打开" },
    }),
  );
  await page.getByLabel("删除笔记", { exact: true }).click();
  await confirm.getByRole("button", { name: "确认", exact: true }).click();
  await expect(page.getByRole("alert")).toContainText("笔记已变化");
  await expect(page.getByLabel("笔记标题", { exact: true })).toHaveValue(
    note.title,
  );
  await page.unroute("**/api/DeleteNote");
  await page.getByLabel("删除笔记", { exact: true }).click();
  await confirm.getByRole("button", { name: "确认", exact: true }).click();
  await expect(page.getByLabel("搜索笔记", { exact: true })).toBeVisible();
  expect(
    (await call(request, "Notes", "")).some((n: any) => n.id === note.id),
  ).toBe(false);
});

test("独立设置等待启动数据，显示已保存的专注时长和窗口配置", async ({
  page,
}) => {
  await page.route("**/api/Boot", async (route) => {
    const response = await route.fetch();
    const boot = await response.json();
    boot.settings.focusMinutes = "37";
    boot.windows = [
      ...boot.windows.filter((w: any) => w.key !== "settings"),
      {
        key: "settings",
        x: 20,
        y: 30,
        width: 760,
        height: 720,
        mode: "top",
        opacity: 0.65,
        collapsed: false,
      },
    ];
    await new Promise((resolve) => setTimeout(resolve, 250));
    await route.fulfill({ response, json: boot });
  });
  await page.goto("/?preview=1&window=settings");
  await expect(
    page.getByRole("spinbutton", { name: "默认专注分钟数", exact: true }),
  ).toHaveValue("37");
  await page.getByRole("button", { name: "外观与窗口", exact: true }).click();
  await expect(
    page.getByRole("combobox", { name: "显示层级", exact: true }),
  ).toContainText("始终置顶");
  await expect(page.getByLabel("背景不透明度", { exact: true })).toHaveValue(
    "0.65",
  );
});
