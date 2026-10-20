import { test, expect } from "@playwright/test";
import { unique, call, root, row, add, select } from "./helpers";

test("行内编辑仍固定版本并支持冲突后显式重试", async ({ page, request }) => {
  const t = await root(request, unique("冲突草稿"));
  await page.goto("/?preview=1");
  await page.getByRole("button", { name: t.title, exact: true }).click();
  const editor = page.locator(".inline-task-editor");
  await editor.getByLabel("任务标题").fill(t.title + "草稿");
  const current = await call(request, "FindTask", t.id);
  await call(request, "Update", {
    commandId: crypto.randomUUID(),
    id: t.id,
    rootId: t.rootId,
    expectedVersion: current.version,
    title: t.title,
    description: "另一个操作已更新",
    date: t.date,
    color: "none",
    postpone: "off",
  });
  await editor.getByRole("button", { name: "保存修改", exact: true }).click();
  await expect(editor.getByRole("alert")).toContainText("CONFLICT");
  await editor
    .getByRole("button", { name: "查看最新版本", exact: true })
    .click();
  page.once("dialog", (d) => d.accept());
  await editor
    .getByRole("button", { name: "基于最新版本重试", exact: true })
    .click();
  await expect(
    page.getByRole("button", { name: t.title + "草稿", exact: true }),
  ).toBeVisible();
});
test("重复模板使用行内编辑器和自定义周期下拉", async ({ page, request }) => {
  const t = await root(request, unique("每天复盘"));
  await page.goto("/?preview=1");
  await page.getByRole("button", { name: t.title, exact: true }).click();
  const editor = page.locator(".inline-task-editor");
  await select(page, "不重复", "每日");
  await editor.getByRole("button", { name: "保存修改", exact: true }).click();
  const updated = await call(request, "FindTask", t.id);
  expect(updated.seriesId).toBeTruthy();
  const rules = await call(request, "Rules");
  expect(rules.find((r: any) => r.id === updated.ruleId).frequency).toBe(
    "daily",
  );
});
test("拖动显示预览和落点并保存根任务顺序", async ({ page, request }) => {
  const a = await root(request, unique("拖动A")),
    b = await root(request, unique("拖动B"));
  await page.goto("/?preview=1");
  await row(page, b.title).hover();
  const data = await page.evaluateHandle(() => new DataTransfer());
  await row(page, b.title)
    .locator(".drag-handle")
    .dispatchEvent("dragstart", { dataTransfer: data });
  await row(page, a.title).dispatchEvent("dragenter", { dataTransfer: data });
  await expect(row(page, a.title)).toHaveClass(/drop-target/);
  await row(page, a.title).dispatchEvent("drop", { dataTransfer: data });
  await expect
    .poll(async () => {
      const q = await call(request, "Query", { view: "day", date: a.date });
      const ar = q.tasks.find((t: any) => t.id === a.id),
        br = q.tasks.find((t: any) => t.id === b.id);
      return br.order < ar.order;
    })
    .toBe(true);
});
test("Esc 仅关闭菜单或子任务输入，取消后恢复父任务焦点", async ({ page }) => {
  await page.goto("/?preview=1");
  const t = unique("取消录入");
  await add(page, t);
  const r = row(page, t),
    trigger = r.getByLabel("更多任务操作");
  await r.hover();
  await trigger.click();
  await trigger.press("Escape");
  await expect(page.locator(".node-menu")).toHaveCount(0);
  await r.focus();
  await r.press("Shift+Enter");
  const input = page.getByPlaceholder("添加子任务，Enter 连续录入");
  await input.fill("不保存");
  await input.press("Escape");
  await expect(input).toHaveCount(0);
  await expect(r).toBeFocused();
  await page.evaluate(() =>
    window.dispatchEvent(
      new CustomEvent("todo:confirm-error", {
        detail: "无法显示确认窗口，请重试：验收错误",
      }),
    ),
  );
  await expect(page.getByRole("alert")).toContainText("无法显示确认窗口");
});
test("输入法确认不误提交，多行粘贴仍先预览", async ({ page }) => {
  await page.goto("/?preview=1");
  const input = page.getByLabel("新待办标题"),
    t = unique("中文输入");
  await input.fill(t);
  await input.dispatchEvent("compositionstart");
  await input.press("Enter");
  await expect(page.getByRole("button", { name: t, exact: true })).toHaveCount(
    0,
  );
  await input.dispatchEvent("compositionend");
  await input.press("Enter");
  await expect(
    page.getByRole("button", { name: t, exact: true }),
  ).toBeVisible();
  await input.evaluate((el) => {
    const dt = new DataTransfer();
    dt.setData("text/plain", "多行导入一\n多行导入二");
    el.dispatchEvent(
      new ClipboardEvent("paste", {
        bubbles: true,
        cancelable: true,
        clipboardData: dt,
      }),
    );
  });
  await expect(page.locator(".transfer-dialog")).toBeVisible();
  await expect(page.locator("textarea")).toHaveValue("多行导入一\n多行导入二");
});

