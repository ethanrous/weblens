import { test, expect, login, createUser } from './fixtures'

/**
 * Tests for multi-user scenarios and non-admin user functionality.
 *
 * These tests exercise:
 * - Login as a second user
 * - Non-admin user settings page (fewer tabs)
 * - stores/user.ts (different user context)
 * - pages/login.vue (different credential handling)
 * - Login error edge cases
 */
test.describe('Multi-User Scenarios', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await createUser(page, 'regular_user', 'regularpass123')
    })

    test('should login as regular user and see limited settings', async ({ page }) => {
        // Log out of admin session and log in as regular_user
        await page.goto('/settings')
        await page.getByRole('button', { name: 'Log Out' }).click()
        await page.waitForURL('**/login')
        await login(page, 'regular_user', 'regularpass123')

        // Navigate to settings
        await page.getByRole('button', { name: 'Settings' }).click()
        await page.waitForURL('**/settings/account')

        // Regular users should see Account and Appearance but NOT Users or Developer
        await expect(page.getByRole('button', { name: 'Account' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'Appearance' })).toBeVisible()

        // Admin-only tabs should not be visible
        await expect(page.getByRole('button', { name: 'Users' })).not.toBeVisible()
        await expect(page.getByRole('button', { name: 'Developer' })).not.toBeVisible()

        // The "Admin" divider should also not be visible
        await expect(page.getByText('Admin', { exact: true })).not.toBeVisible()
    })

    test('should navigate regular user file browser and create folder', async ({ page }) => {
        // Log out of admin session and log in as regular_user
        await page.goto('/settings')
        await page.getByRole('button', { name: 'Log Out' }).click()
        await page.waitForURL('**/login')
        await login(page, 'regular_user', 'regularpass123')

        // Should see the sidebar and file browser
        await expect(page.getByRole('button', { name: 'Home' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'Shared' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'Trash' })).toBeVisible()

        // Should see the file browser header
        await expect(page.locator('h3').filter({ hasText: 'Home' })).toBeVisible()

        // Should see the search bar
        await expect(page.getByPlaceholder('Search Files...')).toBeVisible()

        // Create a folder as regular user
        await page.getByRole('button', { name: 'New Folder' }).click()
        const nameInput = page.locator('.file-context-menu input')
        await expect(nameInput).toBeVisible()
        await nameInput.fill('Regular User Folder')
        await nameInput.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })
        const regularFolderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'Regular User Folder' })
        await expect(regularFolderCard).toBeVisible({ timeout: 15000 })
        await expect(nameInput).not.toBeVisible({ timeout: 3000 })

        // Verify the folder is clickable and can be navigated into
        await regularFolderCard.dblclick()
        await expect(page.locator('h3').filter({ hasText: 'Regular User Folder' })).toBeVisible({ timeout: 15000 })

        // Navigate back
        await page.locator('.tabler-icon-chevron-left').first().click()
        await page.waitForURL('**/files/home')
    })

    test('should validate password change form as regular user', async ({ page }) => {
        // Log out of admin session and log in as regular_user
        await page.goto('/settings')
        await page.getByRole('button', { name: 'Log Out' }).click()
        await page.waitForURL('**/login')
        await login(page, 'regular_user', 'regularpass123')

        // Go to account settings
        await page.goto('/settings/account')
        await page.waitForURL('**/settings/account')

        // The Update Password button should be disabled when fields are empty
        const updateBtn = page.getByRole('button', { name: 'Update Password' })
        await expect(updateBtn).toBeDisabled()

        // Fill old password only — button should still be disabled
        await page.getByPlaceholder('Old Password').fill('regularpass123')
        await expect(updateBtn).toBeDisabled()

        // Fill new password same as old — button should still be disabled
        await page.getByPlaceholder('New Password').fill('regularpass123')
        await expect(updateBtn).toBeDisabled()

        // Fill different new password — button should now be enabled
        await page.getByPlaceholder('New Password').clear()
        await page.getByPlaceholder('New Password').fill('newregularpass456')
        await expect(updateBtn).toBeEnabled()

        // Also verify the display name input is visible
        await expect(page.getByPlaceholder('Jonh Smith')).toBeVisible()
    })
})
