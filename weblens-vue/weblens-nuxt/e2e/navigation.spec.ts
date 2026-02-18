import { test, expect } from './fixtures'

/**
 * Tests for navigation, redirects, and route protection.
 *
 * These tests exercise:
 * - middleware/auth.ts (route protection, unauthenticated redirects)
 * - middleware/redirect.ts (root and /files redirects)
 * - pages/files.vue (authenticated file browser access)
 * - pages/settings.vue (unauthenticated settings behavior)
 * - components/organism/FileSidebar.vue (sidebar visibility)
 */
test.describe('Navigation and Redirects', () => {
    test('should redirect root to login when not authenticated', async ({ page }) => {
        await page.goto('/')
        await page.waitForURL('**/login')
        await expect(page).toHaveURL(/\/login$/)
    })

    test('should redirect unauthenticated user accessing /files/home to /login', async ({ page }) => {
        // Navigate directly to a protected page without logging in
        await page.goto('/files/home')
        await page.waitForURL('**/login')
        await expect(page).toHaveURL(/\/login/)
    })

    test('should show Log In button on settings page when unauthenticated', async ({ page }) => {
        // Settings page does not redirect unauthenticated users to /login.
        // Instead, it shows the settings layout with a "Log In" button.
        await page.goto('/settings')

        await expect(page.getByRole('button', { name: 'Log In' })).toBeVisible({ timeout: 15000 })
    })

    test('should redirect root to files/home when authenticated', async ({ page, login: _login }) => {
        // Navigate to root — should redirect to files/home
        await page.goto('/')
        await page.waitForURL('**/files/home')
        await expect(page).toHaveURL(/\/files\/home$/)
    })

    test('should redirect /files to /files/home when logged in', async ({ page, login: _login }) => {
        // Navigate directly to /files
        await page.goto('/files')
        await page.waitForURL('**/files/home')
        await expect(page).toHaveURL(/\/files\/home/)
    })

    test('should redirect /settings to /settings/account', async ({ page, login: _login }) => {
        // Navigate to /settings — should redirect to /settings/account
        await page.goto('/settings')
        await page.waitForURL('**/settings/account')
        await expect(page).toHaveURL(/\/settings\/account$/)
    })

    test('should handle direct navigation to /files/trash', async ({ page, login: _login }) => {
        // Navigate to trash directly
        await page.goto('/files/trash')

        // Should show the trash page header
        await expect(page.locator('h3').filter({ hasText: 'Trash' })).toBeVisible({
            timeout: 15000,
        })
    })

    test('should show sidebar on files pages but collapsed on settings', async ({ page, login: _login }) => {
        // On files page, sidebar should be visible
        await expect(page.getByRole('button', { name: 'Home' })).toBeVisible()

        // Navigate to settings — sidebar should still be present but collapsed
        await page.goto('/settings')
        await page.waitForURL('**/settings/account')

        // Sidebar buttons should still be visible on settings pages
        await expect(page.getByRole('button', { name: 'Home' })).toBeVisible()
    })
})
