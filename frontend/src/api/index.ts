export const previewMode =
  new URLSearchParams(location.search).get("preview") === "1";
let pendingConfirmation: Promise<boolean> | undefined;
// WKWebView does not implement browser confirm panels in the desktop shell.
// Await the owning window's native dialog before performing a destructive action.
export async function confirmAction(message: string): Promise<boolean> {
  if (previewMode) return window.confirm(message);
  // A second action must not reuse approval of the already visible question.
  if (pendingConfirmation) return false;
  const key = new URLSearchParams(location.search).get("window") || "main";
  return (pendingConfirmation = api<boolean>("Confirm", key, message)
    .catch((error) => {
      window.dispatchEvent(
        new CustomEvent("todo:confirm-error", {
          detail: `无法显示确认窗口，请重试：${error instanceof Error ? error.message : String(error)}`,
        }),
      );
      return false;
    })
    .finally(() => {
      pendingConfirmation = undefined;
    }));
}
export async function api<T = void>(
  method: string,
  ...args: unknown[]
): Promise<T> {
  if (previewMode) {
    if (
      [
        "SetView",
        "WindowAction",
        "SetEdgeBusy",
        "SetEdgeEditing",
        "SetEdgeMenuBusy",
        "SetEdgeModalBusy",
      ].includes(method)
    )
      return undefined as T;
    if (method === "ShowWindow") {
      window.open(`${location.origin}/?preview=1&window=${args[0]}`, "_blank");
      return undefined as T;
    }
    if (method === "ClipboardWrite") {
      await navigator.clipboard.writeText(String(args[0]));
      return undefined as T;
    }
    if (method === "ClipboardRead")
      return (await navigator.clipboard.readText()) as T;
    const r = await fetch(`http://127.0.0.1:9246/api/${method}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(args),
    });
    const body = await r.json();
    if (!r.ok) throw new Error(body.error || "操作失败");
    return body as T;
  }
  const { Call } = await import("@wailsio/runtime");
  return Call.ByName(`main.TodoService.${method}`, ...args);
}
export async function listen(
  name: string,
  fn: (data: any) => void,
): Promise<() => void> {
  if (previewMode) return () => {};
  const { Events } = await import("@wailsio/runtime");
  return Events.On(name, (e) => fn(e.data));
}
