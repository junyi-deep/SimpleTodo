import { animate, type DOMKeyframesDefinition } from "motion";
import type { Directive } from "vue";

type Kind = "pane" | "popover" | "drawer" | "row" | "toast" | "dialog";
const active = new Map<Element, { cancel: () => void; complete: () => void }>();
const preference = matchMedia("(prefers-reduced-motion: reduce)");
let enabled = true;
export function setMotionEnabled(value: boolean) {
  enabled = value;
  if (!motionAllowed()) active.forEach((animation) => animation.complete());
}
preference.addEventListener("change", () => {
  if (!motionAllowed()) active.forEach((animation) => animation.complete());
});
export function motionAllowed() {
  return enabled && !preference.matches;
}
const ease = [0.16, 1, 0.3, 1] as const;
function run(
  el: Element,
  frames: DOMKeyframesDefinition,
  duration: number,
  done = () => {},
) {
  active.get(el)?.cancel();
  if (!motionAllowed()) {
    done();
    return;
  }
  const control = animate(el, frames, { duration, ease: [...ease] });
  const finish = () => {
    if (active.get(el) !== state) return;
    control.cancel(); // Release inline/WAAPI styles; CSS remains the source of truth.
    active.delete(el);
    done();
  };
  const state = {
    cancel: () => {
      control.cancel();
      active.delete(el);
    },
    complete: finish,
  };
  active.set(el, state);
  control.then(finish);
}
export function reveal(el: Element, kind: Kind = "pane", done?: () => void) {
  restoreLeaving(el);
  const frames =
    kind === "drawer"
      ? { opacity: [0.3, 1], transform: ["translateX(28px)", "translateX(0)"] }
      : kind === "dialog"
        ? {
            opacity: [0.4, 1],
            transform: [
              "scale(.97) translateY(10px)",
              "scale(1) translateY(0)",
            ],
          }
        : kind === "popover"
          ? {
              opacity: [0, 1],
              transform: [
                "translateY(-5px) scale(.98)",
                "translateY(0) scale(1)",
              ],
            }
          : kind === "row"
            ? {
                opacity: [0, 1],
                clipPath: ["inset(0 100% 0 0)", "inset(0 0% 0 0)"],
              }
            : {
                opacity: [0.35, 1],
                transform: ["translateY(6px)", "translateY(0)"],
              };
  run(el, frames, kind === "drawer" ? 0.28 : 0.2, done);
}
export function leave(el: Element, kind: Kind = "pane", done?: () => void) {
  if (kind === "row") {
    el.setAttribute("data-motion-leaving", "");
    el.setAttribute("inert", "");
    el.setAttribute("aria-hidden", "true");
  }
  run(
    el,
    {
      opacity: [1, 0],
      ...(kind === "drawer"
        ? { transform: ["translateX(0)", "translateX(18px)"] }
        : {}),
    },
    0.12,
    done,
  );
}
function restoreLeaving(el: Element) {
  if (el.hasAttribute("data-motion-leaving")) {
    el.removeAttribute("data-motion-leaving");
    el.removeAttribute("inert");
    el.removeAttribute("aria-hidden");
  }
}
export function cancelMotion(el: Element) {
  active.get(el)?.cancel();
  restoreLeaving(el);
}
export const vMotion: Directive<
  HTMLElement,
  string | number | boolean | undefined
> = {
  mounted(el, binding) {
    reveal(el, (binding.arg || "pane") as Kind);
  },
  updated(el, binding) {
    if (binding.value !== binding.oldValue)
      reveal(el, (binding.arg || "pane") as Kind);
  },
  beforeUnmount: cancelMotion,
};
export const vComplete: Directive<HTMLElement, boolean> = {
  updated(el, binding) {
    if (binding.value && !binding.oldValue) {
      run(el, { transform: ["scale(1)", "scale(1.18)", "scale(1)"] }, 0.28);
      const path = el.querySelector("path");
      if (path && motionAllowed()) {
        const length = (path as SVGPathElement).getTotalLength();
        run(
          path,
          {
            strokeDasharray: [`${length} ${length}`, `${length} ${length}`],
            strokeDashoffset: [length, 0],
          },
          0.26,
        );
      }
    }
  },
  beforeUnmount(el) {
    cancelMotion(el);
    el.querySelectorAll("path").forEach(cancelMotion);
  },
};
// Animate the icon, never the control's hit box or native drag handle.
export function installControlMotion() {
  function press(event: PointerEvent) {
    const target = (event.target as Element).closest("button,summary");
    if (
      !target ||
      target.hasAttribute("disabled") ||
      target.classList.contains("task-check")
    )
      return;
    const icon = target.querySelector("svg");
    if (icon)
      run(icon, { transform: ["scale(1)", "scale(.82)", "scale(1)"] }, 0.2);
  }
  document.addEventListener("pointerdown", press);
  return () => {
    document.removeEventListener("pointerdown", press);
    active.forEach((a) => a.complete());
  };
}
