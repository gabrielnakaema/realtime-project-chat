import type { Locator } from "@playwright/test";

type RichTextListType = "bulleted" | "numbered";

export async function fillRichTextList(
  container: Locator,
  type: RichTextListType,
  items: string[]
) {
  if (items.length === 0) {
    throw new Error("A rich-text list needs at least one item");
  }

  const editor = container.locator("#description");
  const buttonName = type === "bulleted" ? "Bulleted List" : "Numbered List";

  await editor.click();
  await container.getByRole("button", { name: buttonName }).click();
  await editor.pressSequentially(items[0]);

  for (const item of items.slice(1)) {
    await editor.press("Enter");
    await editor.pressSequentially(item);
  }
}
