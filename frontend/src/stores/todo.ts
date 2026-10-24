import { defineStore } from "pinia";
import { ref, computed, onScopeDispose } from "vue";
import { api, listen, confirmAction } from "../api";
import {
  dayString,
  type Task,
  type Snapshot,
  type Boot,
  type Query,
} from "../api/types";

export const useTodo = defineStore("todo", () => {
  const key = new URLSearchParams(location.search).get("window") || "main";
  const date = ref(dayString(new Date())),
    view = ref(key === "overview" ? "month" : "day"),
    end = ref(date.value),
    limit = ref(500);
  const snapshot = ref<Snapshot>({
    tasks: [],
    counts: {},
    revision: "0",
    today: date.value,
    truncated: false,
  });
  const boot = ref<Boot | null>(null),
    error = ref(""),
    maintenance = ref(false),
    toast = ref(""),
    loading = ref(true),
    saving = ref(false),
    selected = ref<string[]>([]),
    activeTask = ref(""),
    titleEditing = ref(""),
    hideCompleted = ref(false),
    sortMode = ref("manual");
  const query = computed<Query>(() => ({
    view: view.value === "inbox" ? "all" : view.value,
    date: date.value,
    end: view.value === "month" ? end.value : date.value,
    limit: limit.value,
  }));
  const roots = computed(() => {
    let list = snapshot.value.tasks.filter(
      (t) => !t.parentId && t.status !== "skipped",
    );
    if (hideCompleted.value)
      list = list.filter((t) => t.status !== "completed");
    return list.sort((a, b) => {
      if (view.value === "pending" && a.date !== b.date)
        return a.date === null
          ? 1
          : b.date === null
            ? -1
            : a.date.localeCompare(b.date);
      if (
        boot.value?.settings.completedBottom === "true" &&
        a.status !== b.status
      )
        return a.status === "completed" ? 1 : -1;
      if (sortMode.value === "color") {
        const order = [
          "red",
          "orange",
          "yellow",
          "green",
          "blue",
          "purple",
          "none",
        ];
        const n =
          order.indexOf(a.color || "none") - order.indexOf(b.color || "none");
        if (n) return n;
      }
      return a.order - b.order || a.id.localeCompare(b.id);
    });
  });
  const taskMap = computed(
    () => new Map(snapshot.value.tasks.map((t) => [t.id, t])),
  );
  let request = 0,
    refreshTimer: ReturnType<typeof setTimeout> | undefined;
  async function load() {
    if (boot.value?.startupError) {
      loading.value = false;
      return;
    }
    const current = ++request;
    try {
      const result = await api<Snapshot>("Query", query.value);
      if (current === request) {
        snapshot.value = result;
        const ids = new Set(result.tasks.map((t) => t.id));
        selected.value = selected.value.filter((id) => ids.has(id));
      }
    } catch (e) {
      if (current === request) error.value = message(e);
    } finally {
      if (current === request) loading.value = false;
    }
  }
  let timezoneSync = false;
  async function metadata() {
    boot.value = await api<Boot>("Boot");
    const zone = Intl.DateTimeFormat().resolvedOptions().timeZone;
    if (
      !boot.value.startupError &&
      zone &&
      boot.value.timezone &&
      boot.value.timezone !== zone &&
      !timezoneSync
    ) {
      timezoneSync = true;
      try {
        const count = await api<number>("SetTimezone", zone);
        if (count) inform(`系统时区已变更，重新安排 ${count} 条未来重复提醒`);
        boot.value = await api<Boot>("Boot");
      } finally {
        timezoneSync = false;
      }
    }
  }
  async function requestNotifications() {
    if (
      boot.value?.notificationAllowed ||
      boot.value?.settings.notificationPrompted === "true"
    )
      return;
    try {
      await api("Settings", { notificationPrompted: "true" });
      await api("Permission");
      await metadata();
      if (!boot.value?.notificationAllowed)
        inform("通知未获授权，可在设置中查看系统指引；任务已保存");
    } catch (e) {
      inform("内容已保存，通知授权未完成，可在设置中重试");
    }
  }
  function message(e: unknown) {
    return e instanceof Error ? e.message : String(e);
  }
  function inform(text: string) {
    toast.value = text;
    setTimeout(() => {
      if (toast.value === text) toast.value = "";
    }, 3500);
  }
  async function run<T>(fn: () => Promise<T>): Promise<T | undefined> {
    if (maintenance.value) {
      error.value = "正在维护本地数据，请稍后重试";
      return undefined;
    }
    saving.value = true;
    error.value = "";
    try {
      const out = await fn();
      await load();
      return out;
    } catch (e) {
      error.value = message(e);
      return undefined;
    } finally {
      saving.value = false;
    }
  }
  function command(t: Task) {
    const root = taskMap.value.get(t.rootId);
    return {
      commandId: crypto.randomUUID(),
      id: t.id,
      rootId: t.rootId,
      expectedVersion: root?.version || t.version,
    };
  }
  async function complete(t: Task) {
    return run(() =>
      api("Complete", { ...command(t), completed: t.status !== "completed" }),
    );
  }
  async function create(
    title: string,
    parent?: Task,
    plannedDate?: string | null,
    afterId?: string,
  ) {
    const root = parent ? taskMap.value.get(parent.rootId) : undefined;
    return run(() =>
      api<Task[]>("Create", {
        commandId: crypto.randomUUID(),
        title,
        afterId,
        date:
          plannedDate !== undefined
            ? plannedDate
            : view.value === "inbox"
              ? null
              : view.value === "pending"
                ? boot.value?.today || date.value
                : date.value,
        color: "none",
        postpone: "inherit",
        description: "",
        parentId: parent?.id || null,
        rootId: parent?.rootId || "",
        expectedVersion: root?.version || "",
      }),
    );
  }
  async function move(t: Task, date: string | null) {
    if (
      !date &&
      t.remindAt &&
      !(await confirmAction("移入收集箱将取消未发送的提醒。继续？"))
    )
      return;
    return run(() => api<Task[]>("Move", { ...command(t), date }));
  }
  async function remove(t: Task) {
    return run(() => api("Delete", command(t)));
  }
  async function setting(key: string, value: string) {
    await run(() => api("Settings", { [key]: value }));
    await metadata();
  }
  const subscriptions: (() => void)[] = [];
  onScopeDispose(() => {
    clearTimeout(refreshTimer);
    subscriptions.forEach((stop) => stop());
  });
  async function init() {
    try {
      await metadata();
      await load();
      subscriptions.push(
        await listen("todo:changed", () => {
          clearTimeout(refreshTimer);
          refreshTimer = setTimeout(() => {
            void load();
            void metadata().catch((e) => (error.value = message(e)));
          }, 100);
        }),
      );
      subscriptions.push(
        await listen("todo:error", (e) => (error.value = String(e))),
      );
      subscriptions.push(
        await listen("todo:maintenance", (v) => (maintenance.value = !!v)),
      );
    } catch (e) {
      error.value = message(e);
      loading.value = false;
    }
  }
  return {
    key,
    date,
    view,
    end,
    limit,
    snapshot,
    boot,
    error,
    maintenance,
    requestNotifications,
    toast,
    loading,
    saving,
    selected,
    activeTask,
    titleEditing,
    hideCompleted,
    sortMode,
    query,
    roots,
    taskMap,
    load,
    metadata,
    run,
    inform,
    command,
    complete,
    create,
    move,
    remove,
    setting,
    init,
  };
});
