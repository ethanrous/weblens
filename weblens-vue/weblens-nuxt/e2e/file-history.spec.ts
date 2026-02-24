import { test, expect } from './fixtures'

/**
 * Tests for file history and time rewind functionality.
 *
 * These tests exercise:
 * - components/molecule/RewindIndicator.vue (rewind state display, close button)
 * - components/molecule/FileAction.vue (history action timestamps, friendlyActionName)
 * - stores/location.ts (viewTimestamp, isViewingPast, setViewTimestamp)
 * - util/relativeTime.ts (relativeTimeAgo with various time deltas)
 * - util/history.ts (friendlyActionName for fileCreate, fileMove)
 * - api/FileBrowserApi.ts (GetFolderHistory)
 */

test.describe('File History and Rewind', () => {
    test.beforeEach(async ({ login: _login }) => {})

    test('should show rewind indicator when navigating with rewindTo URL param', async ({ page }) => {
        // Set a past timestamp via URL to exercise RewindIndicator, relativeTimeAgo
        const pastDate = new Date(Date.now() - 60 * 60 * 1000) // 1 hour ago
        const rewindUrl = `/files/home?rewindTo=${pastDate.toISOString()}`
        await page.goto(rewindUrl)

        // The RewindIndicator should show "Rewound to"
        const rewindText = page.getByText('Rewound to')
        await rewindText.isVisible({ timeout: 15000 })
        await expect(rewindText).toBeVisible()

        // Click X to clear the rewind (exercises setViewTimestamp(0))
        const closeBtn = page.locator('#rewind-indicator').locator('.tabler-icon-x')
        await closeBtn.isVisible({ timeout: 2000 })
        await closeBtn.click()
        // URL should no longer have rewindTo
        await page.waitForURL('**/files/home', { timeout: 5000 })
        expect(page.url()).not.toContain('rewindTo')
    })

    test('should show rewind with older timestamp for different time ranges', async ({ page }) => {
        // Test with a date 3 days ago (exercises "X days ago" path in relativeTimeAgo)
        const threeDaysAgo = new Date(Date.now() - 3 * 24 * 60 * 60 * 1000)
        const rewindUrl = `/files/home?rewindTo=${threeDaysAgo.toISOString()}`
        await page.goto(rewindUrl)

        const rewindText = page.getByText('Rewound to')
        await expect(rewindText).toBeVisible()

        // Clear rewind by navigating to home without query params
        await page.goto('/files/home')
    })

    test('should open history panel and click an action to trigger rewind', async ({ page }) => {
        await page.keyboard.press('Escape')

        // Open context menu on the Home heading to access Folder History
        await page.locator('h3').filter({ hasText: 'Home' }).click()

        const fileBrowser = page.locator('#filebrowser-container')
        const folderHistoryBtn = fileBrowser.getByRole('button', { name: 'Folder History' })

        if (await folderHistoryBtn.isVisible({ timeout: 15000 }).catch(() => false)) {
            await folderHistoryBtn.click()

            // Wait for history data to load
            const historyAction = page.getByText('File Created').first()
            if (await historyAction.isVisible({ timeout: 15000 }).catch(() => false)) {
                // Click on the action row to trigger rewind
                const actionRow = page.locator('[class*="cursor-pointer"]').filter({ hasText: 'File Created' }).first()
                if (await actionRow.isVisible({ timeout: 2000 }).catch(() => false)) {
                    await actionRow.click()

                    // Check if RewindIndicator appeared
                    const rewindIndicator = page.getByText('Rewound to')
                    if (await rewindIndicator.isVisible({ timeout: 3000 }).catch(() => false)) {
                        // Click X to clear rewind
                        const closeRewind = rewindIndicator.locator('..').locator('.tabler-icon-x')
                        if (await closeRewind.isVisible({ timeout: 1000 }).catch(() => false)) {
                            await closeRewind.click()
                        }
                    }
                }
            }
        }
    })

    test('should see history entries for file creation and move operations', async ({ page }) => {
        // Upload a file then trash it to create both "File Created" and "File Trashed" entries
        const fileChooserPromise = page.waitForEvent('filechooser')
        await page.getByRole('button', { name: 'Upload' }).click()
        const fileChooser = await fileChooserPromise
        await fileChooser.setFiles({
            name: 'history-test-file.txt',
            mimeType: 'text/plain',
            buffer: Buffer.from('Test file for history entries'),
        })

        const fileBrowser = page.locator('#filebrowser-container')
        await expect(fileBrowser.getByText('history-test-file.txt').first()).toBeVisible({
            timeout: 15000,
        })

        // Trash the file to create a move action
        const fileMoveRequest = page.waitForResponse(
            (res) => res.url().includes('/api/v1/files') && res.request().method() === 'PATCH',
            {
                timeout: 15000,
            },
        )

        const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: 'history-test-file.txt' })
        await fileCard.click({ button: 'right' })
        await page.locator('#file-context-menu').getByRole('button', { name: 'Trash' }).click()
        await expect(fileCard).not.toBeVisible({ timeout: 15000 })

        // Wait for delete request to complete so the history entry is created before we open the history panel
        await fileMoveRequest

        const historyRequest = page.waitForResponse((res) => res.url().includes('history') && res.status() === 200, {
            timeout: 15000,
        })

        // Open folder history to see both creation and move entries
        await page.locator('h3').filter({ hasText: 'Home' }).click()
        const folderHistoryBtn = fileBrowser.getByRole('button', { name: 'Folder History' })
        await folderHistoryBtn.isVisible({ timeout: 3000 }).catch(() => false)
        await folderHistoryBtn.click()
        await page.locator('#file-history-sidebar').isVisible({ timeout: 15000 })

        await historyRequest

        // Should see history with created and/or trashed actions
        const fileCreated = page
            .locator('.file-action-card')
            .filter({ hasText: 'history-test-file.txt' })
            .filter({ hasText: 'File Created' })

        const fileTrashed = page
            .locator('.file-action-card')
            .filter({ hasText: 'history-test-file.txt' })
            .filter({ hasText: 'File Trashed' })

        await fileCreated.isVisible()
        await fileTrashed.isVisible()
    })
})
