<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import {
  type Confirmation,
  resolveConfirmation,
} from "../composables/confirmation";
import { reveal, leave, cancelMotion } from "../composables/motion";
import { restoreDialogFocus } from "../composables/dialogFocus";
import { api } from "../api";
const props = defineProps<{ request: Confirmation }>();
const dialog = ref<HTMLDialogElement>(),
  cancel = ref<HTMLButtonElement>();
let closing = false;
const key = new URLSearchParams(location.search).get("window") || "main";
restoreDialogFocus();
onMounted(() => {
  dialog.value?.showModal();
  if (dialog.value) reveal(dialog.value, "dialog");
  cancel.value?.focus();
  void api("SetConfirmationBusy", key, true).catch(() => {});
});
onUnmounted(() => {
  if (dialog.value) cancelMotion(dialog.value);
  resolveConfirmation(props.request.id, false);
  void api("SetConfirmationBusy", key, false).catch(() => {});
});
function finish(ok: boolean) {
  if (closing) return;
  closing = true;
  if (dialog.value)
    leave(dialog.value, "dialog", () =>
      resolveConfirmation(props.request.id, ok),
    );
  else resolveConfirmation(props.request.id, ok);
}
</script>
<template>
  <Teleport to="body"
    ><dialog
      ref="dialog"
      class="component-confirmation"
      aria-modal="true"
      aria-labelledby="confirmation-title"
      aria-describedby="confirmation-message"
      @cancel.prevent="finish(false)"
      @keydown.esc.stop.prevent="finish(false)"
      @keydown.meta.enter.stop.prevent
      @keydown.ctrl.enter.stop.prevent
    >
      <h2 id="confirmation-title">请确认</h2>
      <p id="confirmation-message">{{ request.message }}</p>
      <footer class="dialog-footer">
        <button ref="cancel" class="button" @click="finish(false)">取消</button
        ><button class="button primary" @click="finish(true)">确认</button>
      </footer>
    </dialog></Teleport
  >
</template>
