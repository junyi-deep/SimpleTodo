import { expect, type APIRequestContext, type Page } from "@playwright/test";
export const unique = (prefix: string) => `${prefix}-${Date.now()}`;
export async function call(
  request: APIRequestContext,
  name: string,
  ...args: unknown[]
) {
  const r = await request.post(`http://127.0.0.1:9246/api/${name}`, {
    headers: { Origin: "http://127.0.0.1:9245" },
    data: args,
  });
  expect(r.ok(), await r.text()).toBeTruthy();
  return r.json();
}
export async function root(
  request: APIRequestContext,
  title: string,
  date?: string | null,
) {
  const b = await call(request, "Boot");
  const rows = await call(request, "Create", {
    commandId: crypto.randomUUID(),
    title,
    date: date === undefined ? b.today : date,
    color: "none",
    postpone: "off",
  });
  return rows.find((t: any) => !t.parentId);
}
export function row(page: Page, title: string) {
  return page
    .getByRole("treeitem")
    .filter({ has: page.getByRole("button", { name: title, exact: true }) });
}
export async function add(page: Page, title: string) {
  const input = page.getByRole("textbox", { name: "新待办标题" });
  await input.fill(title);
  await input.press("Enter");
  await expect(
    page.getByRole("button", { name: title, exact: true }),
  ).toBeVisible();
}
export async function tools(page: Page, label: string) {
  await page.getByLabel("工具菜单").click();
  await page.getByRole("button", { name: label, exact: true }).click();
}
export async function select(page: Page, label: string, option: string) {
  await page.getByRole("combobox", { name: label, exact: true }).click();
  await page.getByRole("option", { name: option, exact: true }).click();
}
