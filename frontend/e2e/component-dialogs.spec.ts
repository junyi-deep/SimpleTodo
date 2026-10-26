import { test, expect } from "@playwright/test";
import { call, root, row, select } from "./helpers";
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
    animation: "false",
    theme: "light",
    completedBottom: "false",
  });
});
test("子任务使用普通草稿，空白失焦取消、非空失焦保存，Enter 连续同级", async ({
  page,
  request,
}) => {
  const t = await root(request, "统一录入");
  await page.goto("/?preview=1");
  await row(page, t.title).focus();
  await page.keyboard.press("Shift+Enter");
  const input = page.getByLabel("新待办标题", { exact: true });
  await expect(input).toBeFocused();
  await expect(page.getByPlaceholder("添加子任务，Enter 连续录入")).toHaveCount(
    0,
  );
  await expect(page.getByLabel("取消新增", { exact: true })).toHaveCount(0);
  await page.getByLabel("搜索当前视图").click();
  await expect(input).toHaveCount(0);
  expect(await call(request, "ReadTree", t.id)).toHaveLength(1);
  await row(page, t.title).focus();
  await page.keyboard.press("Shift+Enter");
  await input.fill("子任务失焦保存");
  await page.getByLabel("搜索当前视图").click();
  await expect(input).toHaveCount(0);
  await expect(
    page.getByRole("button", { name: "子任务失焦保存", exact: true }),
  ).toBeVisible();
  await row(page, "子任务失焦保存")
    .getByRole("button", { name: "子任务失焦保存", exact: true })
    .click();
  await page.getByLabel("编辑任务标题", { exact: true }).press("Enter");
  await input.fill("同层下一个");
  await input.press("Enter");
  await expect(input).toHaveValue("");
  await expect(input).toBeFocused();
  await input.press("Escape");
  await expect(input).toHaveCount(0);
  const tree = await call(request, "ReadTree", t.id);
  expect(tree.filter((n: any) => n.parentId === t.id)).toHaveLength(2);
  await page.screenshot({
    path: "../.impeccable/review/component-child-draft.png",
  });
});
test("完成任务不能拖动，未完成任务仍可预览和排序", async ({
  page,
  request,
}) => {
  const a = await root(request, "可以拖动"),
    b = await root(request, "另一个待办"),
    c = await root(request, "已完成不可拖动");
  await call(request, "Complete", {
    commandId: crypto.randomUUID(),
    id: c.id,
    rootId: c.id,
    expectedVersion: c.version,
    completed: true,
  });
  await page.goto("/?preview=1");
  const done = page.locator(`[data-node="${c.id}"] .drag-handle`);
  await expect(done).toHaveAttribute("draggable", "false");
  await expect(done).toHaveCSS("pointer-events", "none");
  const src = page.locator(`[data-node="${a.id}"] .drag-handle`),
    dst = page.locator(`[data-node="${b.id}"]`);
  await expect(src).toHaveAttribute("data-drag-enabled", "true");
  const x = (await src.boundingBox())!,
    y = (await dst.boundingBox())!;
  await page.mouse.move(x.x + 5, x.y + 5);
  await page.mouse.down();
  await page.mouse.move(y.x + 20, y.y + y.height - 2, { steps: 12 });
  await expect(page.locator(".pointer-drag-preview")).toHaveCount(1);
  await page.mouse.up();
  await expect
    .poll(async () => {
      const q = await call(request, "Query", { view: "all" });
      return (
        q.tasks.find((t: any) => t.id === a.id).order >
        q.tasks.find((t: any) => t.id === b.id).order
      );
    })
    .toBe(true);
});
test("应用确认使用组件，取消保留描述，确认放弃，不出现黑色焦点框或时区标签", async ({
  page,
  request,
}) => {
  const t = await root(request, "描述和确认");
  await page.addInitScript(() => {
    window.confirm = () => {
      throw new Error("禁止系统确认框");
    };
  });
  await page.goto("/?preview=1");
  await row(page, t.title).hover();
  await row(page, t.title).getByLabel("任务设置", { exact: true }).click();
  const editor = page.getByRole("dialog", {
    name: "任务详细设置",
    exact: true,
  });
  const desc = editor.getByLabel("描述", { exact: true });
  await desc.fill("多行描述\n继续输入");
  await expect(desc).toHaveCSS("outline-style", "none");
  await expect(page.getByText(/系统时区/)).toHaveCount(0);
  await page.keyboard.press("Escape");
  const confirm = page.locator(".component-confirmation");
  await expect(confirm).toBeVisible();
  await expect(
    confirm.getByRole("button", { name: "取消", exact: true }),
  ).toBeFocused();
  await page.screenshot({
    path: "../.impeccable/review/component-confirmation.png",
  });
  await page.keyboard.press("Enter");
  await expect(confirm).toHaveCount(0);
  await expect(desc).toHaveValue("多行描述\n继续输入");
  await editor.getByRole("button", { name: "取消", exact: true }).click();
  await confirm.getByRole("button", { name: "确认", exact: true }).click();
  await expect(editor).toHaveCount(0);
  expect((await call(request, "FindTask", t.id)).description).toBe("");
});
test("改期弹窗内设置重复，日期和规则一起保存且可跨周生成", async ({
  page,
  request,
}) => {
  const t = await root(request, "改期并重复");
  await page.goto("/?preview=1");
  await row(page, t.title).hover();
  await row(page, t.title).getByLabel("调整计划日期", { exact: true }).click();
  const p = page.getByRole("dialog", { name: "调整计划日期", exact: true });
  await select(page, "计划重复方式", "每日");
  await p.getByLabel("重复间隔", { exact: true }).fill("2");
  await expect(
    p.getByText("从计划日期开始，结束日期留空则一直重复。", { exact: true }),
  ).toBeVisible();
  await page.screenshot({
    path: "../.impeccable/review/component-placement-repeat.png",
  });
  await p.getByRole("button", { name: "确认移动", exact: true }).click();
  await expect(p).toHaveCount(0);
  const r = (await call(request, "Rules"))[0];
  expect(r.frequency).toBe("daily");
  expect(r.interval).toBe(2);
  expect(r.until).toBeNull();
  await page.reload();
  await expect(
    page.locator(".task-repeat").filter({ hasText: "每 2 天 · 持续重复" }),
  ).toBeVisible();
});

