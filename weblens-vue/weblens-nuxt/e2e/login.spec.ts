import { test, expect } from './fixtures'

/**
 * Tests for the login page and authentication flow.
 *
 * These tests exercise:
 * - pages/login.vue (form validation, error display, branding, links)
 * - stores/user.ts (login, authentication state)
 * - middleware/auth.ts (redirect when already authenticated)
 * - components/atom/WeblensButton.vue (disabled state)
 * - components/atom/Logo.vue (branding display)
 */
test.describe('Login Page', () => {
    test('should login with valid credentials', async ({ page }) => {
        await page.goto('/login')

        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()

        await page.waitForURL('**/files/home')
    })

    test('should show error for invalid credentials', async ({ page }) => {
        await page.goto('/login')

        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('wrongpassword')
        await page.getByRole('button', { name: 'Sign in' }).click()

        await expect(page.locator('.text-red-500')).toHaveText('Invalid username or password')
    })

    test('should disable sign in button when fields are empty', async ({ page }) => {
        await page.goto('/login')

        await expect(page.getByRole('button', { name: 'Sign in' })).toBeDisabled()
    })

    test('should disable sign in button when password is too short', async ({ page }) => {
        await page.goto('/login')

        // Fill only username with a short password (< 6 chars)
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('short')

        // Button should still be disabled
        await expect(page.getByRole('button', { name: 'Sign in' })).toBeDisabled()

        // Fill password with exactly 6 chars — button should become enabled
        await page.getByPlaceholder('Password').fill('123456')
        await expect(page.getByRole('button', { name: 'Sign in' })).toBeEnabled()
    })

    test('should show error when logging in with non-existent user', async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('nonexistent_user')
        await page.getByPlaceholder('Password').fill('somepassword')
        await page.getByRole('button', { name: 'Sign in' }).click()

        // Should show error message
        await expect(page.locator('.text-red-500')).toBeVisible({ timeout: 15000 })
    })

    test('should show New Here link and GitHub link on login page', async ({ page }) => {
        await page.goto('/login')

        // Should see "New Here?" text and signup link
        await expect(page.getByText('New Here?')).toBeVisible()
        await expect(page.getByRole('link', { name: 'Request an Account' })).toBeVisible()

        // Should see GitHub link
        await expect(page.getByRole('link', { name: /GitHub/ })).toBeVisible()
    })

    test('should display the Weblens branding on the login page', async ({ page }) => {
        await page.goto('/login')

        // Wait for page to fully load
        await expect(page.getByPlaceholder('Username')).toBeVisible()

        // The "EBLENS" heading exists in the DOM (the "W" is rendered via the Logo component)
        // It has responsive visibility: hidden by default, xs:inline-block at 30rem+
        await expect(page.getByRole('heading', { name: 'EBLENS' })).toHaveCount(1)
    })

    test('should redirect to files/home when visiting login page while already authenticated', async ({ page }) => {
        // Login first
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')

        // Try to go back to login page — should redirect to /files/home
        await page.goto('/login')
        await page.waitForURL('**/files/home')
        await expect(page).toHaveURL(/\/files\/home/)
    })
})
