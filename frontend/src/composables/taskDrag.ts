import { runMotion } from "./motion";
export interface TaskProjection {
  parentId: string | null;
  targetId?: string;
  after: boolean;
  depth: number;
}
export function taskDropProjection(
  event: DragEvent,
): TaskProjection | undefined {
  return (event as DragEvent & { todoProjection?: TaskProjection })
    .todoProjection;
}
type Placement = "before" | "after" | "inside";
export function taskDropPlacement(event: DragEvent): Placement | undefined {
  return (event as DragEvent & { todoPlacement?: Placement }).todoPlacement;
}
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
  const source = handle.closest<HTMLElement>(".task-row,.calendar-task")!;
  const bounds = source.getBoundingClientRect();
  const origin = { x: event.clientX, y: event.clientY };
  let x = origin.x,
    y = origin.y,
    alt = event.altKey,
    ghost: HTMLElement | undefined,
    target: HTMLElement | null = null,
    gap: HTMLElement | undefined,
    projection: TaskProjection | undefined,
    placement: Placement | undefined,
    frame = 0,
    started = false,
    edgeDirection = 0,
    edgeSince = 0,
    lastPage = 0,
    ending = false;
  event.preventDefault();
  handle.setPointerCapture(event.pointerId);
  function clearGap() {
    gap?.remove();
    gap = undefined;
  }
  function projectList(list: HTMLElement) {
    const all = Array.from(list.querySelectorAll<HTMLElement>("[data-node]"));
    const sourceAt = all.indexOf(source),
      sourceDepth = Number(source.style.getPropertyValue("--depth"));
    const excluded = new Set<HTMLElement>([source]);
    for (const row of all.slice(sourceAt + 1)) {
      if (Number(row.style.getPropertyValue("--depth")) <= sourceDepth) break;
      excluded.add(row);
    }
    const rows = all.filter((row) => !excluded.has(row));
    if (!rows.length) {
      projection = undefined;
      clearGap();
      return;
    }
    let row =
      rows.find((row) => y < row.getBoundingClientRect().bottom) ||
      rows.at(-1)!;
    const rect = row.getBoundingClientRect();
    const first = rows[0],
      last = rows.at(-1)!;
    const atStart = y <= first.getBoundingClientRect().top + 5;
    const atEnd = y >= last.getBoundingClientRect().bottom - 5;
    let after = y > rect.top + rect.height / 2;
    let depth = Math.max(
      0,
      Math.min(
        Math.max(0, 5 - Number(source.dataset.height || 1)),
        sourceDepth + Math.trunc((x - origin.x) / 22),
      ),
    );
    if (alt)
      depth = Math.min(depth, Number(row.style.getPropertyValue("--depth")));
    if (atStart || atEnd) {
      depth = 0;
      row = atStart
        ? first
        : rows
            .filter((r) => Number(r.style.getPropertyValue("--depth")) === 0)
            .at(-1) || last;
      after = atEnd;
    }
    const rowDepth = Number(row.style.getPropertyValue("--depth"));
    depth = Math.min(depth, rowDepth + 1);
    let anchor = row,
      lineY = 0;
    if (depth > rowDepth) {
      projection = { parentId: row.dataset.node!, after: true, depth };
      for (const next of rows.slice(rows.indexOf(row) + 1)) {
        if (Number(next.style.getPropertyValue("--depth")) <= rowDepth) break;
        anchor = next;
      }
      lineY = anchor.getBoundingClientRect().bottom;
    } else {
      while (Number(anchor.style.getPropertyValue("--depth")) > depth) {
        const parent = rows.find(
          (r) => r.dataset.node === anchor.dataset.parent,
        );
        if (!parent) break;
        anchor = parent;
      }
      projection = {
        parentId: anchor.dataset.parent || null,
        targetId: anchor.dataset.node,
        after,
        depth,
      };
      if (after) {
        const anchorDepth = Number(anchor.style.getPropertyValue("--depth"));
        for (const next of rows.slice(rows.indexOf(anchor) + 1)) {
          if (Number(next.style.getPropertyValue("--depth")) <= anchorDepth)
            break;
          anchor = next;
        }
      }
      lineY = after
        ? anchor.getBoundingClientRect().bottom
        : anchor.getBoundingClientRect().top;
    }
    if (list.hasAttribute("data-completed-bottom") && projection.targetId) {
      const projectedRow = rows.find(
        (r) => r.dataset.node === projection!.targetId,
      );
      if (projectedRow?.dataset.status === "completed") {
        const peers = rows.filter(
          (r) => (r.dataset.parent || null) === projection!.parentId,
        );
        const pending = peers.filter((r) => r.dataset.status !== "completed");
        const lastPending = pending.at(-1);
        if (lastPending) {
          projection.targetId = lastPending.dataset.node;
          projection.after = true;
          after = true;
          anchor = lastPending;
          const d = Number(anchor.style.getPropertyValue("--depth"));
          for (const next of rows.slice(rows.indexOf(anchor) + 1)) {
            if (Number(next.style.getPropertyValue("--depth")) <= d) break;
            anchor = next;
          }
          lineY = anchor.getBoundingClientRect().bottom;
        } else if (peers[0]) {
          anchor = peers[0];
          projection.targetId = anchor.dataset.node;
          projection.after = false;
          after = false;
          lineY = anchor.getBoundingClientRect().top;
        }
      }
    }
    if (!gap) {
      gap = document.createElement("div");
      gap.className = "task-drop-line";
      gap.setAttribute("aria-hidden", "true");
      document.body.append(gap);
    }
    const base =
      rows.find((r) => Number(r.style.getPropertyValue("--depth")) === 0) ||
      first;
    const left = base.getBoundingClientRect().left + 31 + depth * 22;
    Object.assign(gap.style, {
      left: left + "px",
      top: lineY + "px",
      width:
        Math.max(30, list.getBoundingClientRect().right - left - 12) + "px",
    });
    gap.dataset.depth = String(depth + 1);
    gap.dataset.anchor = anchor.dataset.node;
    gap.dataset.parent = projection.parentId || "";
    gap.dataset.position = after ? "after" : "before";
  }
  function clearTarget() {
    target?.removeAttribute("data-drop-placement");
    target?.removeAttribute("data-calendar-drop");
    target?.removeAttribute("data-tail-drop");
    target = null;
    placement = undefined;
    projection = undefined;
    clearGap();
  }
  function draw() {
    if (!started) return;
    ghost!.style.transform = `translate3d(${x - origin.x + bounds.left}px,${y - origin.y + bounds.top}px,0) scale(1.015)`;
    const under = document.elementFromPoint(x, y);
    const list = source.classList.contains("task-row")
      ? under?.closest<HTMLElement>(".task-list")
      : null;
    const hit =
      list ||
      under?.closest<HTMLElement>(".week-date,.calendar-day,.view-switch") ||
      null;
    if (target !== hit) {
      clearTarget();
      target = hit;
    }
    if (list) projectList(list);
    if (target?.matches(".calendar-day,.week-date"))
      target.setAttribute("data-calendar-drop", "");
    // Dwell at either horizontal edge to page without dropping the pointer preview.
    const grid = document.querySelector<HTMLElement>(".calendar-grid");
    if (source.classList.contains("calendar-task") && grid) {
      const r = grid.getBoundingClientRect();
      const direction =
        y >= r.top && y <= r.bottom
          ? x <= r.left + 28
            ? -1
            : x >= r.right - 28
              ? 1
              : 0
          : 0;
      const now = performance.now();
      if (direction !== edgeDirection) {
        edgeDirection = direction;
        edgeSince = now;
      }
      if (
        !ending &&
        direction &&
        now - edgeSince > 650 &&
        now - lastPage > 950
      ) {
        lastPage = now;
        clearTarget();
        document.dispatchEvent(
          new CustomEvent("todo:calendar-step", { detail: direction }),
        );
      }
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
      ghost.removeAttribute("data-calendar-task");
      ghost.removeAttribute("data-pointer-source");
      ghost.removeAttribute("tabindex");
      ghost.setAttribute("aria-hidden", "true");
      ghost
        .querySelectorAll("button,input,textarea")
        .forEach((el) => el.setAttribute("tabindex", "-1"));
      ghost.className = source.classList.contains("calendar-task")
        ? "calendar-task calendar-drag-preview pointer-drag-preview"
        : "task-row pointer-drag-preview";
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
    document.removeEventListener("keydown", key, true);
    window.removeEventListener("blur", abort);
    if (handle.hasPointerCapture(event.pointerId))
      handle.releasePointerCapture(event.pointerId);
    source.removeAttribute("data-pointer-source");
    document.documentElement.classList.remove("task-dragging");
    const destination = target;
    const chosenPlacement = placement;
    const chosenProjection = projection;
    clearTarget();
    cancelCurrent = undefined;
    if (started) {
      // A pointer drag ends in a click in WebViews; do not open the dragged task.
      const suppress = (event: MouseEvent) => {
        event.preventDefault();
        event.stopImmediatePropagation();
      };
      document.addEventListener("click", suppress, {
        capture: true,
        once: true,
      });
      setTimeout(
        () => document.removeEventListener("click", suppress, true),
        0,
      );
    }
    if (commit && started && destination) {
      const data = new DataTransfer();
      data.setData("text/plain", id);
      const drop = new DragEvent("drop", {
        bubbles: true,
        cancelable: true,
        dataTransfer: data,
        clientX: x,
        clientY: y,
        altKey: alt,
      });
      Object.defineProperty(drop, "todoPlacement", { value: chosenPlacement });
      Object.defineProperty(drop, "todoProjection", {
        value: chosenProjection,
      });
      destination.dispatchEvent(drop);
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
        ending = true;
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
  document.addEventListener("keydown", key, true);
  window.addEventListener("blur", abort);
}
