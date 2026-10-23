import { runMotion } from "./motion";
let cancelCurrent: (() => void) | undefined;
export function cancelTaskDrag() {
  cancelCurrent?.();
}
// Pointer-driven preview: the DOM stays Vue-owned until the existing transactional drop handler commits.
export function startTaskDrag(event: PointerEvent, id: string) {
  if (event.button !== 0 || (event.pointerType === "touch" && !event.isPrimary))
    return;
  cancelCurrent?.();
  const handle = event.currentTarget as HTMLElement;
  const source = handle.closest<HTMLElement>(".task-row")!;
  const bounds = source.getBoundingClientRect();
  const origin = { x: event.clientX, y: event.clientY };
  let x = origin.x,
    y = origin.y,
    alt = event.altKey,
    ghost: HTMLElement | undefined,
    target: HTMLElement | null = null,
    frame = 0,
    started = false;
  event.preventDefault();
  handle.setPointerCapture(event.pointerId);
  function clearTarget() {
    target?.removeAttribute("data-drop-placement");
    target = null;
  }
  function draw() {
    if (!started) return;
    ghost!.style.transform = `translate3d(${x - origin.x + bounds.left}px,${y - origin.y + bounds.top}px,0) scale(1.015)`;
    const hit =
      document
        .elementFromPoint(x, y)
        ?.closest<HTMLElement>(
          ".task-row,.week-date,.calendar-day,.view-switch",
        ) || null;
    if (target !== hit) {
      clearTarget();
      target = hit;
    }
    if (target === source) {
      clearTarget();
    }
    if (target?.classList.contains("task-row")) {
      const r = target.getBoundingClientRect(),
        fraction = (y - r.top) / r.height;
      target.dataset.dropPlacement =
        alt || fraction < 0.25
          ? "before"
          : fraction > 0.75
            ? "after"
            : "inside";
    }
    const scroll = document
      .elementFromPoint(x, y)
      ?.closest<HTMLElement>(".task-list");
    if (scroll) {
      const r = scroll.getBoundingClientRect();
      if (y < r.top + 44)
        scroll.scrollTop -= Math.min(14, (r.top + 44 - y) / 3);
      else if (y > r.bottom - 44)
        scroll.scrollTop += Math.min(14, (y - r.bottom + 44) / 3);
    }
    frame = requestAnimationFrame(draw);
  }
  function move(e: PointerEvent) {
    if (e.pointerId !== event.pointerId) return;
    x = e.clientX;
    y = e.clientY;
    alt = e.altKey;
    if (!started && Math.hypot(x - origin.x, y - origin.y) > 5) {
      started = true;
      source.setAttribute("data-pointer-source", "");
      document.documentElement.classList.add("task-dragging");
      ghost = source.cloneNode(true) as HTMLElement;
      ghost.removeAttribute("data-node");
      ghost.removeAttribute("data-pointer-source");
      ghost.removeAttribute("tabindex");
      ghost.setAttribute("aria-hidden", "true");
      ghost
        .querySelectorAll("button,input,textarea")
        .forEach((el) => el.setAttribute("tabindex", "-1"));
      ghost.className = "task-row pointer-drag-preview";
      Object.assign(ghost.style, {
        width: bounds.width + "px",
        height: bounds.height + "px",
        left: "0",
        top: "0",
      });
      document.body.append(ghost);
      frame = requestAnimationFrame(draw);
    }
    if (started) e.preventDefault();
  }
  function stop(commit = false) {
    cancelAnimationFrame(frame);
    document.removeEventListener("pointermove", move);
    document.removeEventListener("pointerup", up);
    document.removeEventListener("pointercancel", abort);
    document.removeEventListener("keydown", key);
    window.removeEventListener("blur", abort);
    if (handle.hasPointerCapture(event.pointerId))
      handle.releasePointerCapture(event.pointerId);
    source.removeAttribute("data-pointer-source");
    document.documentElement.classList.remove("task-dragging");
    const destination = target;
    clearTarget();
    cancelCurrent = undefined;
    if (commit && started && destination) {
      const data = new DataTransfer();
      data.setData("text/plain", id);
      destination.dispatchEvent(
        new DragEvent("drop", {
          bubbles: true,
          cancelable: true,
          dataTransfer: data,
          clientX: x,
          clientY: y,
          altKey: alt,
        }),
      );
    }
    const preview = ghost;
    if (preview) {
      const r = (
        commit && destination ? destination : source
      ).getBoundingClientRect();
      runMotion(
        preview,
        {
          transform: `translate3d(${r.left}px,${r.top}px,0) scale(1)`,
          opacity: 0,
        },
        0.18,
        () => preview.remove(),
      );
    }
  }
  function up(e: PointerEvent) {
    if (e.pointerId === event.pointerId) {
      x = e.clientX;
      y = e.clientY;
      alt = e.altKey;
      if (started) {
        cancelAnimationFrame(frame);
        draw();
      }
      stop(true);
    }
  }
  function abort() {
    stop();
  }
  function key(e: KeyboardEvent) {
    if (e.key === "Escape") {
      e.preventDefault();
      e.stopPropagation();
      stop();
    }
  }
  cancelCurrent = abort;
  document.addEventListener("pointermove", move, { passive: false });
  document.addEventListener("pointerup", up);
  document.addEventListener("pointercancel", abort);
  document.addEventListener("keydown", key);
  window.addEventListener("blur", abort);
}
