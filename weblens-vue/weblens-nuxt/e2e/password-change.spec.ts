import { test, expect } from './fixtures'

/**
 * Tests for password change functionality on the account settings page.
 *
 * These tests exercise:
 * - pages/settings/account.vue (password form, handleChangePasswordFail, handleChangePasswordSuccess)
 * - components/atom/WeblensButton.vue (error-text state, disabled state)
 * - components/atom/WeblensInput.vue (clear, submit callbacks)
 * - api/AllApi.ts (changePassword endpoint)
 */

test.describe('Password Change', () => {
    test.describe.configure({ mode: 'serial' })

    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
        await page.goto('/settings/account')
        await page.waitForURL('**/settings/account')
    })

    test('should show error when attempting password change with wrong old password', async ({ page }) => {
        await page.getByPlaceholder('Old Password').fill('wrongpassword')
        await page.getByPlaceholder('New Password').fill('newpass123')

        const updateBtn = page.getByRole('button', { name: 'Update Password' })
        await expect(updateBtn).toBeEnabled()
        await updateBtn.click()

        // Should show error text (handleChangePasswordFail returns "Incorrect old password")
        // WeblensButton renders text twice (visible + hidden measurement span), so use .first()
        await expect(page.getByText('Incorrect old password').first()).toBeVisible({ timeout: 15000 })
    })

    test('should change password successfully and change back', async ({ page }) => {
        // Fill correct old password and new password
        await page.getByPlaceholder('Old Password').fill('password123')
        await page.getByPlaceholder('New Password').fill('newpass456')

        const updateBtn = page.getByRole('button', { name: 'Update Password' })
        await expect(updateBtn).toBeEnabled()
        await updateBtn.click()

        // Wait for the password change to complete
        await page.waitForTimeout(1000)

        // Inputs should be cleared after successful change
        await expect(page.getByPlaceholder('Old Password')).toHaveValue('')

        // Change back to original password
        await page.getByPlaceholder('Old Password').fill('newpass456')
        await page.getByPlaceholder('New Password').fill('password123')
        await page.getByRole('button', { name: 'Update Password' }).click()
        await page.waitForTimeout(1000)
    })
})
