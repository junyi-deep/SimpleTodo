import { test, expect } from "@playwright/test";
import { call, root, row, select } from "./helpers";
const defaults = {
  dayShowCompleted: "true",
  inboxShowCompleted: "false",
  dayShowRecurring: "true",
  inboxShowRecurring: "true",
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
    ...defaults,
    animation: "true",
    completedBottom: "false",
    theme: "light",
  });
});
test.afterEach(async ({ request }) => {
  await call(request, "Settings", defaults);
  await call(request, "WindowConfig", "main", "normal", 1, false);
});
test("过滤按视图保存，日计划显示完成项，收集箱隐藏完成项，重新打开恢复选择", async ({
  page,
  request,
}) => {
  const a = await root(request, "完成项"),
    b = await root(request, "周期项");
  await call(request, "Complete", {
    commandId: crypto.randomUUID(),
    id: a.id,
    rootId: a.id,
    expectedVersion: a.version,
    completed: true,
  });
  await call(request, "Repeat", {
    command: {
      commandId: crypto.randomUUID(),
      id: b.id,
      rootId: b.id,
      expectedVersion: b.version,
    },
    rule: { frequency: "daily", interval: 1, weekdays: [], until: null },
    following: true,
  });
  await page.goto("/?preview=1");
  await expect(row(page, a.title)).toBeVisible();
  await page.getByLabel("过滤显示", { exact: true }).click();
  await expect(
    page.getByLabel("显示已完成的任务", { exact: true }),
  ).toBeChecked();
  await page.getByLabel("显示周期任务", { exact: true }).uncheck();
  await expect(row(page, b.title)).toHaveCount(0);
  await page.getByLabel("显示已完成的任务", { exact: true }).uncheck();
  await expect(row(page, a.title)).toHaveCount(0);
  await page.getByRole("button", { name: "日计划", exact: true }).click();
  await expect(
    page.getByRole("button", { name: "收集箱", exact: true }),
  ).toBeVisible();
  await expect(row(page, a.title)).toHaveCount(0);
  await expect(row(page, b.title).first()).toBeVisible();
  await page.getByLabel("过滤显示", { exact: true }).click();
  await page.getByLabel("显示已完成的任务", { exact: true }).check();
  await expect(row(page, a.title)).toBeVisible();
  await page.reload();
  await expect(row(page, a.title)).toHaveCount(0);
  await page.getByRole("button", { name: "日计划", exact: true }).click();
  await expect(row(page, a.title)).toBeVisible();
  await page.getByLabel("过滤显示", { exact: true }).click();
  await expect(
    page.getByLabel("显示已完成的任务", { exact: true }),
  ).toBeChecked();
  await expect(page.getByLabel("显示周期任务", { exact: true })).toBeChecked();
  await page.screenshot({ path: "../.impeccable/review/list-filters.png" });
});
test("上下移按同级位置显示，跨日期收集箱排序不改变日期", async ({
  page,
  request,
}) => {
  const a = await root(request, "第一条"),
    b = await root(request, "中间条"),
    c = await root(request, "最后条", null);
  await page.goto("/?preview=1");
  await page.getByRole("button", { name: "日计划", exact: true }).click();
  await row(page, a.title).hover();
  await expect(
    row(page, a.title).getByLabel("上移任务", { exact: true }),
  ).toHaveCount(0);
  await row(page, a.title).getByLabel("下移任务", { exact: true }).click();
  await expect
    .poll(
      async () =>
        (await call(request, "FindTask", a.id)).order >
        (await call(request, "FindTask", b.id)).order,
    )
    .toBe(true);
  await row(page, a.title).hover();
  await expect(
    row(page, a.title).getByLabel("上移任务", { exact: true }),
  ).toBeVisible();
  await expect(
    row(page, a.title).getByLabel("下移任务", { exact: true }),
  ).toBeVisible();
  await row(page, c.title).hover();
  await expect(
    row(page, c.title).getByLabel("下移任务", { exact: true }),
  ).toHaveCount(0);
  await row(page, c.title).getByLabel("上移任务", { exact: true }).click();
  expect((await call(request, "FindTask", c.id)).date).toBeNull();
  const q = await call(request, "Query", { view: "all" });
  expect(
    q.tasks
      .filter((t: any) => !t.parentId)
      .sort((a: any, b: any) => a.order - b.order)
      .map((t: any) => t.id),
  ).toEqual([b.id, c.id, a.id]);
});
test("拖动有真实间隙，拖到空白区排到末尾，取消拖动移除间隙", async ({
  page,
  request,
}) => {
  const a = await root(request, "拖动首条"),
    b = await root(request, "拖动中间"),
    c = await root(request, "拖动末条");
  await page.goto("/?preview=1");
  const handle = page.locator(`[data-node="${a.id}"] .drag-handle`),
    target = page.locator(`[data-node="${b.id}"]`);
  await row(page, a.title).hover();
  const h = (await handle.boundingBox())!,
    before = (await target.boundingBox())!;
  await page.mouse.move(h.x + 5, h.y + 5);
  await page.mouse.down();
  await page.mouse.move(before.x + 25, before.y + 2, { steps: 10 });
  await expect(page.locator(".task-sort-gap")).toBeVisible();
  await expect
    .poll(async () => (await target.boundingBox())!.y - before.y)
    .toBeGreaterThan(20);
  await page.screenshot({ path: "../.impeccable/review/list-drag-gap.png" });
  await page.keyboard.press("Escape");
  await page.mouse.up();
  await expect(page.locator(".task-sort-gap")).toHaveCount(0);
  await row(page, a.title).hover();
  const h2 = (await handle.boundingBox())!,
    list = (await page.locator(".task-list").boundingBox())!;
  await page.mouse.move(h2.x + 5, h2.y + 5);
  await page.mouse.down();
  await page.mouse.move(list.x + list.width / 2, list.y + list.height - 35, {
    steps: 12,
  });
  await expect(page.locator("[data-tail-drop]")).toBeVisible();
  await page.mouse.up();
  await expect
    .poll(async () => {
      const q = await call(request, "Query", { view: "all" });
      return q.tasks
        .filter((t: any) => !t.parentId)
        .sort((a: any, b: any) => a.order - b.order)
        .map((t: any) => t.id);
    })
    .toEqual([b.id, c.id, a.id]);
  await expect(page.locator(".task-sort-gap")).toHaveCount(0);
});
test("长标题改期不横向溢出，重复信息在标题下显示，透明度即时预览并保存", async ({
  page,
  request,
}) => {
  const a = await root(request, "long".repeat(120));
  await page.goto("/?preview=1");
  await row(page, a.title).hover();
  await row(page, a.title).getByLabel("调整计划日期", { exact: true }).click();
  const dialog = page.getByRole("dialog", {
    name: "调整计划日期",
    exact: true,
  });
  expect(await dialog.evaluate((el) => el.scrollWidth <= el.clientWidth)).toBe(
    true,
  );
  expect(
    await dialog
      .locator(".dialog-body")
      .evaluate((el) => el.scrollWidth <= el.clientWidth),
  ).toBe(true);
  await page.screenshot({
    path: "../.impeccable/review/list-long-placement.png",
  });
  await select(page, "计划重复方式", "每周");
  await dialog.getByRole("button", { name: "确认移动", exact: true }).click();
  await expect(page.locator(".task-repeat")).toContainText("每周");
  await expect(page.locator(".task-indicator")).toHaveCount(0);
  await page.getByRole("button", { name: "打开设置", exact: true }).click();
  await page.getByRole("button", { name: "外观与窗口", exact: true }).click();
  const opacity = page.getByLabel("背景不透明度", { exact: true });
  await opacity.fill("0.6");
  await expect(page.locator(".compact-app")).toHaveCSS("--window-alpha", "0.6");
  expect(
    await page
      .locator(".compact-app")
      .evaluate((el) => getComputedStyle(el).backgroundColor),
  ).toContain("0.6");
  await page.getByRole("button", { name: "应用窗口设置", exact: true }).click();
  await expect
    .poll(
      async () =>
        (await call(request, "Boot")).windows.find((w: any) => w.key === "main")
          ?.opacity,
    )
    .toBe(0.6);
  await page.getByLabel("关闭设置", { exact: true }).click();
  await page.reload();
  await expect(page.locator(".compact-app")).toHaveCSS("--window-alpha", "0.6");
  expect(
    await page
      .locator(".compact-app")
      .evaluate((el) => getComputedStyle(el).backgroundColor),
  ).toContain("0.6");
});

