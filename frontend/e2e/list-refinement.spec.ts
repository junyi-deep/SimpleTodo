import { test, expect } from "@playwright/test";
import { root, row, call, unique } from "./helpers";

test.beforeEach(async ({ request }) => {
  const empty = await call(
    request,
    "PreviewImport",
    JSON.stringify({ formatVersion: 1, tasks: [] }),
    "json",
    null,
    true,
  );
  await call(request, "CommitImport", empty.token);
  await call(request, "Settings", {
    animation: "true",
    completedBottom: "false",
  });
});

test("五层展开图标保持朝向和对齐，第六层快捷新增被阻止", async ({
  page,
  request,
}) => {
  const parent = await root(request, unique("第一层"));
  const chain = [parent];
  for (let i = 2; i <= 5; i++) {
    const latest = await call(request, "FindTask", parent.id);
    const tasks = await call(request, "Create", {
      commandId: crypto.randomUUID(),
      title: `第${i}层`,
      rootId: parent.id,
      parentId: chain.at(-1).id,
      expectedVersion: latest.version,
    });
    chain.push(tasks.find((t: any) => t.title === `第${i}层`));
  }
  await page.goto("/?preview=1");
  for (const task of chain.slice(0, 4)) {
    const r = row(page, task.title);
    const toggle = r.locator(".tree-toggle");
    for (let i = 0; i < 3; i++) {
      await toggle.click();
      await expect(toggle).toHaveAttribute("aria-expanded", "true");
      await expect(toggle.locator("svg")).toHaveCSS("rotate", "90deg");
      if (i < 2) await toggle.click();
    }
    const arrow = (await toggle.boundingBox())!,
      title = (await r.locator(".task-title").boundingBox())!;
    expect(
      Math.abs(arrow.y + arrow.height / 2 - title.y - title.height / 2),
    ).toBeLessThan(3);
  }
  await row(page, chain[4].title).click();
  await page.keyboard.press("Shift+Enter");
  await expect(page.getByRole("status")).toContainText("最多支持 5 层");
  await expect(page.locator("[data-new-child]")).toHaveCount(0);
  await page.screenshot({
    path: "../.impeccable/review/list-five-levels.png",
    animations: "disabled",
  });
});

test("双击光标在行尾，柔和编辑样式，完成后标题变浅并划线", async ({
  page,
  request,
}) => {
  const t = await root(request, "无需覆盖的原始标题");
  await page.goto("/?preview=1");
  await row(page, t.title)
    .getByRole("button", { name: t.title, exact: true })
    .dblclick();
  const input = page.getByLabel("编辑任务标题", { exact: true });
  await expect(input).toBeFocused();
  expect(
    await input.evaluate((el: HTMLTextAreaElement) => [
      el.selectionStart,
      el.selectionEnd,
    ]),
  ).toEqual([t.title.length, t.title.length]);
  await expect(input).toHaveCSS("box-shadow", "none");
  await expect(input).toHaveCSS("outline-style", "none");
  await input.press("End");
  await page.keyboard.insertText("追加");
  await input.press("Enter");
  const updated = row(page, t.title + "追加"),
    title = updated.locator(".task-title");
  const pendingColor = await title.evaluate((el) => getComputedStyle(el).color);
  await updated.getByLabel("完成任务及所有后代", { exact: true }).click();
  await expect(title).toHaveCSS("text-decoration-line", "line-through");
  expect(await title.evaluate((el) => getComputedStyle(el).color)).not.toBe(
    pendingColor,
  );
});

test("计划日期为模态框，不挤动列表，Esc 返回任务", async ({
  page,
  request,
}) => {
  const t = await root(request, unique("日期弹窗"));
  await page.goto("/?preview=1");
  const r = row(page, t.title);
  await expect
    .poll(() =>
      page.locator(".task-list").evaluate((el) => el.getAnimations().length),
    )
    .toBe(0);
  const before = (await r.boundingBox())!;
  await r.getByLabel("更多任务操作").click();
  await page
    .getByRole("button", { name: "调整计划日期…", exact: true })
    .click();
  const dialog = page.getByRole("dialog", {
    name: "调整计划日期",
    exact: true,
  });
  await expect(dialog).toBeVisible();
  expect(await dialog.evaluate((el) => el.matches(":modal"))).toBe(true);
  expect((await r.boundingBox())!.y).toBe(before.y);
  await page.screenshot({
    path: "../.impeccable/review/list-date-dialog.png",
    animations: "disabled",
  });
  await page.keyboard.press("Escape");
  await expect(dialog).toHaveCount(0);
  await expect(r).toBeFocused();
});

