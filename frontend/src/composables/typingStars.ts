import type { Directive } from "vue";
import { motionAllowed } from "./motion";
type Field = HTMLInputElement | HTMLTextAreaElement;
type Star = {
  x: number;
  y: number;
  vx: number;
  vy: number;
  born: number;
  size: number;
  color: string;
};
const owners = new WeakMap<Field, () => void>();
const particles: Star[] = [];
let canvas: HTMLCanvasElement | undefined,
  frame = 0;
// Measure a mirror with identical wrapping so particles follow a multiline caret.
function caret(field: Field) {
  const box = field.getBoundingClientRect(),
    css = getComputedStyle(field);
  const mirror = document.createElement("div");
  for (const key of [
    "font",
    "letter-spacing",
    "line-height",
    "padding",
    "border",
    "box-sizing",
    "word-break",
    "overflow-wrap",
    "text-align",
  ])
    mirror.style.setProperty(key, css.getPropertyValue(key));
  Object.assign(mirror.style, {
    position: "fixed",
    visibility: "hidden",
    left: box.left + "px",
    top: box.top + "px",
    width: box.width + "px",
    whiteSpace: field instanceof HTMLTextAreaElement ? "pre-wrap" : "pre",
  });
  mirror.textContent = field.value.slice(
    0,
    field.selectionStart ?? field.value.length,
  );
  const marker = document.createElement("span");
  marker.textContent = "\u200b";
  mirror.append(marker);
  document.body.append(mirror);
  const r = marker.getBoundingClientRect();
  mirror.remove();
  return {
    x: Math.min(
      box.right - 3,
      Math.max(box.left + 3, r.left - field.scrollLeft),
    ),
    y: Math.min(
      box.bottom - 3,
      Math.max(box.top + 3, r.top + r.height / 2 - field.scrollTop),
    ),
  };
}
function draw(now: number) {
  const ctx = canvas?.getContext("2d");
  if (!ctx || !canvas) return;
  if (!motionAllowed()) particles.length = 0;
  ctx.clearRect(0, 0, innerWidth, innerHeight);
  for (let i = particles.length - 1; i >= 0; i--) {
    const p = particles[i],
      age = (now - p.born) / 480;
    if (age >= 1) {
      particles.splice(i, 1);
      continue;
    }
    ctx.save();
    ctx.translate(p.x + p.vx * age, p.y + p.vy * age + 12 * age * age);
    ctx.rotate(age * 1.3);
    ctx.globalAlpha = (1 - age) * 0.85;
    ctx.fillStyle = p.color;
    ctx.beginPath();
    for (let j = 0; j < 8; j++) {
      const angle = (j * Math.PI) / 4,
        radius = p.size * (j % 2 ? 0.25 : 1) * (1 - age * 0.5);
      ctx.lineTo(Math.cos(angle) * radius, Math.sin(angle) * radius);
    }
    ctx.closePath();
    ctx.fill();
    ctx.restore();
  }
  if (particles.length) frame = requestAnimationFrame(draw);
  else {
    canvas.remove();
    canvas = undefined;
    frame = 0;
  }
}
function burst(field: Field) {
  if (!motionAllowed() || !field.value || document.hidden) return;
  if (!canvas) {
    canvas = document.createElement("canvas");
    canvas.className = "typing-stars";
    canvas.setAttribute("aria-hidden", "true");
    const scale = Math.min(devicePixelRatio || 1, 2);
    canvas.width = innerWidth * scale;
    canvas.height = innerHeight * scale;
    canvas.style.width = innerWidth + "px";
    canvas.style.height = innerHeight + "px";
    canvas.getContext("2d")?.scale(scale, scale);
    document.body.append(canvas);
  }
  const point = caret(field),
    palette = ["#d39b27", "#579fe5", "#aa83d8"];
  for (let i = 0; i < 5; i++)
    particles.push({
      ...point,
      vx: (Math.random() - 0.5) * 56,
      vy: -12 - Math.random() * 28,
      born: performance.now(),
      size: 2 + Math.random() * 2.5,
      color: palette[i % palette.length],
    });
  if (particles.length > 80) particles.splice(0, particles.length - 80);
  if (!frame) frame = requestAnimationFrame(draw);
}
export const vSparks: Directive<Field> = {
  mounted(field) {
    const input = (event: Event) => {
      const e = event as InputEvent;
      if (!e.isComposing && e.inputType?.startsWith("insert")) burst(field);
    };
    const composition = () => burst(field);
    field.addEventListener("input", input);
    field.addEventListener("compositionend", composition);
    owners.set(field, () => {
      field.removeEventListener("input", input);
      field.removeEventListener("compositionend", composition);
    });
  },
  beforeUnmount(field) {
    owners.get(field)?.();
    owners.delete(field);
  },
};
