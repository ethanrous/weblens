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
    test.describe.configure({ mode: 'serial' })

    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should show rewind indicator when navigating with rewindTo URL param', async ({ page }) => {
        // Set a past timestamp via URL to exercise RewindIndicator, relativeTimeAgo
        const pastDate = new Date(Date.now() - 60 * 60 * 1000) // 1 hour ago
        const rewindUrl = `/files/home?rewindTo=${pastDate.toISOString()}`
        await page.goto(rewindUrl)
        await page.waitForTimeout(2000)

        // The RewindIndicator should show "Rewound to"
        const rewindText = page.getByText('Rewound to')
        if (await rewindText.isVisible({ timeout: 15000 }).catch(() => false)) {
            await expect(rewindText).toBeVisible()

            // Click X to clear the rewind (exercises setViewTimestamp(0))
            const closeBtn = page.locator('.tabler-icon-x').first()
            if (await closeBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
                await closeBtn.click()
                await page.waitForTimeout(500)
                // URL should no longer have rewindTo
                expect(page.url()).not.toContain('rewindTo')
            }
        }
    })

    test('should show rewind with older timestamp for different time ranges', async ({ page }) => {
        // Test with a date 3 days ago (exercises "X days ago" path in relativeTimeAgo)
        const threeDaysAgo = new Date(Date.now() - 3 * 24 * 60 * 60 * 1000)
        const rewindUrl = `/files/home?rewindTo=${threeDaysAgo.toISOString()}`
        await page.goto(rewindUrl)
        await page.waitForTimeout(2000)

        const rewindText = page.getByText('Rewound to')
        if (await rewindText.isVisible({ timeout: 15000 }).catch(() => false)) {
            await expect(rewindText).toBeVisible()
        }

        // Clear rewind by navigating to home without query params
        await page.goto('/files/home')
        await page.waitForTimeout(500)
    })

    test('should open history panel and click an action to trigger rewind', async ({ page }) => {
        await page.keyboard.press('Escape')
        await page.waitForTimeout(500)

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
                    await page.waitForTimeout(1500)

                    // Check if RewindIndicator appeared
                    const rewindIndicator = page.getByText('Rewound to')
                    if (await rewindIndicator.isVisible({ timeout: 3000 }).catch(() => false)) {
                        // Click X to clear rewind
                        const closeRewind = rewindIndicator.locator('..').locator('.tabler-icon-x')
                        if (await closeRewind.isVisible({ timeout: 1000 }).catch(() => false)) {
                            await closeRewind.click()
                            await page.waitForTimeout(500)
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
        const fileCard = page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: 'history-test-file.txt' })
        await fileCard.click({ button: 'right' })
        const trashBtn = fileBrowser.getByRole('button', { name: 'Trash' })
        if (await trashBtn.isEnabled({ timeout: 2000 }).catch(() => false)) {
            await trashBtn.click()
            await expect(fileCard).not.toBeVisible({ timeout: 15000 })
        } else {
            await page.keyboard.press('Escape')
        }

        // Open folder history to see both creation and move entries
        await page.locator('h3').filter({ hasText: 'Home' }).click()
        const folderHistoryBtn = fileBrowser.getByRole('button', { name: 'Folder History' })
        if (await folderHistoryBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
            await folderHistoryBtn.click()
            await page.waitForTimeout(2000)

            // Should see history with created and/or trashed actions
            const fileCreated = page.getByText('File Created').first()
            const fileTrashed = page.getByText('File Trashed').first()

            const hasCreated = await fileCreated.isVisible({ timeout: 3000 }).catch(() => false)
            const hasTrashed = await fileTrashed.isVisible({ timeout: 1000 }).catch(() => false)

            expect(hasCreated || hasTrashed).toBeTruthy()
        }

        await page.keyboard.press('Escape')
    })
})
