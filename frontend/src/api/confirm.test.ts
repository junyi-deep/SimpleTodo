import { beforeEach, expect, it, vi } from "vitest";
const { call } = vi.hoisted(() => ({ call: vi.fn() }));
vi.mock("@wailsio/runtime", () => ({ Call: { ByName: call } }));
beforeEach(() => {
  vi.resetModules();
  vi.clearAllMocks();
});
it("桌面确认等待原生结果并绑定主窗口", async () => {
  vi.stubGlobal("location", { search: "" });
  let answer!: (value: boolean) => void;
  call.mockReturnValueOnce(new Promise((resolve) => (answer = resolve)));
  const { confirmAction } = await import("./index");
  let completed = false;
  const pending = confirmAction("恢复备份？").then((ok) => {
    completed = true;
    return ok;
  });
  await vi.waitFor(() =>
    expect(call).toHaveBeenCalledWith(
      "main.TodoService.Confirm",
      "main",
      "恢复备份？",
    ),
  );
  expect(completed).toBe(false);
  expect(await confirmAction("另一项删除？")).toBe(false);
  expect(call).toHaveBeenCalledTimes(1);
  answer(true);
  expect(await pending).toBe(true);
});
it("浮窗取消通过原生结果返回 false", async () => {
  vi.stubGlobal("location", { search: "?window=mini" });
  call.mockResolvedValueOnce(false);
  const { confirmAction } = await import("./index");
  expect(await confirmAction("删除分支？")).toBe(false);
  expect(call).toHaveBeenCalledWith(
    "main.TodoService.Confirm",
    "mini",
    "删除分支？",
  );
});
it("开发预览使用浏览器确认且不调用桌面服务", async () => {
  vi.stubGlobal("location", { search: "?preview=1&window=overview" });
  const confirm = vi.fn().mockReturnValueOnce(false).mockReturnValueOnce(true);
  vi.stubGlobal("window", { confirm });
  const { confirmAction } = await import("./index");
  expect(await confirmAction("取消恢复？")).toBe(false);
  expect(await confirmAction("确认恢复？")).toBe(true);
  expect(call).not.toHaveBeenCalled();
});
it("桥接失败安全取消、显示原因并允许再次确认", async () => {
  vi.stubGlobal("location", { search: "?window=main" });
  const dispatchEvent = vi.fn();
  vi.stubGlobal("window", { dispatchEvent });
  call
    .mockRejectedValueOnce(new Error("窗口不可用"))
    .mockResolvedValueOnce(true);
  const { confirmAction } = await import("./index");
  expect(await confirmAction("恢复备份？")).toBe(false);
  expect(dispatchEvent).toHaveBeenCalledTimes(1);
  const event = dispatchEvent.mock.calls[0][0] as CustomEvent;
  expect(event.type).toBe("todo:confirm-error");
  expect(event.detail).toContain("窗口不可用");
  expect(await confirmAction("重新恢复？")).toBe(true);
});
