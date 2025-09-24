import { test, expect } from "@playwright/test";

test("homepage shows dashboard heading", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { level: 1, name: "VPN MVP Dashboard" })).toBeVisible();
});
