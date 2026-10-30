<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api } from "../api";
import { useTodo } from "../stores/todo";
import {
  shiftDay,
  type LocalCalendarState,
  type LocalCalendarSource,
} from "../api/types";
const s = useTodo(),
  state = ref<LocalCalendarState>({ sources: [], lastSync: 0, events: [] }),
  sources = ref<LocalCalendarSource[]>([]),
  selected = ref<string[]>([]),
  busy = ref(false),
  error = ref(""),
  loaded = ref(false);
const from = () => s.boot?.today || s.date;
onMounted(async () => {
  try {
    state.value = await api("LocalCalendarState", from(), shiftDay(from(), 42));
    selected.value = [...state.value.sources];
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e);
  }
});
async function discover() {
  busy.value = true;
  error.value = "";
  try {
    sources.value = await api("LocalCalendarSources");
    loaded.value = true;
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e);
  } finally {
    busy.value = false;
  }
}
async function save(ids: string[]) {
  busy.value = true;
  error.value = "";
  try {
    state.value = await api(
      "ConfigureLocalCalendar",
      ids,
      from(),
      shiftDay(from(), 42),
    );
    selected.value = [...state.value.sources];
    document.dispatchEvent(new Event("todo:local-calendar-changed"));
    s.inform(
      ids.length ? "本机日历已启用，只读显示日程" : "已停止显示本机日历",
    );
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e);
  } finally {
    busy.value = false;
  }
}
</script>
<template>
  <section class="local-calendar-settings" aria-label="本机 Outlook 日程">
    <h3>本机 Outlook 日程</h3>
    <p class="muted small">只读显示本机日程，不修改原日程，也不会上传待办。</p>
    <p v-if="s.boot?.capabilities.platform === 'darwin'" class="muted small">
      读取系统日历中的账户。请先在 macOS 系统设置中添加并启用 Outlook / Exchange
      日历；仅存在于 Outlook
      应用内的日历无法读取。系统会要求日历访问权限，本应用只使用读取操作。
    </p>
    <p v-else class="muted small">
      需要已配置账户的经典 Outlook。新版 Outlook 不支持本地日历接口。
    </p>
    <div class="inline-actions">
      <button class="button" :disabled="busy" @click="discover">
        {{ busy ? "读取中…" : "选择本机日历" }}</button
      ><button
        v-if="state.sources.length"
        class="button"
        :disabled="busy"
        @click="save([])"
      >
        停止显示
      </button>
    </div>
    <div v-if="loaded" class="local-calendar-choices">
      <p v-if="!sources.length" class="muted small">
        未找到本机日历，请检查系统账户和日历访问权限。
      </p>
      <label v-for="source in sources" :key="source.id"
        ><input
          v-model="selected"
          type="checkbox"
          :value="source.id"
          :disabled="busy"
        />{{ source.name
        }}<small class="muted">{{ source.source }}</small></label
      >
      <button
        v-if="sources.length"
        class="button primary"
        :disabled="busy"
        @click="save(selected)"
      >
        保存并读取日程
      </button>
    </div>
    <small v-if="state.lastSync" class="muted"
      >上次读取：{{ new Date(state.lastSync).toLocaleString() }}</small
    >
    <p v-if="error" class="error" role="alert">{{ error }}</p>
  </section>
</template>
