import { test, expect } from "@playwright/test";

async function loadCaptureImages(page: import("@playwright/test").Page) {
  await page.locator("img").evaluateAll(async (items) => {
    await Promise.all(
      items.map(async (item) => {
        const img = item as HTMLImageElement;
        img.loading = "eager";
        await img.decode();
      }),
    );
  });
}

test("桌面官网：真实图片、下载入口、功能切换与键盘操作", async ({ page }) => {
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.goto("./");
  await expect(
    page.getByRole("heading", { name: /少一点纷扰，\s*多一点完成。/ }),
  ).toBeVisible();
  await expect(
    page.getByRole("link", { name: "下载 SimpleTodo macOS Apple Silicon 版" }),
  ).toHaveAttribute(
    "href",
    "https://github.com/junyi-deep/SimpleTodo/releases/latest/download/SimpleTodo-macos-arm64.zip",
  );
  await expect(
    page.getByRole("link", { name: "下载 SimpleTodo Windows 64 位版" }),
  ).toHaveAttribute(
    "href",
    "https://github.com/junyi-deep/SimpleTodo/releases/latest/download/SimpleTodo-windows-amd64.zip",
  );
  for (const tab of ["任务清单", "日历规划", "专注时钟", "复盘笔记"]) {
    await page.getByRole("tab", { name: tab, exact: true }).click();
    await expect(
      page.getByRole("tab", { name: tab, exact: true }),
    ).toHaveAttribute("aria-selected", "true");
    await expect(page.locator(".feature-image img")).toHaveAttribute(
      "alt",
      new RegExp(tab),
    );
    await expect
      .poll(() =>
        page
          .locator(".feature-image img")
          .evaluate((img) => (img as HTMLImageElement).naturalWidth),
      )
      .toBeGreaterThan(0);
  }
  await page.getByRole("tab", { name: "复盘笔记", exact: true }).press("Home");
  await expect(
    page.getByRole("tab", { name: "任务清单", exact: true }),
  ).toBeFocused();
  await page
    .getByRole("tab", { name: "任务清单", exact: true })
    .press("ArrowRight");
  await expect(
    page.getByRole("tab", { name: "日历规划", exact: true }),
  ).toBeFocused();
  await page
    .getByRole("button", { name: /Shift.*Tab.*让层级随想法调整/ })
    .click();
  await expect(page.locator(".shortcut-description")).toHaveText(
    "将子任务往上提升一级，带上它的全部步骤。",
  );
  await page.getByText("断网时还能使用吗？", { exact: true }).click();
  await expect(
    page
      .locator("details")
      .filter({ has: page.getByText("断网时还能使用吗？", { exact: true }) }),
  ).toHaveAttribute("open", "");
  await page.getByRole("tab", { name: "任务清单", exact: true }).click();
  await page.evaluate(() => window.scrollTo(0, 0));
  await loadCaptureImages(page);
  await page.screenshot({
    path: "../.impeccable/review/site-desktop.png",
    fullPage: true,
    animations: "disabled",
  });
  await page.screenshot({
    path: "../.impeccable/review/site-desktop-hero.png",
    animations: "disabled",
  });
  for (const link of await page.locator('a[href^="#"]').all()) {
    const target = await link.getAttribute("href");
    expect(await page.locator(target!).count()).toBe(1);
  }
  const images = await page
    .locator("img")
    .evaluateAll((items) =>
      items.map((img) => ({
        src: img.getAttribute("src"),
        complete: (img as HTMLImageElement).complete,
        width: (img as HTMLImageElement).naturalWidth,
      })),
    );
  expect(
    images.every((img) => img.complete && img.width > 0),
    JSON.stringify(images),
  ).toBe(true);
  expect(
    await page.locator("body").evaluate((el) => el.scrollWidth <= innerWidth),
  ).toBe(true);
  expect(errors).toEqual([]);
});

test("移动官网：菜单、锚点、FAQ和图片在窄屏可用", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.emulateMedia({ colorScheme: "light", reducedMotion: "reduce" });
  await page.goto("./");
  await page.getByRole("button", { name: "打开导航菜单" }).click();
  await expect(
    page.getByRole("navigation", { name: "移动端导航" }),
  ).toBeVisible();
  await page
    .getByRole("navigation", { name: "移动端导航" })
    .getByRole("link", { name: "功能", exact: true })
    .click();
  await expect(
    page.getByRole("navigation", { name: "移动端导航" }),
  ).toHaveCount(0);
  await page.getByRole("tab", { name: "复盘笔记", exact: true }).click();
  await expect(page.getByRole("tabpanel")).toContainText(
    "看见进展，也记下思考。",
  );
  await page.getByRole("tab", { name: "任务清单", exact: true }).click();
  await page.getByRole("button", { name: "打开导航菜单" }).click();
  await page.getByRole("button", { name: "关闭导航菜单" }).press("Escape");
  await expect(
    page.getByRole("button", { name: "打开导航菜单" }),
  ).toBeFocused();
  await page.getByRole("button", { name: "打开导航菜单" }).blur();
  await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
  await page.evaluate(() => window.scrollTo(0, 0));
  await loadCaptureImages(page);
  await page.screenshot({
    path: "../.impeccable/review/site-mobile.png",
    fullPage: true,
    animations: "disabled",
  });
  await page.screenshot({
    path: "../.impeccable/review/site-mobile-hero.png",
    animations: "disabled",
  });
  expect(
    await page.locator("body").evaluate((el) => el.scrollWidth <= innerWidth),
  ).toBe(true);
  expect(
    await page
      .locator("body")
      .evaluate((el) => getComputedStyle(el).backgroundColor),
  ).toBe("rgb(8, 9, 11)");
  await page.setViewportSize({ width: 375, height: 812 });
  expect(
    await page.locator("body").evaluate((el) => el.scrollWidth <= innerWidth),
  ).toBe(true);
});

test("1024桌面与正常动效：导航单行、按钮可见且无页面错误", async ({ page }) => {
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  await page.setViewportSize({ width: 1024, height: 800 });
  await page.emulateMedia({ reducedMotion: "no-preference" });
  await page.goto("./");
  await expect(page.locator(".hero-copy h1")).toBeVisible();
  const cta = await page.locator(".hero-actions").boundingBox();
  expect(cta!.y + cta!.height).toBeLessThan(800);
  await page.getByRole("tab", { name: "专注时钟", exact: true }).click();
  await expect
    .poll(() =>
      page
        .locator(".feature-image img")
        .evaluate((img) => (img as HTMLImageElement).naturalWidth),
    )
    .toBeGreaterThan(0);
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.getByRole("tab", { name: "任务清单", exact: true }).click();
  await page.evaluate(() => window.scrollTo(0, 0));
  await loadCaptureImages(page);
  await page.screenshot({
    path: "../.impeccable/review/site-1024.png",
    fullPage: true,
    animations: "disabled",
  });
  expect(
    await page.locator("body").evaluate((el) => el.scrollWidth <= innerWidth),
  ).toBe(true);
  expect(errors).toEqual([]);
});
