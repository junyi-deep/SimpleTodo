<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { X } from "lucide-vue-next";
import { api, listen } from "../api";
import { restoreDialogFocus } from "../composables/dialogFocus";
const props = defineProps<{ info: { version: string; releaseURL: string } }>();
const emit = defineEmits<{ close: [] }>();
const dialog = ref<HTMLDialogElement>(),
  busy = ref(false),
  progress = ref(""),
  error = ref("");
let stop: (() => void) | undefined;
restoreDialogFocus();
onMounted(async () => {
  dialog.value?.showModal();
  stop = await listen(
    "todo:update-progress",
    (v) => (progress.value = String(v)),
  );
});
onUnmounted(() => stop?.());
async function skip() {
  try {
    await api("SkipUpdate", props.info.version);
    emit("close");
  } catch (e) {
    error.value = String(e);
  }
}
async function install() {
  busy.value = true;
  error.value = "";
  progress.value = "正在下载安装包…";
  try {
    await api("InstallUpdate");
    progress.value = "正在安装，即将重启…";
  } catch (e) {
    error.value = String(e);
    progress.value = "";
    busy.value = false;
  }
}
</script>
<template>
  <Teleport to="body"
    ><dialog
      ref="dialog"
      v-motion
      class="component-confirmation update-prompt"
      aria-label="发现新版本"
      @cancel.prevent="!busy && emit('close')"
      @keydown.esc.stop.prevent="!busy && emit('close')"
    >
      <header class="dialog-header">
        <h2>发现新版本 {{ info.version }}</h2>
        <button
          class="icon-button"
          aria-label="稍后更新"
          :disabled="busy"
          @click="emit('close')"
        >
          <X :size="16" />
        </button>
      </header>
      <p>
        立即更新会下载、校验并替换应用，然后重启。任务数据保留，进行中的专注与提醒会短暂停止。
      </p>
      <p v-if="progress" class="muted small" role="status">{{ progress }}</p>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <footer class="dialog-footer">
        <button class="button" :disabled="busy" @click="skip">跳过此版本</button
        ><button class="button primary" :disabled="busy" @click="install">
          {{ busy ? "正在更新…" : "立即更新" }}
        </button>
      </footer>
    </dialog></Teleport
  >
</template>
