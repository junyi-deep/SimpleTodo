import { shallowRef } from "vue";
export interface Confirmation {
  id: number;
  message: string;
}
export const confirmation = shallowRef<Confirmation | null>(null);
let sequence = 0;
let answer: ((ok: boolean) => void) | undefined;
export function requestConfirmation(message: string): Promise<boolean> {
  if (confirmation.value || !message.trim()) return Promise.resolve(false);
  return new Promise((resolve) => {
    answer = resolve;
    confirmation.value = { id: ++sequence, message };
  });
}
export function resolveConfirmation(id: number, ok: boolean) {
  if (confirmation.value?.id !== id) return;
  const complete = answer;
  answer = undefined;
  confirmation.value = null;
  complete?.(ok);
}
