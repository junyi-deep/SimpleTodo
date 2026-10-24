import { test, expect } from "@playwright/test";
import { unique, call, root, row, add, tools, select } from "./helpers";
test("列表内连续录入、多级快捷键和完成汇总", async ({ page }) => {
  await page.goto("/?preview=1");
  const title = unique("准备发布"),
    child = unique("整理材料"),
    leaf = unique("核对清单");
  await add(page, title);
  await row(page, title).focus();
  await page.keyboard.press("Shift+Enter");
  let input = page.getByLabel("新待办标题", { exact: true });
  await input.fill(child);
  await input.press("Enter");
  await expect(input).toHaveValue("");
  await input.press("Escape");
  await row(page, child).focus();
  await page.keyboard.press("Shift+Enter");
  input = page.getByLabel("新待办标题", { exact: true });
  await input.fill(leaf);
  await input.press("Enter");
  await expect(input).toHaveValue("");
  await input.press("Escape");
  await row(page, leaf).getByLabel("完成任务及所有后代").click();
  await expect(
    row(page, title).getByRole("button", { name: "取消完成", exact: true }),
  ).toHaveAttribute("aria-pressed", "true");
  await page.reload();
  await expect(
    row(page, title).getByRole("button", { name: "取消完成", exact: true }),
  ).toHaveAttribute("aria-pressed", "true");
});
test("收集箱显示全部任务，详情弹窗保存，无颜色入口", async ({
  page,
  request,
}) => {
  const scheduled = await root(request, unique("已安排")),
    inbox = await root(request, unique("未安排"), null);
  await page.goto("/?preview=1");
  await page.getByRole("button", { name: "日计划", exact: true }).click();
  await expect(
    page.getByRole("button", { name: scheduled.title, exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: inbox.title, exact: true }),
  ).toBeVisible();
  await row(page, inbox.title).hover();
  await row(page, inbox.title).getByLabel("任务设置").click();
  const editor = page.locator(".task-settings-dialog");
  await expect(editor).toBeVisible();
  await expect(page.locator("dialog")).toHaveCount(1);
  await expect(editor.getByText("颜色", { exact: true })).toHaveCount(0);
  await editor.getByLabel("任务标题").fill(inbox.title + "更新");
  await editor
    .getByRole("textbox", { name: "描述", exact: true })
    .fill("可在列表中直接编辑");
  await page.screenshot({
    path: "../.impeccable/review/redesign-inline-editor.png",
    animations: "disabled",
  });
  await editor.getByRole("button", { name: "保存修改" }).click();
  await expect(editor).toHaveCount(0);
  await expect(
    page.getByRole("button", { name: inbox.title + "更新", exact: true }),
  ).toBeVisible();
});
test("自定义下拉支持键盘，主题持久化，页面没有原生 select", async ({
  page,
}) => {
  await page.goto("/?preview=1");
  await tools(page, "设置");
  await page.getByRole("button", { name: "外观与窗口", exact: true }).click();
  const theme = page.getByRole("combobox", { name: "主题", exact: true });
  await theme.focus();
  await theme.press("Enter");
  await page.screenshot({
    path: "../.impeccable/review/redesign-custom-select.png",
    animations: "disabled",
  });
  await page.getByRole("option", { name: "深色", exact: true }).click();
  await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
  await page.reload();
  await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
  await tools(page, "设置");
  await page.getByRole("button", { name: "外观与窗口", exact: true }).click();
  await select(page, "主题", "浅色");
  await expect(page.locator("select")).toHaveCount(0);
  await expect(page.locator("dialog")).toHaveCount(0);
});
test("月周日历共用页面，宽抽屉可用 Esc 和外部点击关闭", async ({
  page,
  request,
}) => {
  const task = await root(request, unique("日历任务"));
  let popups = 0;
  page.on("popup", () => popups++);
  await page.goto("/?preview=1");
  await page.setViewportSize({ width: 1080, height: 720 });
  await page.getByLabel("日历视图", { exact: true }).click();
  await expect(page.locator(".calendar-day")).toHaveCount(42);
  await page.getByRole("button", { name: task.date, exact: true }).click();
  const drawer = page.getByLabel("日期任务抽屉");
  await expect(drawer).toBeVisible();
  expect((await drawer.boundingBox())!.width).toBeGreaterThanOrEqual(360);
  await expect(
    drawer.getByRole("button", { name: task.title, exact: true }),
  ).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(drawer).toHaveCount(0);
  await page.getByRole("button", { name: task.date, exact: true }).click();
  await page.locator(".calendar-month-picker").click();
  await expect(drawer).toHaveCount(0);
  await select(page, "日历范围", "周视图");
  await expect(page.locator(".calendar-day")).toHaveCount(7);
  await page.screenshot({
    path: "../.impeccable/review/redesign-week.png",
    animations: "disabled",
  });
  expect(popups).toBe(0);
});
test("无更多菜单、标题省略完整提示、提醒气泡保存和移除", async ({
  page,
  request,
}) => {
  const task = await root(request, unique("这是很长的任务标题".repeat(8)));
  await call(request, "Settings", { notificationPrompted: "true" });
  await page.goto("/?preview=1");
  const target = row(page, task.title);
  await expect(target.locator(".task-title")).toHaveAttribute(
    "title",
    task.title,
  );
  await expect(target.locator(".task-title")).toHaveCSS(
    "white-space",
    "nowrap",
  );
  await target.hover();
  await expect(target.getByLabel("更多任务操作")).toHaveCount(0);
  await target.getByLabel("设置提醒", { exact: true }).click();
  const popover = page.locator(".reminder-popover");
  await expect(popover).toBeVisible();
  await page.screenshot({
    path: "../.impeccable/review/redesign-reminder.png",
    animations: "disabled",
  });
  const d = new Date(Date.now() + 3600000),
    value = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}T${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
  await popover.getByLabel("提醒时间").fill(value);
  await popover.getByRole("button", { name: "保存提醒", exact: true }).click();
  await expect(target.locator(".task-reminder")).toBeVisible();
  await target.hover();
  await target.getByLabel("设置提醒", { exact: true }).click();
  await popover.getByRole("button", { name: "移除提醒", exact: true }).click();
  await expect(target.locator(".task-reminder")).toHaveCount(0);
});
test("多级 Markdown 导入使用自定义格式选择并实际提交", async ({ page }) => {
  await page.goto("/?preview=1");
  await tools(page, "导入待办");
  await select(page, "导入格式", "Markdown 多级清单");
  const title = unique("导入根");
  await page
    .locator("textarea")
    .fill(`- [ ] ${title}\n  - [ ] 一层\n    - [ ] 二层`);
  await page.getByRole("button", { name: "预览导入" }).click();
  await expect(page.getByText("1 个根任务 · 3 个节点")).toBeVisible();
  await page.getByRole("button", { name: "确认执行" }).click();
  await expect(
    page.getByRole("button", { name: title, exact: true }),
  ).toBeVisible();
});
test("专注计时和设置备份仍可使用", async ({ page }) => {
  await page.goto("/?preview=1");
  await page.getByRole("button", { name: "专注与倒计时", exact: true }).click();
  await page.getByRole("button", { name: "开始专注", exact: true }).click();
  await page.getByRole("button", { name: "暂停", exact: true }).click();
  await expect(
    page.getByRole("button", { name: "继续", exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "结束", exact: true }).click();
  await page
    .locator(".component-confirmation")
    .getByRole("button", { name: "确认", exact: true })
    .click();
  await tools(page, "设置");
  await page.getByRole("button", { name: "本地数据", exact: true }).click();
  await page.getByRole("button", { name: "立即备份", exact: true }).click();
  await expect(page.locator(".backup-list>div").first()).toBeVisible();
});
