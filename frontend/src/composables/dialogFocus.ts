import { onUnmounted } from "vue";

export function restoreDialogFocus(taskId?: string) {
  const before = document.activeElement as HTMLElement | null;
  onUnmounted(() => {
    if (before?.isConnected) before.focus();
    else if (taskId)
      document.querySelector<HTMLElement>(`[data-node="${taskId}"]`)?.focus();
    else
      document.querySelector<HTMLElement>('[aria-label="工具菜单"]')?.focus();
  });
}
