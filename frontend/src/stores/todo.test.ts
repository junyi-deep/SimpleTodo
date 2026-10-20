import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useTodo } from "./todo";
import { api } from "../api";
import { shiftDay, type Snapshot } from "../api/types";
vi.mock("../api", () => ({ api: vi.fn(), listen: vi.fn() }));
const snapshot = (revision: string): Snapshot => ({
  tasks: [],
  counts: { pending: 0 },
  revision,
  today: "2026-10-20",
  truncated: false,
});
beforeEach(() => {
  vi.clearAllMocks();
  vi.stubGlobal("location", { search: "" });
  setActivePinia(createPinia());
});
describe("跨窗口与异步视图状态", () => {
  it("快速切换日期时丢弃迟到的旧查询结果", async () => {
    let first!: (s: Snapshot) => void, second!: (s: Snapshot) => void;
    vi.mocked(api)
      .mockImplementationOnce(() => new Promise((r) => (first = r)))
      .mockImplementationOnce(() => new Promise((r) => (second = r)));
    const store = useTodo();
    const a = store.load();
    store.date = "2026-10-21";
    const b = store.load();
    second(snapshot("2"));
    await b;
    first(snapshot("1"));
    await a;
    expect(store.snapshot.revision).toBe("2");
    expect(store.date).toBe("2026-10-21");
  });
  it("保存冲突不清空原快照且释放保存状态", async () => {
    const store = useTodo();
    store.snapshot = snapshot("9");
    await store.run(() =>
      Promise.reject(new Error("CONFLICT: 任务已在其他窗口更新")),
    );
    expect(store.snapshot.revision).toBe("9");
    expect(store.error).toContain("CONFLICT");
    expect(store.saving).toBe(false);
  });
  it("读取失败不把既有数据误显示为空库", async () => {
    const store = useTodo();
    store.snapshot = snapshot("8");
    vi.mocked(api).mockRejectedValueOnce(new Error("IO_ERROR: 读取失败"));
    await store.load();
    expect(store.snapshot.revision).toBe("8");
    expect(store.error).toContain("读取失败");
  });
});
it("日期栏日历运算跨越月末、闰日和年末", () => {
  expect(shiftDay("2024-02-28", 1)).toBe("2024-02-29");
  expect(shiftDay("2024-02-29", 1)).toBe("2024-03-01");
  expect(shiftDay("2026-12-31", 1)).toBe("2027-01-01");
});

it("数据维护期间阻止新写入并保留原快照", async () => {
  const store = useTodo();
  store.snapshot = snapshot("12");
  store.maintenance = true;
  const write = vi.fn(async () => true);
  const result = await store.run(write);
  expect(write).not.toHaveBeenCalled();
  expect(result).toBeUndefined();
  expect(store.snapshot.revision).toBe("12");
  expect(store.error).toContain("维护");
});