test("展开子分支与根行下缘的排序占位在完整后代之后，落地一致", async ({
  page,
  request,
}) => {
  const parent = await root(request, "展开根任务");
  async function child(title: string, parentId: string) {
    const latest = await call(request, "FindTask", parent.id);
    const tree = await call(request, "Create", {
      commandId: crypto.randomUUID(),
      title,
      parentId,
      rootId: parent.id,
      expectedVersion: latest.version,
    });
    return tree.find((n: any) => n.title === title);
  }
  const branch = await child("有后代的同级", parent.id),
    leaf = await child("分支后代", branch.id),
    source = await child("移动的同级", parent.id);
  for (const target of [branch, parent]) {
    const latest = await call(request, "FindTask", parent.id);
    await call(request, "Reparent", {
      commandId: crypto.randomUUID(),
      id: source.id,
      rootId: parent.id,
      expectedVersion: latest.version,
      parentId: parent.id,
      beforeId: branch.id,
    });
    await page.goto("/?preview=1");
    await row(page, parent.title)
      .getByLabel("展开子任务", { exact: true })
      .click();
    await row(page, branch.title)
      .getByLabel("展开子任务", { exact: true })
      .click();
    const src = page.locator(`[data-node="${source.id}"] .drag-handle`),
      dst = page.locator(`[data-node="${target.id}"]`);
    await row(page, source.title).hover();
    const a = (await src.boundingBox())!,
      b = (await dst.boundingBox())!;
    await page.mouse.move(a.x + 5, a.y + 5);
    await page.mouse.down();
    await page.mouse.move(b.x + 28, b.y + b.height - 2, { steps: 10 });
    const gap = page.locator(".task-sort-gap");
    await expect(gap).toBeVisible();
    expect(
      await gap.evaluate(
        (el) =>
          el.previousElementSibling?.querySelector<HTMLElement>("[data-node]")
            ?.dataset.node,
      ),
    ).toBe(leaf.id);
    await page.screenshot({
      path: `../.impeccable/review/list-gap-${target.id === parent.id ? "root" : "branch"}.png`,
    });
    await page.mouse.up();
    await expect
      .poll(async () => {
        const tree = await call(request, "ReadTree", parent.id);
        return tree
          .filter((n: any) => n.parentId === parent.id)
          .sort((a: any, b: any) => a.order - b.order)
          .map((n: any) => n.id);
      })
      .toEqual([branch.id, source.id]);
    await expect(gap).toHaveCount(0);
  }
});

