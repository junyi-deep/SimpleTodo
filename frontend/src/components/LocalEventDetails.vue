<script setup lang="ts">
import { ref, onMounted } from "vue";
import { X } from "lucide-vue-next";
import { restoreDialogFocus } from "../composables/dialogFocus";
import type { LocalCalendarEvent } from "../api/types";
defineProps<{ event: LocalCalendarEvent }>();
const emit = defineEmits<{ close: [] }>();
const dialog = ref<HTMLDialogElement>();
restoreDialogFocus();
onMounted(() => {
  dialog.value?.showModal();
  dialog.value?.querySelector<HTMLButtonElement>("button")?.focus();
});
</script>
<template>
  <Teleport to="body"
    ><dialog
      ref="dialog"
      v-motion
      class="task-settings-dialog local-event-details"
      aria-label="本机日程详情"
      @pointerdown.stop
      @cancel.prevent="emit('close')"
      @keydown.esc.stop.prevent="emit('close')"
    >
      <header class="dialog-header">
        <strong>本机日程 · 只读</strong
        ><button
          class="icon-button"
          aria-label="关闭本机日程"
          @click="emit('close')"
        >
          <X :size="16" />
        </button>
      </header>
      <div class="dialog-body">
        <h2>{{ event.title }}</h2>
        <p>
          {{ event.allDay ? "全天" : new Date(event.start).toLocaleString()
          }}<template v-if="!event.allDay">
            — {{ new Date(event.end).toLocaleString() }}</template
          >
        </p>
        <p v-if="event.location">{{ event.location }}</p>
        <small class="muted"
          >从本机日历读取；如需修改，请在 Outlook 或系统日历中操作。</small
        >
      </div>
    </dialog></Teleport
  >
</template>
