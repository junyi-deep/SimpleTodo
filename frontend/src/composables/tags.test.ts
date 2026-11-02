import { expect, it } from "vitest";
import { tagColors, tagStyle } from "./tags";

it("标签填充色为预设深色或导入浅色选择可读文字", () => {
  for (const color of tagColors) {
    expect(
      tagStyle({ id: "t", name: "标签", order: 0, color: color.value })[
        "--tag-ink"
      ],
    ).toBe("#ffffff");
  }
  expect(
    tagStyle({ id: "t", name: "浅色", order: 0, color: "#ffffff" })[
      "--tag-ink"
    ],
  ).toBe("#101014");
  expect(tagStyle({ id: "t", name: "旧标签", order: 0 })["--tag-color"]).toBe(
    "#64748b",
  );
});