test("编辑中的未完成任务拖动时先保存标题再排序", async ({ page, request }) => {
  const a = await root(request, "边编辑边排序"),
    b = await root(request, "排序目标");
  await page.goto("/?preview=1");
  const r = page.locator(`[data-node="${a.id}"]`);
  await r.getByRole("button", { name: a.title, exact: true }).click();
  await page
    .getByLabel("编辑任务标题", { exact: true })
    .fill("拖动前保留我的修改");
  const h = (await r.locator(".drag-handle").boundingBox())!,
    d = (await page.locator(`[data-node="${b.id}"]`).boundingBox())!;
  await page.mouse.move(h.x + 5, h.y + 5);
  await page.mouse.down();
  await page.mouse.move(d.x + 20, d.y + d.height - 2, { steps: 12 });
  await page.mouse.up();
  await expect
    .poll(async () => {
      const q = await call(request, "Query", { view: "all" }),
        x = q.tasks.find((t: any) => t.id === a.id),
        y = q.tasks.find((t: any) => t.id === b.id);
      return x.title === "拖动前保留我的修改" && x.order > y.order;
    })
    .toBe(true);
});

test("深色窄窗口确认框、输入焦点和 Esc 取消可用", async ({ page, request }) => {
  await call(request, "Settings", { theme: "dark", animation: "true" });
  const t = await root(request, "深色确认");
  await page.setViewportSize({ width: 400, height: 720 });
  await page.goto("/?preview=1");
  await row(page, t.title).hover();
  await row(page, t.title).getByLabel("任务设置", { exact: true }).click();
  const e = page.getByRole("dialog", { name: "任务详细设置", exact: true });
  await e.getByLabel("描述", { exact: true }).fill("深色输入没有外框");
  await expect(e.getByLabel("描述", { exact: true })).toHaveCSS(
    "outline-style",
    "none",
  );
  await page.keyboard.press("Escape");
  const c = page.locator(".component-confirmation");
  await expect(c).toBeVisible();
  await expect(
    c.getByRole("button", { name: "确认", exact: true }),
  ).toBeInViewport();
  await page.screenshot({
    path: "../.impeccable/review/component-confirmation-dark-400.png",
    animations: "disabled",
  });
  await page.keyboard.press("Escape");
  await expect(c).toHaveCount(0);
  await expect(e.getByLabel("描述", { exact: true })).toHaveValue(
    "深色输入没有外框",
  );
});

