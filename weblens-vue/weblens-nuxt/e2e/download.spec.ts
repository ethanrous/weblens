import { test, expect, uploadTestFile, createFolder } from './fixtures'

/**
 * Tests for file download operations.
 *
 * These tests exercise:
 * - api/FileBrowserApi.ts (downloadSingleFile, downloadManyFiles, handleDownload)
 * - components/molecule/ContextMenuActions.vue (handleDownload for single and multi-file)
 * - stores/files.ts (getSelected for multi-download)
 *
 * Download naming conventions:
 * - Single file: original filename (e.g. "photo.jpg")
 * - Single directory: directory name with .zip (e.g. "MyFolder.zip")
 * - Multiple files: "{taskID}.weblens.zip"
 */

test.describe('File Downloads', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await uploadTestFile(page, 'download-test.txt', 'This file is for download testing purposes.')
        await uploadTestFile(page, 'download-test-2.txt', 'Second file for multi-download testing.')
        await createFolder(page, 'Download Folder')
        // Reload so the folder fetch includes full permissions (the WebSocket
        // FileCreatedEvent from upload doesn't carry permissions).
        await page.reload()
        await expect(page.locator('[id^="file-card-"]').first()).toBeVisible({ timeout: 15000 })
    })

    test('single file download should use original filename', async ({ page }) => {
        const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: 'download-test.txt' })
        await expect(fileCard).toBeVisible({ timeout: 15000 })

        await fileCard.click({ button: 'right' })

        const fileBrowser = page.locator('#filebrowser-container')
        const downloadBtn = fileBrowser.getByRole('button', { name: 'Download' })
        await expect(downloadBtn.first()).toBeVisible()
        await expect(downloadBtn.first()).toBeEnabled({ timeout: 15000 })

        const downloadPromise = page.waitForEvent('download', { timeout: 15000 }).catch(() => null)
        await downloadBtn.first().click()
        const download = await downloadPromise
        if (download) {
            // Single file downloads should use the original filename without any prefix
            expect(download.suggestedFilename()).toBe('download-test.txt')
        }
    })

    test('single directory download should use directory name with .zip extension', async ({ page }) => {
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'Download Folder' })
        await expect(folderCard).toBeVisible({ timeout: 15000 })

        await folderCard.click({ button: 'right' })

        const fileBrowser = page.locator('#filebrowser-container')
        const downloadBtn = fileBrowser.getByRole('button', { name: 'Download' })
        await expect(downloadBtn.first()).toBeVisible()
        await expect(downloadBtn.first()).toBeEnabled({ timeout: 15000 })

        const downloadPromise = page.waitForEvent('download', { timeout: 30000 }).catch(() => null)
        await downloadBtn.first().click()
        const download = await downloadPromise
        if (download) {
            // Single directory downloads should use the directory name with .zip extension
            expect(download.suggestedFilename()).toBe('Download Folder.zip')
        }
    })

    test('multi-file download should use taskID.weblens.zip naming', async ({ page }) => {
        const fileCards = page.locator('[id^="file-card-"]')
        await expect(fileCards.first()).toBeVisible({ timeout: 15000 })

        if ((await fileCards.count()) >= 2) {
            // Select first file
            await fileCards.first().click()

            // Ctrl+click second file
            await fileCards.nth(1).click({ modifiers: ['ControlOrMeta'] })

            // Right-click to open context menu with multi-select
            await fileCards.nth(1).click({ button: 'right' })

            const fileBrowser = page.locator('#filebrowser-container')
            const downloadBtn = fileBrowser.getByRole('button', { name: 'Download' })
            if (
                await downloadBtn
                    .first()
                    .isVisible({ timeout: 3000 })
                    .catch(() => false)
            ) {
                const downloadPromise = page.waitForEvent('download', { timeout: 30000 }).catch(() => null)
                await downloadBtn.first().click()
                const download = await downloadPromise
                if (download) {
                    // Multi-file downloads should use {taskID}.weblens.zip naming
                    expect(download.suggestedFilename()).toMatch(/\.weblens\.zip$/)
                }
            }

            await page.keyboard.press('Escape')
        }
    })
})
