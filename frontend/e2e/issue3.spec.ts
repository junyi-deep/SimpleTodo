import { test, expect } from "@playwright/test";
import { call, root, row } from "./helpers";
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
    expandDepth: "2",
    animation: "false",
    theme: "light",
    completedBottom: "false",
    autoPostpone: "true",
  });
});
async function child(request: any, parent: any, rootTask: any, title: string) {
  const latest = await call(request, "FindTask", rootTask.id);
  return (
    await call(request, "Create", {
      commandId: crypto.randomUUID(),
      parentId: parent.id,
      rootId: rootTask.id,
      expectedVersion: latest.version,
      title,
    })
  ).find((n: any) => n.title === title);
}
test("默认展开两层，眼睛滑条显示五层并持久化，自动延期默认开启", async ({
  page,
  request,
}) => {
  const a = await root(request, "层级一"),
    b = await child(request, a, a, "层级二"),
    c = await child(request, b, a, "层级三"),
    d = await child(request, c, a, "层级四");
  await page.goto("/?preview=1");
  await expect(row(page, b.title)).toBeVisible();
  await expect(row(page, c.title)).toHaveCount(0);
  await page.getByLabel("展开层级", { exact: true }).click();
  await page.getByLabel("显示层级", { exact: true }).fill("4");
  await expect(row(page, d.title)).toBeVisible();
  await expect
    .poll(async () => (await call(request, "Boot")).settings.expandDepth)
    .toBe("4");
  await page.screenshot({ path: "../.impeccable/review/issue3-depth.png" });
  await page.reload();
  await expect(row(page, d.title)).toBeVisible();
  await page.getByLabel("打开设置", { exact: true }).click();
  await expect(
    page.getByRole("checkbox", { name: /未完成任务自动延期/ }),
  ).toBeChecked();
});
test("左拖线条缩进减少，子任务带整棵分支提升主任务，右拖仍可嵌套", async ({
  page,
  request,
}) => {
  const a = await root(request, "原来的主任务"),
    b = await child(request, a, a, "需要拆出的分支"),
    c = await child(request, b, a, "分支内的叶子"),
    dest = await root(request, "另一个主任务");
  await page.goto("/?preview=1");
  const src = page.locator(`[data-node="${b.id}"] .drag-handle`),
    dst = page.locator(`[data-node="${dest.id}"]`);
  await row(page, b.title).hover();
  const s = (await src.boundingBox())!,
    t = (await dst.boundingBox())!;
  await page.mouse.move(s.x + 5, s.y + 5);
  await page.mouse.down();
  await page.mouse.move(s.x + 27, t.y + 10, { steps: 10 });
  await expect(page.locator(".task-drop-line")).toHaveAttribute(
    "data-depth",
    "2",
  );
  await page.mouse.move(t.x + 3, t.y + 10, { steps: 10 });
  await expect(page.locator(".task-drop-line")).toHaveAttribute(
    "data-depth",
    "1",
  );
  await page.screenshot({
    path: "../.impeccable/review/issue3-promote-line.png",
  });
  await page.mouse.up();
  await expect
    .poll(async () => (await call(request, "FindTask", b.id)).parentId)
    .toBeNull();
  expect(
    (await call(request, "ReadTree", b.id)).find((n: any) => n.id === c.id)
      .rootId,
  ).toBe(b.id);
  expect((await call(request, "FindTask", b.id)).date).toBe(a.date);
  await row(page, b.title).hover();
  const h = (await page
      .locator(`[data-node="${b.id}"] .drag-handle`)
      .boundingBox())!,
    z = (await dst.boundingBox())!;
  await page.mouse.move(h.x + 5, h.y + 5);
  await page.mouse.down();
  await page.mouse.move(h.x + 30, z.y + z.height / 2, { steps: 12 });
  await expect(page.locator(".task-drop-line")).toHaveAttribute(
    "data-depth",
    "2",
  );
  await page.mouse.up();
  await expect
    .poll(
      async () =>
        (await call(request, "ReadTree", dest.id)).find(
          (n: any) => n.id === b.id,
        )?.parentId,
    )
    .toBe(dest.id);
  expect(
    (await call(request, "ReadTree", dest.id)).find((n: any) => n.id === c.id)
      .rootId,
  ).toBe(dest.id);
});
test("最顶部和最底部落点固定为一级，Esc 清理投影且不保存", async ({
  page,
  request,
}) => {
  const a = await root(request, "顶部父任务"),
    b = await child(request, a, a, "子任务"),
    last = await root(request, "底部主任务");
  await page.goto("/?preview=1");
  async function start() {
    await row(page, b.title).hover();
    const h = (await page
      .locator(`[data-node="${b.id}"] .drag-handle`)
      .boundingBox())!;
    await page.mouse.move(h.x + 5, h.y + 5);
    await page.mouse.down();
    return h;
  }
  const h = await start(),
    first = (await row(page, a.title).boundingBox())!;
  await page.mouse.move(h.x + 100, first.y + 2, { steps: 10 });
  await expect(page.locator(".task-drop-line")).toHaveAttribute(
    "data-depth",
    "1",
  );
  await page.keyboard.press("Escape");
  await page.mouse.up();
  await expect(page.locator(".task-drop-line")).toHaveCount(0);
  expect(
    (await call(request, "ReadTree", a.id)).find((n: any) => n.id === b.id)
      .parentId,
  ).toBe(a.id);
  await start();
  const list = (await page.locator(".task-list").boundingBox())!;
  await page.mouse.move(list.x + 200, list.y + list.height - 15, { steps: 10 });
  await expect(page.locator(".task-drop-line")).toHaveAttribute(
    "data-depth",
    "1",
  );
  await page.mouse.up();
  await expect
    .poll(async () => (await call(request, "FindTask", b.id)).parentId)
    .toBeNull();
  expect((await call(request, "FindTask", b.id)).order).toBeGreaterThan(
    (await call(request, "FindTask", last.id)).order,
  );
});
test("升级并重启使用组件确认，取消不安装，失败保留重试和手动下载", async ({
  page,
}) => {
  let installs = 0;
  await page.route("**/api/CheckUpdate", (r) =>
    r.fulfill({
      json: {
        version: "0.2.0",
        available: true,
        downloadURL:
          "https://github.com/junyi-deep/SimpleTodo/releases/download/v0.2.0/SimpleTodo-macos-arm64.zip",
      },
    }),
  );
  await page.route("**/api/InstallUpdate", (r) => {
    installs++;
    return r.fulfill({
      status: 500,
      json: { error: "应用所在目录不可写，原应用未更改" },
    });
  });
  await page.goto("/?preview=1");
  await page.getByLabel("工具菜单").click();
  await page.getByRole("button", { name: "帮助与关于", exact: true }).click();
  await page.getByRole("button", { name: "检查更新", exact: true }).click();
  await page.getByRole("button", { name: "升级并重启", exact: true }).click();
  await expect(page.locator(".component-confirmation")).toBeVisible();
  await page
    .locator(".component-confirmation")
    .getByRole("button", { name: "取消", exact: true })
    .click();
  expect(installs).toBe(0);
  await page.getByRole("button", { name: "升级并重启", exact: true }).click();
  await page
    .locator(".component-confirmation")
    .getByRole("button", { name: "确认", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText("原应用未更改");
  expect(installs).toBe(1);
  await expect(
    page.getByRole("button", { name: "升级并重启", exact: true }),
  ).toBeEnabled();
  await expect(
    page.getByRole("button", { name: "仅下载安装包", exact: true }),
  ).toBeVisible();
  await page.screenshot({ path: "../.impeccable/review/issue3-update.png" });
});

test("完成项置底时根排序与子树提升的线条均停在完成项之前", async ({
  page,
  request,
}) => {
  await call(request, "Settings", { completedBottom: "true" });
  const a = await root(request, "待办 A"),
    done = await root(request, "完成 B"),
    c = await root(request, "待办 C");
  await call(
    request,
    "Reorder",
    [a.id, done.id, c.id],
    (await call(request, "Query", { view: "all" })).revision,
  );
  const doneLatest = await call(request, "FindTask", done.id);
  await call(request, "Complete", {
    commandId: crypto.randomUUID(),
    id: done.id,
    rootId: done.id,
    expectedVersion: doneLatest.version,
    completed: true,
  });
  await page.goto("/?preview=1");
  async function tail(title: string) {
    await row(page, title).hover();
    const h = (await row(page, title).locator(".drag-handle").boundingBox())!,
      list = (await page.locator(".task-list").boundingBox())!;
    await page.mouse.move(h.x + 5, h.y + 5);
    await page.mouse.down();
    await page.mouse.move(list.x + 200, list.y + list.height - 20, {
      steps: 10,
    });
    await expect(page.locator(".task-drop-line")).toHaveAttribute(
      "data-depth",
      "1",
    );
    const line = (await page.locator(".task-drop-line").boundingBox())!,
      box = (await row(page, done.title).boundingBox())!;
    expect(line.y).toBeLessThanOrEqual(box.y + 2);
    await page.screenshot({
      path: "../.impeccable/review/issue3-completed-bottom.png",
    });
    await page.mouse.up();
  }
  await tail(a.title);
  await expect
    .poll(async () => {
      const q = await call(request, "Query", { view: "all" });
      return q.tasks
        .filter((t: any) => !t.parentId && t.status === "pending")
        .sort((a: any, b: any) => a.order - b.order)
        .map((t: any) => t.id);
    })
    .toEqual([c.id, a.id]);
  const branch = await child(request, a, a, "拆出分支"),
    leaf = await child(request, branch, a, "保留叶子");
  await page.reload();
  await tail(branch.title);
  await expect
    .poll(async () => {
      const q = await call(request, "Query", { view: "all" });
      return q.tasks
        .filter((t: any) => !t.parentId && t.status === "pending")
        .sort((a: any, b: any) => a.order - b.order)
        .map((t: any) => t.id);
    })
    .toEqual([c.id, a.id, branch.id]);
  expect(
    (await call(request, "ReadTree", branch.id)).find(
      (t: any) => t.id === leaf.id,
    ).rootId,
  ).toBe(branch.id);
});
