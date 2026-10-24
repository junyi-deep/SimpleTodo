import { beforeEach, expect, it, vi } from "vitest";
beforeEach(() => {
  vi.resetModules();
  vi.clearAllMocks();
  vi.stubGlobal("location", { search: "" });
});
it("确认由组件状态驱动，重复请求不会复用批准", async () => {
  const { confirmAction } = await import("./index");
  const { confirmation, resolveConfirmation } =
    await import("../composables/confirmation");
  let done = false;
  const result = confirmAction("恢复备份？").then((ok) => {
    done = true;
    return ok;
  });
  expect(done).toBe(false);
  expect(confirmation.value?.message).toBe("恢复备份？");
  expect(await confirmAction("删除任务？")).toBe(false);
  resolveConfirmation(confirmation.value!.id, true);
  expect(await result).toBe(true);
  expect(confirmation.value).toBeNull();
});
it("取消安全返回 false，旧组件卸载不会取消新的问题", async () => {
  const { requestConfirmation, confirmation, resolveConfirmation } =
    await import("../composables/confirmation");
  const a = requestConfirmation("问题一");
  const first = confirmation.value!.id;
  resolveConfirmation(first, false);
  expect(await a).toBe(false);
  const b = requestConfirmation("问题二");
  resolveConfirmation(first, false);
  expect(confirmation.value?.message).toBe("问题二");
  resolveConfirmation(confirmation.value!.id, true);
  expect(await b).toBe(true);
});
it("浏览器预览也使用同一组件，不调用浏览器确认或桌面 API", async () => {
  vi.stubGlobal("location", { search: "?preview=1" });
  const native = vi.fn();
  vi.stubGlobal("window", { confirm: native });
  const { confirmAction } = await import("./index");
  const { confirmation, resolveConfirmation } =
    await import("../composables/confirmation");
  const result = confirmAction("放弃修改？");
  resolveConfirmation(confirmation.value!.id, false);
  expect(await result).toBe(false);
  expect(native).not.toHaveBeenCalled();
});
it("空问题不打开弹窗", async () => {
  const { requestConfirmation, confirmation } =
    await import("../composables/confirmation");
  expect(await requestConfirmation("  ")).toBe(false);
  expect(confirmation.value).toBeNull();
});