test("浅深主题下已选和未选星期、已完成复选框都有细内描边焦点", async ({
  page,
  request,
}) => {
  for (const theme of ["light", "dark"]) {
    await call(request, "Settings", { theme, animation: "false" });
    const t = await root(request, theme + "焦点");
    const done = await root(request, theme + "已完成");
    await call(request, "Complete", {
      commandId: crypto.randomUUID(),
      id: done.id,
      rootId: done.id,
      expectedVersion: done.version,
      completed: true,
    });
    await page.goto("/?preview=1");
    const check = page.locator(`[data-node="${done.id}"] .task-check`);
    await check.focus();
    await page.keyboard.press("Tab");
    await page.keyboard.press("Shift+Tab");
    await expect(check).toBeFocused();
    await expect(check).toHaveCSS("outline-style", "none");
    expect(
      await check.evaluate((el) => getComputedStyle(el).boxShadow),
    ).not.toBe("none");
    await row(page, t.title).hover();
    await row(page, t.title)
      .getByLabel("调整计划日期", { exact: true })
      .click();
    await select(page, "计划重复方式", "每周");
    const picker = page.getByRole("group", { name: "重复星期", exact: true });
    for (const pressed of ["true", "false"]) {
      const b = picker.locator(`[aria-pressed="${pressed}"]`).first();
      const bg = await b.evaluate((el) => getComputedStyle(el).backgroundColor);
      await b.focus();
      await page.keyboard.press("Tab");
      await page.keyboard.press("Shift+Tab");
      await expect(b).toBeFocused();
      await expect(b).toHaveCSS("outline-style", "none");
      await expect(b).toHaveCSS("background-color", bg);
      expect(await b.evaluate((el) => getComputedStyle(el).boxShadow)).not.toBe(
        "none",
      );
    }
    await page.screenshot({
      path: `../.impeccable/review/component-weekday-focus-${theme}.png`,
    });
  }
});

test("关闭重复后改期不受隐藏的旧结束日期限制", async ({ page, request }) => {
  const t = await root(request, "取消重复设置");
  await page.goto("/?preview=1");
  await row(page, t.title).hover();
  await row(page, t.title).getByLabel("调整计划日期", { exact: true }).click();
  const d = page.getByRole("dialog", { name: "调整计划日期", exact: true });
  await select(page, "计划重复方式", "每日");
  await d.getByLabel("结束日期（可选）", { exact: true }).fill(t.date);
  await select(page, "计划重复方式", "不重复");
  const date = new Date(t.date + "T12:00:00");
  date.setDate(date.getDate() + 1);
  const next = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
  await d.getByLabel("计划日期", { exact: true }).fill(next);
  await d.getByRole("button", { name: "确认移动", exact: true }).click();
  await expect(d).toHaveCount(0);
  const saved = await call(request, "FindTask", t.id);
  expect(saved.date).toBe(next);
  expect(saved.seriesId).toBeNull();
});
