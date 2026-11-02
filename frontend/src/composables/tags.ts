import type { Tag } from "../api/types";

export const tagColors = [
  { name: "灰色", value: "#64748b" },
  { name: "红色", value: "#b93838" },
  { name: "橙色", value: "#b45309" },
  { name: "金色", value: "#8a650b" },
  { name: "绿色", value: "#23734b" },
  { name: "青色", value: "#0f766e" },
  { name: "蓝色", value: "#3563c1" },
  { name: "紫色", value: "#7851ad" },
] as const;

export function tagStyle(tag: Tag) {
  const color = tag.color || tagColors[0].value;
  const channels = [1, 3, 5].map((i) => {
    const value = parseInt(color.slice(i, i + 2), 16) / 255;
    return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
  });
  const luminance =
    channels[0] * 0.2126 + channels[1] * 0.7152 + channels[2] * 0.0722;
  return {
    "--tag-color": color,
    "--tag-ink": 1.05 / (luminance + 0.05) >= 4.5 ? "#ffffff" : "#101014",
  };
}
