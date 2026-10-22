<script setup lang="ts">
import CustomSelect from "./CustomSelect.vue";
import { ref, computed, onMounted, onUnmounted } from "vue";
import {
  Play,
  Pause,
  Square,
  Plus,
  Trash2,
  Timer,
  Clock3,
  Bell,
} from "lucide-vue-next";
import { api, confirmAction } from "../api";
import { type Countdown } from "../api/types";
import { useTodo } from "../stores/todo";
const s = useTodo(),
  mode = ref("focus"),
  minutes = ref(Number(s.boot?.settings.focusMinutes || 25)),
  taskID = ref(""),
  now = ref(Date.now()),
  editing = ref(false),
  name = ref(""),
  target = ref(""),
  allDay = ref(false),
  frequency = ref("none"),
  offset = ref(0),
  editID = ref("");
let tick: ReturnType<typeof setInterval>,
  meta = 0;
onMounted(() => {
  tick = setInterval(() => {
    now.value = Date.now();
    meta++;
    if (meta % 5 === 0) void s.metadata();
  }, 1000);
});
onUnmounted(() => clearInterval(tick));
const focus = computed(() => s.boot?.focus),
  remaining = computed(() =>
    focus.value?.state === "running"
      ? Math.max(0, Math.ceil((focus.value.deadline - now.value) / 1000))
      : focus.value?.state === "paused"
        ? focus.value.remaining
        : minutes.value * 60,
  );
