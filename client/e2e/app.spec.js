// @ts-check
import { test, expect } from '@playwright/test';

test('homepage has "I AM HERE TO WORK FOR YOU" and a link to google', async ({ page }) => {
  await page.goto('http://localhost:8081/');

  // Add a longer wait to ensure the React app has time to render
  await page.waitForSelector('nav', { state: 'visible', timeout: 10000 });

  // Print page content for debugging
  const pageContent = await page.content();
  console.log(pageContent);

  // Expect a title "to contain" a substring.
  await expect(page).toHaveTitle(/Vite \+ React/);

  // Expect the main message to be visible
  await expect(page.locator('p').getByText('I AM HERE TO WORK FOR YOU')).toBeVisible();

  // Expect the link to be visible and have the correct href
  const googleLink = page.locator('a').getByText('Go to Google');
  await expect(googleLink).toBeVisible();
  await expect(googleLink).toHaveAttribute('href', 'https://google.com');

  // Test navigation to the About page
  const aboutLink = page.locator('nav').getByText('About');
  await expect(aboutLink).toBeVisible();
  await aboutLink.click();
  await expect(page.locator('h2').getByText('About Us')).toBeVisible();
  await expect(page.locator('p').getByText('This is the about page.')).toBeVisible();

  // Test navigation back to the Home page
  const homeLink = page.locator('nav').getByText('Home');
  await expect(homeLink).toBeVisible();
  await homeLink.click();
  await expect(page.locator('p').getByText('I AM HERE TO WORK FOR YOU')).toBeVisible();
});

test('non-existent path falls back to index.html and loads SPA', async ({ page }) => {
  await page.goto('http://localhost:8081/this-path-does-not-exist');

  // Ensure the SPA has time to render
  await page.waitForSelector('nav', { state: 'visible', timeout: 10000 });

  // Expect the page title to still be from the SPA's index.html
  await expect(page).toHaveTitle(/Vite \+ React/);

  // Expect a key element from the SPA to be visible, confirming it loaded
  await expect(page.locator('p').getByText('I AM HERE TO WORK FOR YOU')).toBeVisible();
});

test('static asset (horse.webp) is served correctly via network interception', async ({ page }) => {
  // Navigate to the homepage
  await page.goto('http://localhost:8081/');

  // Wait for the response of the horse.webp image
  // The regex `/\/assets\/horse-.*\.webp/` will match the dynamically hashed filename
  const response = await page.waitForResponse(response =>
    response.url().match(/.\/assets\/horse-.*\.webp/) && response.request().resourceType() === 'image'
  );

  // Expect the response status to be 200 OK
  expect(response.status()).toBe(200);

  // Optionally, check content type
  expect(response.headers()['content-type']).toContain('image/webp');
});