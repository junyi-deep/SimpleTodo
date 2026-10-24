import { defineStore } from "pinia";
import { ref, computed, onScopeDispose } from "vue";
import { api, listen, confirmAction } from "../api";
import {
  dayString,
  type Task,
  type Snapshot,
  type Boot,
  type Query,
  type Rule,
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
    rules = ref<Rule[]>([]),
    sortMode = ref("manual");
  const filterView = computed(() => (view.value === "inbox" ? "inbox" : "day"));
  const hideCompleted = computed(
    () =>
      (boot.value?.settings[filterView.value + "ShowCompleted"] ??
        (filterView.value === "day" ? "true" : "false")) !== "true",
  );
  const showRecurring = computed(
    () => boot.value?.settings[filterView.value + "ShowRecurring"] !== "false",
  );
  async function setFilter(kind: "Completed" | "Recurring", checked: boolean) {
    await setting(filterView.value + "Show" + kind, String(checked));
    selected.value = [];
  }
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
    if (!showRecurring.value) list = list.filter((t) => !t.seriesId);
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
        if (
          result.tasks.some(
            (t) => t.ruleId && !rules.value.some((r) => r.id === t.ruleId),
          )
        )
          rules.value = await api<Rule[]>("Rules");
        if (current !== request) return;
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
    rules.value = boot.value.startupError ? [] : await api<Rule[]>("Rules");
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
        await api<number>("SetTimezone", zone);
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
  async function reorderRoot(source: Task, target?: Task, after = false) {
    if (
      source.status !== "pending" ||
      source.parentId ||
      sortMode.value !== "manual"
    )
      return;
    const peers = snapshot.value.tasks
      .filter((t) => !t.parentId && t.status !== "skipped")
      .sort((a, b) => a.order - b.order || a.id.localeCompare(b.id));
    const ids = peers.filter((t) => t.id !== source.id).map((t) => t.id);
    if (target && !ids.includes(target.id)) return;
    const lastPending = [...peers]
      .reverse()
      .find(
        (t) =>
          t.id !== source.id &&
          t.status === "pending" &&
          roots.value.some((r) => r.id === t.id),
      );
    const at = target
      ? ids.indexOf(target.id) + (after ? 1 : 0)
      : lastPending
        ? ids.indexOf(lastPending.id) + 1
        : ids.length;
    ids.splice(at, 0, source.id);
    return run(() => api("Reorder", ids, snapshot.value.revision));
  }
  function repeatSummary(t: Task) {
    const r = rules.value.find((r) => r.id === t.ruleId);
    if (!r) return "重复任务";
    const unit =
      r.frequency === "daily" ? "天" : r.frequency === "weekly" ? "周" : "月";
    let text =
      r.interval === 1
        ? { daily: "每天", weekly: "每周", monthly: "每月" }[r.frequency] ||
          "重复"
        : `每 ${r.interval} ${unit}`;
    if (r.frequency === "weekly")
      text +=
        " " +
        [...r.weekdays]
          .sort((a, b) => ((a + 6) % 7) - ((b + 6) % 7))
          .map((n) => "周" + "日一二三四五六"[n])
          .join("、");
    if (r.frequency === "monthly") text += ` ${r.monthDay} 日`;
    return text + (r.until ? ` · 至 ${r.until}` : " · 持续重复");
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
    showRecurring,
    setFilter,
    reorderRoot,
    repeatSummary,
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