test("真实指针拖动可取消，边缘排序保持同级并持久化", async ({
  page,
  request,
}) => {
  const a = await root(request, "排序 A"),
    b = await root(request, "排序 B"),
    c = await root(request, "排序 C");
  await page.goto("/?preview=1");
  async function dragStart() {
    await row(page, c.title).locator(".drag-handle").hover();
    const grip = (await row(page, c.title)
      .locator(".drag-handle")
      .boundingBox())!;
    await page.mouse.move(grip.x + grip.width / 2, grip.y + grip.height / 2);
    await page.mouse.down();
    await page.mouse.move(grip.x + 15, grip.y - 8, { steps: 4 });
    await expect(page.locator(".pointer-drag-preview")).toHaveCount(1);
  }
  await dragStart();
  await page.keyboard.press("Escape");
  await page.mouse.up();
  await expect(page.locator(".pointer-drag-preview")).toHaveCount(0);
  expect((await call(request, "FindTask", c.id)).parentId).toBeNull();
  await dragStart();
  const target = (await row(page, a.title).boundingBox())!;
  await page.mouse.move(target.x + 100, target.y + 2, { steps: 12 });
  await expect(row(page, a.title)).toHaveAttribute(
    "data-drop-placement",
    "before",
  );
  await page.mouse.up();
  await expect
    .poll(async () =>
      (await call(request, "Query", { view: "day", date: a.date })).tasks
        .filter((t: any) => !t.parentId)
        .sort((a: any, b: any) => a.order - b.order)
        .map((t: any) => t.id),
    )
    .toEqual([c.id, a.id, b.id]);
  await page.reload();
  await expect(page.locator("[data-node]").first()).toHaveAttribute(
    "data-node",
    c.id,
  );
});

test("日计划连续切换日期滑块落点准确，减少动态效果立即结束动效", async ({
  page,
  request,
}) => {
  await root(request, "今日计划");
  await page.goto("/?preview=1");
  const days = page.locator(".week-date");
  for (const i of [1, 5, 2, 6]) await days.nth(i).click();
  const selected = page.locator(".week-date.selected"),
    pill = page.locator(".timeline-pill");
  await expect(selected).toHaveAttribute("aria-current", "date");
  await expect
    .poll(async () =>
      Math.abs(
        (await pill.boundingBox())!.x - (await selected.boundingBox())!.x,
      ),
    )
    .toBeLessThan(1);
  await days.nth(0).click();
  await page.emulateMedia({ reducedMotion: "reduce" });
  await expect
    .poll(() => pill.evaluate((el) => el.getAnimations().length))
    .toBe(0);
  expect(
    Math.abs((await pill.boundingBox())!.x - (await selected.boundingBox())!.x),
  ).toBeLessThan(1);
  await page.screenshot({
    path: "../.impeccable/review/list-day-timeline.png",
    animations: "disabled",
  });
});

test("跨根拖到子任务边缘插入同级", async ({ page, request }) => {
  const parent = await root(request, "目标分支"),
    source = await root(request, "外部任务");
  let last: any;
  for (const title of ["目标 A", "目标 B"]) {
    const latest = await call(request, "FindTask", parent.id);
    const tree = await call(request, "Create", {
      commandId: crypto.randomUUID(),
      title,
      rootId: parent.id,
      parentId: parent.id,
      expectedVersion: latest.version,
    });
    last = tree.find((t: any) => t.title === title);
  }
  await page.goto("/?preview=1");
  await row(page, parent.title).getByLabel("展开子任务").click();
  await row(page, source.title).locator(".drag-handle").hover();
  const grip = (await row(page, source.title)
    .locator(".drag-handle")
    .boundingBox())!;
  const target = (await row(page, last.title).boundingBox())!;
  await page.mouse.move(grip.x + grip.width / 2, grip.y + grip.height / 2);
  await page.mouse.down();
  await page.mouse.move(target.x + 100, target.y + 2, { steps: 15 });
  await expect(row(page, last.title)).toHaveAttribute(
    "data-drop-placement",
    "before",
  );
  await page.mouse.up();
  await expect
    .poll(async () =>
      (await call(request, "ReadTree", parent.id))
        .filter((t: any) => t.parentId === parent.id)
        .sort((a: any, b: any) => a.order - b.order)
        .map((t: any) => t.title),
    )
    .toEqual(["目标 A", "外部任务", "目标 B"]);
});