const time = computed(
  () =>
    `${String(Math.floor(remaining.value / 60)).padStart(2, "0")}:${String(remaining.value % 60).padStart(2, "0")}`,
);
const active = computed(() =>
  ["running", "paused"].includes(focus.value?.state || ""),
);
async function action(kind: string) {
  if (kind === "stop" && !(await confirmAction("结束这次专注？"))) return;
  await s.run(() => api("Focus", kind, minutes.value, taskID.value || null));
  await s.metadata();
  if (kind === "start") await s.requestNotifications();
}
async function save() {
  const ms = new Date(
    target.value + (allDay.value ? "T00:00:00" : ""),
  ).getTime();
  if (!name.value.trim() || !Number.isFinite(ms)) {
    s.error = "请填写事项名称和目标时间";
    return;
  }
  await s.run(() =>
    api("SaveCountdown", {
      id: editID.value,
      name: name.value,
      target: ms,
      allDay: allDay.value,
      frequency: frequency.value,
      offset: offset.value,
      enabled: true,
    }),
  );
  if (!s.error) {
    editing.value = false;
    await s.metadata();
    await s.requestNotifications();
  }
}
function edit(c?: Countdown) {
  editID.value = c?.id || "";
  name.value = c?.name || "";
  allDay.value = c?.allDay || false;
  frequency.value = c?.frequency || "none";
  offset.value = c?.offset || 0;
  if (c) {
    const d = new Date(c.target);
    d.setMinutes(d.getMinutes() - d.getTimezoneOffset());
    target.value = d.toISOString().slice(0, c.allDay ? 10 : 16);
  } else target.value = "";
  editing.value = true;
}
function left(c: Countdown) {
  const diff = c.target - now.value;
  if (diff <= 0) return "已到期";
  if (c.allDay) {
    const d = new Date(c.target);
    d.setHours(0, 0, 0, 0);
    const today = new Date(now.value);
    today.setHours(0, 0, 0, 0);
    return `${Math.round((d.getTime() - today.getTime()) / 86400000)} 天`;
  }
  const sec = Math.floor(diff / 1000),
    days = Math.floor(sec / 86400);
  return `${days ? days + " 天 " : ""}${String(Math.floor((sec % 86400) / 3600)).padStart(2, "0")}:${String(Math.floor((sec % 3600) / 60)).padStart(2, "0")}:${String(sec % 60).padStart(2, "0")}`;
}
async function remove(c: Countdown) {
  if (await confirmAction(`删除倒计时「${c.name}」？`)) {
    await s.run(() => api("DeleteCountdown", c.id));
    await s.metadata();
  }
}
async function enabled(c: Countdown) {
  await s.run(() => api("ToggleCountdown", c.id, !c.enabled));
  await s.metadata();
}
</script>
<template>
  <section v-motion="mode" class="focus-page">
    <div class="segmented">
      <button :class="{ active: mode === 'focus' }" @click="mode = 'focus'">
        <Timer :size="15" />专注时钟</button
      ><button
        :class="{ active: mode === 'countdown' }"
        @click="mode = 'countdown'"
      >
        <Clock3 :size="15" />事项倒计时
      </button>
    </div>
    <template v-if="mode === 'focus'"
      ><div
        class="focus-orbit"
        :class="{ running: focus?.state === 'running' }"
      >
        <div class="focus-face" aria-hidden="true"><i /><i /><span /></div>
      </div>
      <p class="focus-state">
        {{
          focus?.state === "running"
            ? "留一点时间，专心做好这一件事"
            : focus?.state === "paused"
              ? "慢一点没关系，准备好再继续"
              : focus?.state === "completed"
                ? "这一段专注已完成，做得不错"
                : "把注意力留给当下"
        }}
      </p>
      <div class="clock-digits" role="timer" aria-live="off">{{ time }}</div>
      <p class="muted small">
        {{ active ? focus?.title : "开始一段不被打扰的时间" }}
      </p>
      <div v-if="!active" class="focus-config">
        <label
          >专注时长<input
            v-model.number="minutes"
            type="number"
            min="1"
            max="180"
          />分钟</label
        ><CustomSelect v-model="taskID" aria-label="关联任务">
          <option value="">自由专注，不关联任务</option>
          <option v-for="task in s.roots" :key="task.id" :value="task.id">
            {{ task.title }}
          </option>
        </CustomSelect>
      </div>
      <div class="focus-actions">
        <button
          v-if="!active"
          class="button primary large"
          @click="action('start')"
        >
          <Play :size="16" />开始专注</button
        ><template v-else
          ><button
            class="button primary"
            @click="action(focus?.state === 'running' ? 'pause' : 'resume')"
          >
            <Pause v-if="focus?.state === 'running'" :size="16" /><Play
              v-else
              :size="16"
            />{{ focus?.state === "running" ? "暂停" : "继续" }}</button
          ><button class="button" @click="action('stop')">
            <Square :size="14" />结束
          </button></template
        >
      </div>
      <p class="quiet-note">
        关闭面板后仍会计时 · 完全退出后下次启动恢复状态
      </p></template
    >
    <template v-else
      ><div class="section-heading">
        <h2>值得期待的事</h2>
        <button class="button" @click="edit()"><Plus :size="15" />新建</button>
      </div>
      <div v-if="!s.boot?.countdowns.length && !editing" class="empty-state">
        <Clock3 :size="30" />
        <h3>为重要的日子留个位置</h3>
        <p>设置目标时间和提醒，一切都在本机保存。</p>
        <button class="text-button" @click="edit()">创建第一个倒计时</button>
      </div>
      <div v-for="c in s.boot?.countdowns" :key="c.id" class="countdown-row">
        <button class="countdown-main" @click="edit(c)">
          <span
            >{{ c.name
            }}<small
              >{{ new Date(c.target).toLocaleString("zh-CN") }} ·
              {{ c.frequency === "none" ? "单次" : "重复" }}</small
            ></span
          ><strong>{{ left(c) }}</strong></button
        ><button
          class="icon-button"
          :class="{ active: c.enabled }"
          :title="c.enabled ? '关闭此事项提醒' : '启用此事项提醒'"
          @click="enabled(c)"
        >
          <Bell :size="15" /></button
        ><button
          class="icon-button danger"
          aria-label="删除倒计时"
          @click="remove(c)"
        >
          <Trash2 :size="15" />
        </button>
      </div>
      <form v-if="editing" class="countdown-form" @submit.prevent="save">
        <label class="field"
          ><span>事项名称</span
          ><input
            v-model="name"
            required
            maxlength="500"
            placeholder="比如：下一次旅行" /></label
        ><label class="check-label"
          ><input
            v-model="allDay"
            type="checkbox"
            @change="target = ''"
          />按天显示</label
        ><label class="field"
          ><span>目标日期{{ allDay ? "" : "与时间" }}</span
          ><input
            v-model="target"
            :type="allDay ? 'date' : 'datetime-local'"
            required
        /></label>
        <div class="field-grid">
          <label class="field"
            ><span>重复</span
            ><CustomSelect v-model="frequency">
              <option value="none">不重复</option>
              <option value="daily">每日</option>
              <option value="weekly">每周</option>
              <option value="monthly">每月</option>
            </CustomSelect></label
          ><label class="field"
            ><span>提醒</span
            ><CustomSelect v-model.number="offset">
              <option :value="0">到点</option>
              <option :value="300">提前 5 分钟</option>
              <option :value="900">提前 15 分钟</option>
              <option :value="3600">提前 1 小时</option>
              <option :value="86400">提前 1 天</option>
            </CustomSelect></label
          >
        </div>
        <footer class="inline-actions">
          <button class="button" type="button" @click="editing = false">
            取消</button
          ><button class="button primary" :disabled="s.saving">保存事项</button>
        </footer>
      </form></template
    >
  </section>
</template>
