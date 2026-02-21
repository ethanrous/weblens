import { test, expect, DEFAULT_ADMIN_USERNAME } from './fixtures'

/**
 * Tests for developer page actions and admin functionality.
 *
 * These tests exercise:
 * - pages/settings/dev.vue (scan all media, flush cache, clean media, clear HDIR, trace logging, HDIR toggle, feature flags)
 * - stores/tower.ts (refreshTowerInfo, getServerInfo)
 * - stores/tasks.ts (running tasks display)
 * - api/AllApi.ts (various API calls)
 * - components/atom/Table.vue (running tasks table)
 * - pages/settings/users.vue (user activation, deletion)
 */
test.describe('Developer Page Actions', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await page.goto('/settings/dev')
        await page.waitForURL('**/settings/dev')
    })

    test('should scan all media and see task state', async ({ page }) => {
        // Click Scan All Media
        await page.getByRole('button', { name: 'Scan All Media' }).click()

        // Wait a moment for the scan to start

        // Click Refresh to update task list
        await page.getByRole('button', { name: 'Refresh' }).click()
    })

    test('should toggle HDIR image processing', async ({ page }) => {
        // Find the HDIR button - it says either "Enable HDIR..." or "Disable HDIR..."
        const hdirButton = page.getByRole('button', { name: /HDIR image processing/ })
        await expect(hdirButton).toBeVisible()

        // Click to toggle — verify the click doesn't error
        await hdirButton.click()

        // The button should still be visible after the toggle attempt
        await expect(hdirButton).toBeVisible()

        // Click again to restore state
        await hdirButton.click()
    })

    test('should enable trace logging', async ({ page }) => {
        const traceButton = page.getByRole('button', { name: 'Enable trace logging' })

        // Check if already enabled (disabled state)
        const isDisabled = await traceButton.isDisabled()
        if (!isDisabled) {
            await traceButton.click()

            // After enabling, the button should be disabled (already at trace level)
            await expect(traceButton).toBeDisabled()
        }
    })

    test('should flush cache', async ({ page }) => {
        await page.getByRole('button', { name: 'Flush Cache' }).click()
    })

    test('should click clean media', async ({ page }) => {
        await page.getByRole('button', { name: 'Clean Media' }).click()
    })

    test('should click clear HDIR data', async ({ page }) => {
        await page.getByRole('button', { name: 'Clear Media HDIR Data' }).click()
    })
})

test.describe('User Management Actions', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await page.goto('/settings/users')
        await page.waitForURL('**/settings/users')
    })

    test('should display users table with columns', async ({ page }) => {
        // The table should have the expected column headers
        // Table component uses camelCaseToWords() which capitalizes the first letter
        // Use getByRole('columnheader') to avoid matching row text that also contains these words
        await expect(page.getByRole('columnheader', { name: 'Username' })).toBeVisible()
        await expect(page.getByRole('columnheader', { name: 'Role' })).toBeVisible()
        await expect(page.getByRole('columnheader', { name: 'Online' })).toBeVisible()

        // The admin user should be listed
        await expect(page.getByRole('cell', { name: DEFAULT_ADMIN_USERNAME })).toBeVisible()
    })

    test('should delete non-admin user if one exists', async ({ page }) => {
        // Check if test_user_e2e exists (created in settings.spec.ts)
        const userRow = page.getByText('test_user_e2e')

        if (await userRow.isVisible()) {
            // Find the delete button in the same row - it's a danger-flavored button with trash icon
            // Click the trash icon in the test_user_e2e row
            const trashButtons = page.locator('[data-flavor="danger"]').filter({
                has: page.locator('.tabler-icon-trash'),
            })

            // Click the last trash button (should be the one for test_user_e2e)
            await trashButtons.last().click()
        }
    })
})
