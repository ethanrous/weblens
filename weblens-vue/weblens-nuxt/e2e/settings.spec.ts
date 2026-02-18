import { test, expect, DEFAULT_ADMIN_USERNAME } from './fixtures'

test.describe('Settings Page', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await page.waitForURL('**/files/home')
        await page.goto('/settings')
        await page.waitForURL('**/settings/account')
    })

    test('should display settings nav with all tabs', async ({ page }) => {
        // User info header
        await expect(page.locator('h3').filter({ hasText: 'Weblens Admin' })).toBeVisible()
        await expect(page.locator('h5').filter({ hasText: DEFAULT_ADMIN_USERNAME })).toBeVisible()

        // Navigation buttons
        await expect(page.getByRole('button', { name: 'Account' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'Appearance' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'Users' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'Developer' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'Log Out' })).toBeVisible()

        // Admin divider (exact match to avoid matching "Weblens Admin" heading)
        await expect(page.getByText('Admin', { exact: true })).toBeVisible()
    })

    test('should change display name', async ({ page }) => {
        // The display name input has placeholder "Jonh Smith"
        const nameInput = page.getByPlaceholder('Jonh Smith')
        await expect(nameInput).toBeVisible()

        // Clear and type a new name
        await nameInput.clear()
        await nameInput.fill('Updated Admin')

        // Click Update Name button
        await page.getByRole('button', { name: 'Update Name' }).click()

        // Navigate away and back to verify persistence
        await page.getByRole('button', { name: 'Appearance' }).click()
        await page.waitForURL('**/settings/appearance')
        await page.getByRole('button', { name: 'Account' }).click()
        await page.waitForURL('**/settings/account')

        // The input should still show the updated name
        await expect(page.getByPlaceholder('Jonh Smith')).toHaveValue('Updated Admin')

        // Restore original name
        await nameInput.clear()
        await nameInput.fill('Weblens Admin')
        await page.getByRole('button', { name: 'Update Name' }).click()
    })

    test('should validate password change form', async ({ page }) => {
        const updatePasswordBtn = page.getByRole('button', { name: 'Update Password' })

        // Button should be disabled when fields are empty
        await expect(updatePasswordBtn).toBeDisabled()

        // Fill old password only — button should still be disabled
        await page.getByPlaceholder('Old Password').fill('adminadmin1')
        await expect(updatePasswordBtn).toBeDisabled()

        // Fill new password same as old — button should still be disabled
        await page.getByPlaceholder('New Password').fill('adminadmin1')
        await expect(updatePasswordBtn).toBeDisabled()

        // Fill different new password — button should now be enabled
        await page.getByPlaceholder('New Password').clear()
        await page.getByPlaceholder('New Password').fill('newpassword456')
        await expect(updatePasswordBtn).toBeEnabled()
    })

    test('should create and delete API key', async ({ page }) => {
        // Initially should show "No API Keys found."
        await expect(page.getByText('No API Keys found.')).toBeVisible()

        // Fill in API key name and create
        await page.getByPlaceholder('API Key Name').fill('Test Key')
        await page.getByRole('button', { name: 'Create New API Key' }).click()

        // The key should appear in the list
        await expect(page.getByText('Test Key')).toBeVisible({ timeout: 15000 })
        await expect(page.getByText('No API Keys found.')).not.toBeVisible()

        // Delete the API key — click the trash icon button next to it
        await page
            .locator('[data-flavor="danger"]')
            .filter({ has: page.locator('.tabler-icon-trash') })
            .first()
            .click()

        // Confirm dialog should appear
        await expect(page.getByText('Are you sure?')).toBeVisible()
        await page.getByRole('button', { name: 'Confirm' }).click()

        // Key should be removed
        await expect(page.getByText('No API Keys found.')).toBeVisible({ timeout: 15000 })
    })

    test('should toggle dark mode', async ({ page }) => {
        await page.getByRole('button', { name: 'Appearance' }).click()
        await page.waitForURL('**/settings/appearance')

        await expect(page.getByText('Appearance Settings')).toBeVisible()
        await expect(page.getByRole('heading', { name: 'Theme' })).toBeVisible()

        // Toggle the dark mode checkbox
        const checkbox = page.locator('input[type="checkbox"]')
        await expect(checkbox).toBeVisible()

        // Get the initial state and toggle
        const wasChecked = await checkbox.isChecked()
        await checkbox.click()

        // Verify the checkbox state changed
        if (wasChecked) {
            await expect(checkbox).not.toBeChecked()
        } else {
            await expect(checkbox).toBeChecked()
        }

        // Toggle back to restore original state
        await checkbox.click()
    })

    test('should navigate between settings tabs', async ({ page }) => {
        // Click Appearance
        await page.getByRole('button', { name: 'Appearance' }).click()
        await page.waitForURL('**/settings/appearance')
        await expect(page.getByText('Appearance Settings')).toBeVisible()

        // Click Users
        await page.getByRole('button', { name: 'Users' }).click()
        await page.waitForURL('**/settings/users')
        await expect(page.locator('h4').filter({ hasText: 'Users' }).first()).toBeVisible()

        // Click Developer
        await page.getByRole('button', { name: 'Developer' }).click()
        await page.waitForURL('**/settings/dev')
        await expect(page.getByRole('button', { name: 'Refresh' })).toBeVisible()

        // Click Account
        await page.getByRole('button', { name: 'Account' }).click()
        await page.waitForURL('**/settings/account')
        await expect(page.locator('h3').filter({ hasText: 'Account' })).toBeVisible()
    })

    test('should create a new user', async ({ page }) => {
        await page.getByRole('button', { name: 'Users' }).click()
        await page.waitForURL('**/settings/users')

        // Fill in new user details
        await page.getByPlaceholder('Username').fill('test_user_e2e')
        await page.getByPlaceholder('Password').fill('testpass123')

        // Click Create User
        await page.getByRole('button', { name: 'Create User' }).click()

        // The new user should appear in the users table
        await expect(page.getByText('test_user_e2e')).toBeVisible({ timeout: 15000 })
    })

    test('should display developer page controls', async ({ page }) => {
        await page.getByRole('button', { name: 'Developer' }).click()
        await page.waitForURL('**/settings/dev')

        // Verify all expected buttons are visible
        await expect(page.getByRole('button', { name: 'Refresh' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'Scan All Media' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'Enable trace logging' })).toBeVisible()
        await expect(page.getByRole('button', { name: /HDIR image processing/ })).toBeVisible()

        // Danger zone buttons
        await expect(page.getByRole('button', { name: 'Clear Media HDIR Data' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'Clean Media' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'Flush Cache' })).toBeVisible()

        // Task table should show empty state
        await expect(page.getByText('No running tasks')).toBeVisible()
    })

    test('should log out', async ({ page }) => {
        await page.getByRole('button', { name: 'Log Out' }).click()

        // Should redirect to login page
        await page.waitForURL('**/login')
        await expect(page).toHaveURL(/\/login$/)
    })
})
