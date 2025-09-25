import { test, expect } from "@playwright/test";

const testKey = "gPJJ8GJv0FakePublicKeyForTests1234567=";

test("user can add a device and download config", async ({ page }) => {
  await page.goto("/devices");

  await page.getByLabel("Cihaz adı").fill("Test Device");
  await page.getByLabel("Public key").fill(testKey);
  await page.getByRole("button", { name: "Cihaz ekle" }).click();

  await expect(page.getByText("Test Device")).toBeVisible();
  await expect(page.getByText(testKey)).toBeVisible();
});
