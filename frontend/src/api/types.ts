export interface Task {
  id: string;
  parentId: string | null;
  rootId: string;
  title: string;
  description: string | null;
  date: string | null;
  originalDate: string | null;
  status: "pending" | "completed" | "skipped";
  color: string | null;
  order: number;
  completedAt: number | null;
  postpone: string | null;
  seriesId: string | null;
  ruleId: string | null;
  occurrenceDate: string | null;
  exception: boolean;
  remindAt: number | null;
  reminderTime: string | null;
  reminderZone: string | null;
  reminderGeneration: number;
  version: string;
  createdAt: number;
  updatedAt: number;
}
export interface Query {
  view: string;
  date: string;
  end: string;
  limit: number;
}
export interface Snapshot {
  tasks: Task[];
  counts: Record<string, number>;
  revision: string;
  today: string;
  truncated: boolean;
}
export interface Focus {
  id: string;
  taskId: string | null;
  title: string;
  duration: number;
  remaining: number;
  deadline: number;
  state: string;
  startedAt: number;
  finishedAt: number | null;
}
export interface Countdown {
  id: string;
  name: string;
  allDay: boolean;
  anchorDate: string;
  anchorTime: string;
  target: number;
  zone: string;
  frequency: string;
  enabled: boolean;
  offset: number;
  generation: number;
}
export interface WindowState {
  key: string;
  x: number;
  y: number;
  width: number;
  height: number;
  mode: string;
  opacity: number;
  collapsed: boolean;
}
export interface Notice {
  id: string;
  title: string;
  fireAt: number;
  state: string;
  sourceId: string;
  sourceType: string;
}
export interface Boot {
  timezone?: string;
  startupError?: string;
  settings: Record<string, string>;
  version: string;
  dataDir: string;
  today: string;
  focus: Focus | null;
  countdowns: Countdown[];
  windows: WindowState[];
  notices: Notice[];
  warnings: string[];
  autostart: boolean;
  notificationAllowed: boolean;
  capabilities: { desktopMode: boolean; edgeHide: boolean; platform: string };
}
export interface Rule {
  id: string;
  seriesId: string;
  from: string;
  until: string | null;
  anchor: string;
  frequency: string;
  interval: number;
  weekdays: number[];
  monthDay: number;
  template: Task[];
  generatedThrough: string;
}
export interface Preview {
  token: string;
  tasks: Task[];
  roots: number;
  nodes: number;
  warnings: string[];
  replacing: boolean;
}
export interface Backup {
  name: string;
  size: number;
  createdAt: number;
}
export interface Holiday {
  date: string;
  name: string;
  kind: string;
  source: string;
  version: string;
}
export const dayString = (d: Date) =>
  `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
export const parseDay = (s: string) => new Date(`${s}T12:00:00`);
export const shiftDay = (s: string, n: number) => {
  const d = parseDay(s);
  d.setDate(d.getDate() + n);
  return dayString(d);
};
export const colors: Record<string, string> = {
  none: "transparent",
  red: "#c45353",
  orange: "#c7823c",
  yellow: "#b69b30",
  green: "#528369",
  blue: "#5c87bb",
  purple: "#9377b6",
};
export const colorNames: Record<string, string> = {
  none: "无色",
  red: "红色",
  orange: "橙色",
  yellow: "黄色",
  green: "绿色",
  blue: "蓝色",
  purple: "紫色",
};
