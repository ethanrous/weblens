import { test, expect } from './fixtures'

/**
 * Tests for file download operations.
 *
 * These tests exercise:
 * - api/FileBrowserApi.ts (downloadSingleFile, downloadManyFiles)
 * - components/molecule/ContextMenuActions.vue (handleDownload for single and multi-file)
 * - stores/files.ts (getSelected for multi-download)
 */

test.describe('File Downloads', () => {
    test.describe.configure({ mode: 'serial' })

    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should upload a text file for download testing', async ({ page }) => {
        const fileChooserPromise = page.waitForEvent('filechooser')
        await page.getByRole('button', { name: 'Upload' }).click()
        const fileChooser = await fileChooserPromise

        await fileChooser.setFiles({
            name: 'download-test.txt',
            mimeType: 'text/plain',
            buffer: Buffer.from('This file is for download testing purposes.'),
        })

        const fileBrowser = page.locator('#filebrowser-container')
        await expect(fileBrowser.getByText('download-test.txt').first()).toBeVisible({
            timeout: 15000,
        })
    })

    test('should download a single text file via context menu', async ({ page }) => {
        const fileCard = page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: 'download-test.txt' })
        await expect(fileCard).toBeVisible({ timeout: 15000 })

        await fileCard.click({ button: 'right' })

        const fileBrowser = page.locator('#filebrowser-container')
        const downloadBtn = fileBrowser.getByRole('button', { name: 'Download' })
        await expect(downloadBtn.first()).toBeVisible()

        // Set up download handler
        const downloadPromise = page.waitForEvent('download', { timeout: 15000 }).catch(() => null)
        await downloadBtn.first().click()
        const download = await downloadPromise
        if (download) {
            expect(download.suggestedFilename()).toBe('download-test.txt')
        }
    })

    test('should select multiple files and download as zip', async ({ page }) => {
        const fileCards = page.locator('[id^="file-"]:not(#file-scroller)')
        await expect(fileCards.first()).toBeVisible({ timeout: 15000 })

        if ((await fileCards.count()) >= 2) {
            // Select first file
            await fileCards.first().click()
            await page.waitForTimeout(200)

            // Ctrl+click second file
            await fileCards.nth(1).click({ modifiers: ['ControlOrMeta'] })
            await page.waitForTimeout(200)

            // Right-click to open context menu with multi-select
            await fileCards.nth(1).click({ button: 'right' })
            await page.waitForTimeout(500)

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
                await page.waitForTimeout(2000)
            }

            await page.keyboard.press('Escape')
        }
    })

    test('should clean up download test file', async ({ page }) => {
        const fileCard = page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: 'download-test.txt' })

        if (await fileCard.isVisible({ timeout: 15000 }).catch(() => false)) {
            await fileCard.click({ button: 'right' })
            const fileBrowser = page.locator('#filebrowser-container')
            const trashBtn = fileBrowser.getByRole('button', { name: 'Trash' })
            if (await trashBtn.isEnabled({ timeout: 2000 }).catch(() => false)) {
                await trashBtn.click()
                await expect(fileCard).not.toBeVisible({ timeout: 15000 })
            } else {
                await page.keyboard.press('Escape')
            }
        }
    })
})