test("子任务拖到空白时只排到当前父任务末尾，间隙不会跑到其它根任务后", async ({
  page,
  request,
}) => {
  const parent = await root(request, "尾部落点父任务");
  const a = (
    await call(request, "Create", {
      commandId: crypto.randomUUID(),
      title: "前一个子任务",
      parentId: parent.id,
      rootId: parent.id,
      expectedVersion: parent.version,
    })
  ).find((n: any) => n.parentId);
  const latest = await call(request, "FindTask", parent.id);
  const b = (
    await call(request, "Create", {
      commandId: crypto.randomUUID(),
      title: "后一个子任务",
      parentId: parent.id,
      rootId: parent.id,
      expectedVersion: latest.version,
    })
  ).find((n: any) => n.title === "后一个子任务");
  await root(request, "下一个根任务");
  await page.goto("/?preview=1");
  await row(page, parent.title)
    .getByLabel("展开子任务", { exact: true })
    .click();
  await row(page, a.title).hover();
  const h = (await page
      .locator(`[data-node="${a.id}"] .drag-handle`)
      .boundingBox())!,
    list = (await page.locator(".task-list").boundingBox())!;
  await page.mouse.move(h.x + 5, h.y + 5);
  await page.mouse.down();
  await page.mouse.move(list.x + list.width / 2, list.y + list.height - 35, {
    steps: 12,
  });
  const gap = page.locator(".task-sort-gap");
  await expect(gap).toBeVisible();
  expect(
    await gap.evaluate(
      (el) =>
        el.previousElementSibling?.querySelector<HTMLElement>("[data-node]")
          ?.dataset.node,
    ),
  ).toBe(b.id);
  await page.mouse.up();
  await expect(gap).toHaveCount(0);
  await expect
    .poll(async () =>
      (await call(request, "ReadTree", parent.id))
        .filter((n: any) => n.parentId === parent.id)
        .sort((a: any, b: any) => a.order - b.order)
        .map((n: any) => n.id),
    )
    .toEqual([b.id, a.id]);
});
