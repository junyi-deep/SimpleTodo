import { test, expect } from "@playwright/test";
import { root, row, add, call, unique, select, tools } from "./helpers";
test("单击选中显示全文，双击在原条目编辑，单独图标打开详细设置", async ({
  page,
  request,
}) => {
  const t = await root(
    request,
    unique("这是需要完整阅读的清单任务".repeat(12)),
  );
  await page.goto("/?preview=1");
  const r = row(page, t.title);
  await expect(r.locator(".task-title")).toHaveCSS("white-space", "nowrap");
  await r.getByRole("button", { name: t.title, exact: true }).click();
  await expect(r).toHaveClass(/selected/);
  await expect(r.locator(".task-title")).toHaveCSS("white-space", "normal");
  await expect(page.locator("dialog")).toHaveCount(0);
  await r.getByRole("button", { name: t.title, exact: true }).dblclick();
  const input = page.getByLabel("编辑任务标题", { exact: true });
  await expect(input).toBeFocused();
  await page.screenshot({
    path: "../.impeccable/review/checklist-inline-title.png",
    animations: "disabled",
  });
  await input.fill("就地修改完成");
  await input.press("Enter");
  await expect(row(page, "就地修改完成")).toBeVisible();
  expect((await call(request, "FindTask", t.id)).title).toBe("就地修改完成");
  await row(page, "就地修改完成").getByLabel("任务设置").click();
  await expect(
    page.getByRole("dialog", { name: "任务详细设置" }),
  ).toBeVisible();
  await page.screenshot({
    path: "../.impeccable/review/checklist-settings-modal.png",
    animations: "disabled",
  });
});
test("加号草稿自动聚焦，空白离开与 Esc 取消，非空离开保存", async ({
  page,
  request,
}) => {
  await page.goto("/?preview=1");
  await expect(page.getByLabel("新待办标题")).toHaveCount(0);
  await page.getByLabel("新增待办", { exact: true }).click();
  const input = page.getByLabel("新待办标题");
  await expect(input).toBeFocused();
  await page.getByLabel("搜索当前视图").click();
  await expect(input).toHaveCount(0);
  await page.getByLabel("新增待办", { exact: true }).click();
  await input.fill("取消草稿");
  await input.press("Escape");
  await expect(input).toHaveCount(0);
  await expect(row(page, "取消草稿")).toHaveCount(0);
  const title = unique("失焦保存");
  await page.getByLabel("新增待办", { exact: true }).click();
  await input.fill(title);
  await page.getByLabel("搜索当前视图").click();
  await expect(row(page, title)).toBeVisible();
});
test("搜索位于最左，统计固定底部，选择右侧仅有三种批量操作", async ({
  page,
}) => {
  await page.goto("/?preview=1");
  await add(page, unique("批量任务"));
  const toolbar = page.locator(".list-toolbar");
  await expect(toolbar.getByLabel("切换已完成显示")).toHaveCount(0);
  expect(
    (await toolbar.getByLabel("搜索当前视图").boundingBox())!.x,
  ).toBeLessThan(
    (await toolbar
      .getByRole("button", { name: "选择", exact: true })
      .boundingBox())!.x,
  );
  await toolbar.getByRole("button", { name: "选择", exact: true }).click();
  await expect(toolbar.getByRole("button")).toHaveCount(4);
  await expect(toolbar.getByLabel("批量完成")).toBeDisabled();
  await page.locator(".root-select").first().check();
  await expect(toolbar.getByLabel("批量完成")).toBeEnabled();
  const footer = page.getByLabel("任务完成统计");
  const box = (await footer.boundingBox())!;
  expect(box.y + box.height).toBeCloseTo(760, 0);
  await page
    .locator(".task-list")
    .evaluate((el) => (el.scrollTop = el.scrollHeight));
  expect((await footer.boundingBox())!.y).toBe(box.y);
});
test("拖动到任务中央作为子任务，整个分支持久化", async ({ page, request }) => {
  const a = await root(request, unique("父任务")),
    b = await root(request, unique("将被嵌套"));
  await page.goto("/?preview=1");
  const r = row(page, a.title),
    box = (await r.boundingBox())!;
  const data = await page.evaluateHandle((id) => {
    const d = new DataTransfer();
    d.setData("text/plain", id);
    return d;
  }, b.id);
  await r.dispatchEvent("drop", {
    dataTransfer: data,
    clientY: box.y + box.height / 2,
    clientX: box.x + box.width / 2,
  });
  await expect(row(page, b.title)).toBeVisible();
  await expect
    .poll(async () => {
      const q = await call(request, "Query", { view: "day", date: a.date });
      return q.tasks.find((t: any) => t.id === b.id)?.parentId;
    })
    .toBe(a.id);
  await page.reload();
  await row(page, a.title).getByLabel("展开子任务").click();
  await expect(row(page, b.title)).toBeVisible();
});
test("提醒气泡可以创建每周重复及选择星期", async ({ page, request }) => {
  const t = await root(request, unique("重复提醒"));
  await call(request, "Settings", { notificationPrompted: "true" });
  await page.goto("/?preview=1");
  await row(page, t.title).hover();
  await row(page, t.title).getByLabel("设置提醒", { exact: true }).click();
  await select(page, "重复提醒", "每周");
  await page.getByLabel("重复间隔", { exact: true }).fill("2");
  await page.screenshot({
    path: "../.impeccable/review/checklist-repeat-reminder.png",
    animations: "disabled",
  });
  await page.getByRole("button", { name: "保存提醒", exact: true }).click();
  await expect(page.locator(".reminder-popover")).toHaveCount(0);
  const task = await call(request, "FindTask", t.id);
  expect(task.seriesId).toBeTruthy();
  const rules = await call(request, "Rules");
  expect(rules.find((r: any) => r.id === task.ruleId).interval).toBe(2);
  expect(rules.find((r: any) => r.id === task.ruleId).frequency).toBe("weekly");
});
test("关于直接读取 GitHub 最新发布并提供对应下载，无自定义发布地址", async ({
  page,
}) => {
  await page.route("**/api/CheckUpdate", (route) =>
    route.fulfill({
      json: {
        version: "9.0.0",
        available: true,
        downloadURL:
          "https://github.com/junyi-deep/SimpleTodo/releases/download/v9.0.0/SimpleTodo-macos-arm64.zip",
      },
    }),
  );
  await page.goto("/?preview=1");
  await tools(page, "帮助与关于");
  await expect(
    page.getByText("项目发布页（HTTPS）", { exact: true }),
  ).toHaveCount(0);
  await page.getByRole("button", { name: "检查更新", exact: true }).click();
  await expect(
    page.getByRole("button", { name: "下载 9.0.0", exact: true }),
  ).toBeVisible();
});
test("Windows 标题栏包含视图和窗口操作且可在窄窗口使用", async ({ page }) => {
  await page.route("**/api/Boot", async (route) => {
    const response = await route.fetch();
    const boot = await response.json();
    boot.capabilities.platform = "windows";
    await route.fulfill({ response, json: boot });
  });
  await page.setViewportSize({ width: 400, height: 720 });
  await page.goto("/?preview=1");
  const nav = page.getByRole("navigation", { name: "窗口与视图" });
  await expect(nav.getByLabel("关闭窗口", { exact: true })).toBeVisible();
  await expect(nav.getByLabel("任务列表", { exact: true })).toBeVisible();
  await nav.getByLabel("最小化窗口", { exact: true }).click();
  await nav.getByLabel("放大或还原窗口", { exact: true }).click();
  await nav.getByLabel("关闭窗口", { exact: true }).click();
  expect(
    await page.locator("body").evaluate((el) => el.scrollWidth <= innerWidth),
  ).toBe(true);
  await page.screenshot({
    path: "../.impeccable/review/checklist-windows-400.png",
    animations: "disabled",
  });
});
test("拖到子任务下缘放在该任务之后而非同级末尾", async ({ page, request }) => {
  const parent = await root(request, unique("同级排序"));
  const children: any[] = [];
  for (const title of ["A", "B", "C", "D"]) {
    const current = await call(request, "FindTask", parent.id);
    const tasks = await call(request, "Create", {
      commandId: crypto.randomUUID(),
      rootId: parent.id,
      parentId: parent.id,
      expectedVersion: current.version,
      title,
    });
    children.push(tasks.find((t: any) => t.title === title));
  }
  await page.goto("/?preview=1");
  await row(page, parent.title).getByLabel("展开子任务").click();
  const target = row(page, "B"),
    box = (await target.boundingBox())!;
  const data = await page.evaluateHandle((id) => {
    const d = new DataTransfer();
    d.setData("text/plain", id);
    return d;
  }, children[0].id);
  await target.dispatchEvent("drop", {
    dataTransfer: data,
    clientY: box.y + box.height - 1,
    clientX: box.x + box.width / 2,
  });
  await expect
    .poll(async () => {
      const q = await call(request, "Query", {
        view: "day",
        date: parent.date,
      });
      return q.tasks
        .filter((t: any) => t.parentId === parent.id)
        .sort((a: any, b: any) => a.order - b.order)
        .map((t: any) => t.title)
        .join(",");
    })
    .toBe("B,A,C,D");
  await row(page, "D").scrollIntoViewIfNeeded();
  await page.screenshot({
    path: "../.impeccable/review/checklist-sibling-order.png",
    animations: "disabled",
  });
});