test("抽屉草稿在点外部、换日及切视图取消时保留，确认放弃后继续", async ({
  page,
  request,
}) => {
  const t = await root(request, unique("抽屉草稿"));
  await page.setViewportSize({ width: 1080, height: 720 });
  await page.goto("/?preview=1");
  await page.getByLabel("日历视图", { exact: true }).click();
  await page.getByRole("button", { name: t.date, exact: true }).click();
  const drawer = page.getByLabel("日期任务抽屉");
  await drawer.getByRole("button", { name: t.title, exact: true }).click();
  const editor = drawer.locator(".inline-task-editor");
  await editor.getByLabel("任务标题").fill(t.title + "未保存");
  await editor
    .getByRole("textbox", { name: "描述", exact: true })
    .fill("请保留这段草稿");
  page.once("dialog", (d) => d.dismiss());
  await page.locator(".calendar-month-picker").click();
  await expect(drawer).toBeVisible();
  await expect(editor.getByLabel("任务标题")).toHaveValue(t.title + "未保存");
  const next = new Date(t.date + "T12:00:00");
  next.setDate(next.getDate() + 1);
  const nextDay = `${next.getFullYear()}-${String(next.getMonth() + 1).padStart(2, "0")}-${String(next.getDate()).padStart(2, "0")}`;
  page.once("dialog", (d) => d.dismiss());
  await page.getByRole("button", { name: nextDay, exact: true }).click();
  await expect(editor).toBeVisible();
  await expect(
    editor.getByRole("textbox", { name: "描述", exact: true }),
  ).toHaveValue("请保留这段草稿");
  page.once("dialog", (d) => d.dismiss());
  await page.getByLabel("任务列表", { exact: true }).click();
  await expect(drawer).toBeVisible();
  await expect(page.locator(".calendar-grid")).toBeVisible();
  page.once("dialog", (d) => d.accept());
  await page.getByRole("button", { name: nextDay, exact: true }).click();
  await expect(editor).toHaveCount(0);
  await expect(drawer.getByRole("heading")).not.toContainText("未保存");
  const stored = await call(request, "FindTask", t.id);
  expect(stored.title).toBe(t.title);
  expect(stored.description || "").not.toContain("请保留这段草稿");
});

test("日历复制仍以选区首日为相对日期锚点", async ({
  page,
  request,
  context,
}) => {
  await context.grantPermissions(["clipboard-read", "clipboard-write"]);
  const t = await root(request, unique("范围复制"));
  const date = new Date(t.date + "T12:00:00");
  date.setDate(date.getDate() - 1);
  const anchor = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
  await page.setViewportSize({ width: 1080, height: 720 });
  await page.goto("/?preview=1");
  await page.getByLabel("日历视图", { exact: true }).click();
  await page.getByRole("button", { name: anchor, exact: true }).click();
  await page
    .getByRole("button", { name: t.date, exact: true })
    .click({ modifiers: ["Shift"] });
  const sent = page.waitForRequest((r) => r.url().endsWith("/api/CopyTasks"));
  await page.getByLabel("复制任务", { exact: true }).click();
  expect((await sent).postDataJSON()[0].date).toBe(anchor);
});
