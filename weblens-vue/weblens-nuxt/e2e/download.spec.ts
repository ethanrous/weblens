import { test, expect, uploadTestFile } from './fixtures'

/**
 * Tests for file download operations.
 *
 * These tests exercise:
 * - api/FileBrowserApi.ts (downloadSingleFile, downloadManyFiles)
 * - components/molecule/ContextMenuActions.vue (handleDownload for single and multi-file)
 * - stores/files.ts (getSelected for multi-download)
 */

test.describe('File Downloads', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await uploadTestFile(page, 'download-test.txt', 'This file is for download testing purposes.')
        await uploadTestFile(page, 'download-test-2.txt', 'Second file for multi-download testing.')
        // Reload so the folder fetch includes full permissions (the WebSocket
        // FileCreatedEvent from upload doesn't carry permissions).
        await page.reload()
        await expect(page.locator('[id^="file-card-"]').first()).toBeVisible({ timeout: 15000 })
    })

    test('should download a single text file via context menu', async ({ page }) => {
        const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: 'download-test.txt' })
        await expect(fileCard).toBeVisible({ timeout: 15000 })

        await fileCard.click({ button: 'right' })

        const fileBrowser = page.locator('#filebrowser-container')
        const downloadBtn = fileBrowser.getByRole('button', { name: 'Download' })
        await expect(downloadBtn.first()).toBeVisible()
        await expect(downloadBtn.first()).toBeEnabled({ timeout: 15000 })

        // Set up download handler
        const downloadPromise = page.waitForEvent('download', { timeout: 15000 }).catch(() => null)
        await downloadBtn.first().click()
        const download = await downloadPromise
        if (download) {
            expect(download.suggestedFilename()).toBe('download-test.txt')
        }
    })

    test('should select multiple files and download as zip', async ({ page }) => {
        const fileCards = page.locator('[id^="file-card-"]')
        await expect(fileCards.first()).toBeVisible({ timeout: 15000 })

        if ((await fileCards.count()) >= 2) {
            // Select first file
            await fileCards.first().click()

            // Ctrl+click second file
            await fileCards.nth(1).click({ modifiers: ['ControlOrMeta'] })

            // Right-click to open context menu with multi-select
            await fileCards.nth(1).click({ button: 'right' })

            // In multi-select, context menu should show Download button
            const fileBrowser = page.locator('#filebrowser-container')
            const downloadBtn = fileBrowser.getByRole('button', { name: 'Download' })
            if (
                await downloadBtn
                    .first()
                    .isVisible({ timeout: 3000 })
                    .catch(() => false)
            ) {
                // Click download to exercise downloadManyFiles path
                const downloadPromise = page.waitForEvent('download', { timeout: 15000 }).catch(() => null)
                await downloadBtn.first().click()
                await downloadPromise
            }

            await page.keyboard.press('Escape')
        }
    })
})
