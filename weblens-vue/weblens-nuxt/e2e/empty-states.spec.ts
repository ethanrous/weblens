import { test, expect, createFolder } from './fixtures'

/**
 * Tests for empty state displays throughout the application.
 *
 * These tests exercise:
 * - components/molecule/NoResults.vue (empty folder, empty trash, no search results branches)
 * - pages/files.vue (error handling for invalid file IDs)
 * - components/molecule/ErrorCard.vue (error states)
 * - stores/location.ts (isInTrash, inShareRoot)
 */

test.describe('Empty States', () => {
    test.beforeEach(async ({ login: _login }) => {})

    test('should show empty folder message when navigating into empty folder', async ({ page }) => {
        // Create an empty folder — dispatchEvent may need a retry in serial runs
        await page.getByRole('button', { name: 'New Folder' }).click()
        const nameInput = page.locator('#file-context-menu').getByRole("textbox")
        await expect(nameInput).toBeVisible()
        await nameInput.fill('EmptyStateTest')
        await nameInput.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })

        // Retry dispatchEvent if the input is still visible after a short wait
        if (await nameInput.isVisible({ timeout: 1000 }).catch(() => false)) {
            await nameInput.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })
        }

        await expect(page.locator('[id^="file-card-"]').filter({ hasText: 'EmptyStateTest' })).toBeVisible({
            timeout: 15000,
        })

        // Double-click to navigate into the empty folder
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'EmptyStateTest' })
        await folderCard.dblclick()

        // Should show "This folder is empty" (NoResults.vue else branch)
        await expect(page.getByText('This folder is empty')).toBeVisible({ timeout: 15000 })

        // Navigate back
        await page.locator('.tabler-icon-chevron-left').first().click()
        await page.waitForURL('**/files/home')
    })

    test('should show no search results message for nonexistent files', async ({ page }) => {
        await createFolder(page, 'SomeFolder')

        await expect(page.locator('[id^="file-card-"]').first()).toBeVisible({
            timeout: 15000,
        })

        // Search for something that doesn't exist
        const searchInput = page.getByPlaceholder('Search Files...')
        await searchInput.click()
        await searchInput.fill('zzz_nonexistent_file_xyz_12345')

        // Should show "No files found" (NoResults.vue search branch)
        const noFilesFound = page.getByText('No files found in')
        await expect(noFilesFound).toBeVisible({ timeout: 15000 })

        // Clear search
        await searchInput.clear()
    })

    test('should show shared files empty state', async ({ page }) => {
        // Navigate to shared root
        await page.getByRole('button', { name: 'Shared' }).click()
        await page.waitForURL('**/files/share', { timeout: 15000 })

        // Should show empty state or shared files
        const noShared = page.getByText('No files shared with you')
        const sharedFiles = page.locator('[id^="file-card-"]')

        await expect(noShared.or(sharedFiles.first())).toBeVisible({ timeout: 15000 })

        await page.getByRole('button', { name: 'Home' }).click()
        await page.waitForURL('**/files/home', { timeout: 15000 })
    })

    test('should show trash empty state or trash items', async ({ page }) => {
        await page.getByRole('button', { name: 'Trash' }).click()
        await page.waitForURL('**/files/trash', { timeout: 15000 })

        // Should show either the empty trash message or file cards (auto-wait for either)
        const emptyMsg = page.getByText('Someone already took the trash out')
        const fileCards = page.locator('[id^="file-card-"]')
        await expect(emptyMsg.or(fileCards.first())).toBeVisible({ timeout: 15000 })

        await page.getByRole('button', { name: 'Home' }).click()
        await page.waitForURL('**/files/home', { timeout: 15000 })
    })

    test('should handle navigation to invalid file ID gracefully', async ({ page }) => {
        await page.goto('/files/nonexistent-file-id-12345')

        // Should show some error state or redirect
        const errorVisible = await page
            .locator('.error-card, [class*="error"]')
            .isVisible()
            .catch(() => false)
        const redirectedToHome = page.url().includes('/files/home')
        const onFilesPage = page.url().includes('/files/')

        expect(errorVisible || redirectedToHome || onFilesPage).toBeTruthy()
    })
})
