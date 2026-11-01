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
    name: "任务详情",
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
test("时间弹窗保存起止范围，结束日之后不生成记录", async ({
  page,
  request,
}) => {
  const t = await root(request, "时间范围");
  await page.goto("/?preview=1");
  await row(page, t.title).hover();
  await row(page, t.title).getByLabel("调整任务时间", { exact: true }).click();
  const p = page.getByRole("dialog", { name: "调整任务时间", exact: true });
  await p.getByLabel("开始时间", { exact: true }).fill(t.date + "T00:00");
  await p.getByLabel("结束时间", { exact: true }).fill(t.date + "T23:59");
  await p.getByRole("button", { name: "确认调整", exact: true }).click();
  await expect(p).toHaveCount(0);
  const saved = await call(request, "FindTask", t.id);
  expect(saved.endAt).toBeGreaterThan(saved.startAt);
  const d = new Date(t.date + "T12:00:00");
  d.setDate(d.getDate() + 1);
  const next = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
  const q = await call(request, "Query", { view: "day", date: next });
  expect(q.tasks.some((n: any) => n.title === t.title)).toBe(false);
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
  const e = page.getByRole("dialog", { name: "任务详情", exact: true });
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

test("浅深主题下完成框与标签选择的键盘焦点可见", async ({ page, request }) => {
  for (const theme of ["light", "dark"]) {
    await call(request, "Settings", { theme, animation: "false" });
    const t = await root(request, theme + "焦点"),
      done = await root(request, theme + "已完成");
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
    expect(
      await check.evaluate((el) => getComputedStyle(el).boxShadow),
    ).not.toBe("none");
    await row(page, t.title).hover();
    await row(page, t.title)
      .getByLabel("设置标签 " + t.title, { exact: true })
      .click();
    const pop = page.getByRole("dialog", { name: "选择任务标签", exact: true });
    for (const name of ["重要", "紧急"]) {
      const box = pop.getByRole("checkbox", { name, exact: true });
      await box.focus();
      await page.keyboard.press("Tab");
      await page.keyboard.press("Shift+Tab");
      await expect(box).toBeFocused();
      await expect(box).toHaveCSS("outline-style", "none");
    }
    await page.keyboard.press("Escape");
  }
});

test("清空结束时间后恢复持续的每日任务", async ({ page, request }) => {
  const t = await root(request, "清空结束时间");
  await page.goto("/?preview=1");
  await row(page, t.title).hover();
  await row(page, t.title).getByLabel("调整任务时间", { exact: true }).click();
  const d = page.getByRole("dialog", { name: "调整任务时间", exact: true });
  await d.getByLabel("结束时间", { exact: true }).fill(t.date + "T23:59");
  await d.getByLabel("结束时间", { exact: true }).fill("");
  const next = new Date(t.date + "T12:00:00");
  next.setDate(next.getDate() + 1);
  const day = `${next.getFullYear()}-${String(next.getMonth() + 1).padStart(2, "0")}-${String(next.getDate()).padStart(2, "0")}`;
  await d.getByLabel("开始时间", { exact: true }).fill(day + "T09:00");
  await d.getByRole("button", { name: "确认调整", exact: true }).click();
  await expect(d).toHaveCount(0);
  const saved = await call(request, "FindTask", t.id);
  expect(saved.endAt).toBeNull();
  const q = await call(request, "Query", { view: "day", date: day });
  expect(q.tasks.some((n: any) => n.title === t.title)).toBe(true);
});
