import { test, expect } from "@playwright/test";
import { tools, select, root, row, call, unique } from "./helpers";

test("标题栏顺序、关于文案及代理设置持久化", async ({ page, request }) => {
  await call(request, "SaveUpdateProxy", {
    mode: "system",
    address: "",
    username: "",
    password: "",
  });
  await page.goto("/?preview=1");
  const nav = page.getByRole("navigation", { name: "窗口与视图" });
  const labels = await nav
    .locator("button,summary")
    .evaluateAll((nodes) => nodes.map((n) => n.getAttribute("aria-label")));
  expect(labels.slice(-5)).toEqual([
    "固定窗口",
    "工具菜单",
    "最小化窗口",
    "放大或还原窗口",
    "关闭窗口",
  ]);
  await tools(page, "帮助与关于");
  await expect(
    page.getByText("所有任务数据仅存放在本地电脑上。", { exact: false }),
  ).toBeVisible();
  await expect(page.getByText("使用提示", { exact: true })).toHaveCount(0);
  await expect(page.getByText("由 Go、Wails", { exact: false })).toHaveCount(0);
  await expect(page.getByRole("combobox", { name: "更新代理" })).toHaveText(
    "系统代理",
  );
  await select(page, "更新代理", "自定义代理");
  await page
    .getByLabel("HTTP 代理地址", { exact: true })
    .fill("socks5://127.0.0.1:7890");
  await page.getByRole("button", { name: "保存代理设置" }).click();
  await expect(page.getByRole("alert")).toContainText("HTTP 代理");
  await page
    .getByLabel("HTTP 代理地址", { exact: true })
    .fill("http://127.0.0.1:7890");
  await page.getByLabel("代理用户名", { exact: true }).fill("demo");
  await page.getByLabel("代理密码", { exact: true }).fill("test-only-secret");
  await page.getByRole("button", { name: "保存代理设置" }).click();
  await expect(page.getByRole("status")).toContainText("更新代理已保存");
  const config = await call(request, "UpdateProxy");
  expect(config.password).toBe("");
  expect(config.hasPassword).toBe(true);
  await page.reload();
  await tools(page, "帮助与关于");
  await expect(page.getByLabel("HTTP 代理地址", { exact: true })).toHaveValue(
    "http://127.0.0.1:7890",
  );
  await expect(page.getByLabel("代理密码", { exact: true })).toHaveValue("");
  await page.screenshot({
    path: "../.impeccable/review/motion-proxy-about.png",
    animations: "disabled",
  });
  await page.getByRole("button", { name: "清除已保存的密码" }).click();
  await page.getByRole("button", { name: "保存代理设置" }).click();
  await expect
    .poll(async () => (await call(request, "UpdateProxy")).hasPassword || false)
    .toBe(false);
  await select(page, "更新代理", "无代理");
  await page.getByRole("button", { name: "保存代理设置" }).click();
  await expect
    .poll(async () => (await call(request, "UpdateProxy")).mode)
    .toBe("none");
  await call(request, "SaveUpdateProxy", {
    mode: "system",
    address: "",
    username: "",
    password: "",
  });
});
test("Motion 完成反馈、快速切换及减少动态效果", async ({ page, request }) => {
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  await call(request, "Settings", { animation: "true" });
  const t = await root(request, unique("完成反馈"));
  await page.goto("/?preview=1");
  await row(page, t.title)
    .getByLabel("完成任务及所有后代", { exact: true })
    .click();
  await expect(
    row(page, t.title).getByRole("button", { name: "取消完成", exact: true }),
  ).toHaveAttribute("aria-pressed", "true");
  await expect
    .poll(() =>
      page
        .locator(".task-check")
        .evaluateAll((nodes) => nodes.flatMap((n) => n.getAnimations()).length),
    )
    .toBe(0);
  await page.getByLabel("日历视图", { exact: true }).click();
  await page.locator(".calendar-day").first().click();
  await expect(page.getByLabel("日期任务抽屉", { exact: true })).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByLabel("日期任务抽屉", { exact: true })).toHaveCount(0);
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.locator(".calendar-day").first().click();
  expect(
    await page
      .getByLabel("日期任务抽屉", { exact: true })
      .evaluate((el) => el.getAnimations().length),
  ).toBe(0);
  await page.keyboard.press("Escape");
  await page.getByLabel("任务列表", { exact: true }).click();
  await page.getByLabel("新增待办", { exact: true }).click();
  await expect(page.getByLabel("新待办标题", { exact: true })).toBeFocused();
  await page.keyboard.press("Escape");
  expect(errors).toEqual([]);
});

test("macOS 最窄标题栏完整，详情退出期间保持可见", async ({
  page,
  request,
}) => {
  await page.route("**/api/Boot", async (route) => {
    const response = await route.fetch();
    const b = await response.json();
    b.capabilities.platform = "darwin";
    await route.fulfill({ response, json: b });
  });
  await page.setViewportSize({ width: 400, height: 720 });
  await call(request, "Settings", { animation: "true" });
  const t = await root(request, unique("退出动效验收"));
  await page.goto("/?preview=1");
  const close = await page
    .getByLabel("关闭窗口", { exact: true })
    .boundingBox();
  expect(close!.x + close!.width).toBeLessThanOrEqual(400);
  await row(page, t.title).hover();
  await row(page, t.title).getByLabel("任务设置").click();
  await page.keyboard.press("Escape");
  const closing = page.locator("dialog.motion-closing");
  if (await closing.count())
    await expect(closing).toHaveCSS("visibility", "visible");
  await expect(page.getByRole("dialog")).toHaveCount(0);
  await page.screenshot({
    path: "../.impeccable/review/motion-mac-400.png",
    animations: "disabled",
  });
});
