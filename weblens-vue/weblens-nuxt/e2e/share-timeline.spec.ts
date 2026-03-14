import { test, expect, createFolder, uploadTestFile } from './fixtures'

/**
 * Tests for timeline view in shared folders.
 *
 * These tests exercise:
 * - pages/files/share/[shareID]/[fileID].vue (shared folder content with timeline toggle)
 * - stores/media.ts (fetchMoreMedia with shareID for unauthenticated users)
 * - stores/location.ts (isInTimeline in share context)
 * - components/organism/MediaTimeline.vue (timeline rendering in share context)
 * - routers/api/v1/api.go (media batch endpoint auth — GET /media requires share-based access)
 *
 * Bug: Switching to timeline view inside a shared folder shows an auth error
 * ("not authenticated") instead of the timeline content. The GET /api/v1/media
 * endpoint has RequireSignIn middleware that blocks unauthenticated users even
 * when a valid shareID is provided, unlike GET /folder/{folderID} which supports
 * share-based access without authentication.
 */

test.describe('Share Timeline', () => {
    let shareUrl: string

    test.beforeEach(async ({ page, login: _login }) => {
        await createFolder(page, 'ShareTimelineFolder')

        // Navigate into the folder and upload a test file
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'ShareTimelineFolder' })
        await folderCard.dblclick()
        await expect(page.locator('h3').filter({ hasText: 'ShareTimelineFolder' })).toBeVisible({ timeout: 15000 })
        await uploadTestFile(page, 'timeline-test.txt', 'File for timeline share test.')

        // Navigate back to home
        await page.locator('.tabler-icon-chevron-left').first().click()
        await page.waitForURL('**/files/home')

        // Create a public share for the folder
        const card = page.locator('[id^="file-card-"]').filter({ hasText: 'ShareTimelineFolder' })
        await expect(card).toBeVisible({ timeout: 15000 })
        await card.click({ button: 'right' })
        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()

        const shareModal = page.locator('.fullscreen-modal')
        await expect(shareModal.locator('h4').filter({ hasText: 'Share' })).toBeVisible({ timeout: 15000 })

        // Make the share public
        const publicBtn = shareModal
            .getByRole('button', { name: 'Private' })
            .or(shareModal.getByRole('button', { name: 'Public' }))
        const btnText = await publicBtn.first().textContent()
        if (btnText?.includes('Private')) {
            await publicBtn.first().click()
        }

        // Close and re-open to get the share link
        await shareModal.getByRole('button', { name: 'Done' }).first().click()

        await card.click({ button: 'right' })
        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()

        const shareModal2 = page.locator('.fullscreen-modal')
        await expect(shareModal2.locator('h4').filter({ hasText: 'Share' })).toBeVisible({ timeout: 15000 })

        // Capture the share URL
        shareUrl = ''
        const copyBox = shareModal2.getByText(/\/files\/share\//)
        if (await copyBox.isVisible({ timeout: 5000 }).catch(() => false)) {
            const linkText = await copyBox.textContent()
            if (linkText) {
                const match = linkText.match(/(https?:\/\/[^\s]+\/files\/share\/[^\s]+)/)
                if (match) {
                    shareUrl = match[1]
                }
            }
        }

        await shareModal2.getByRole('button', { name: 'Done' }).first().click()
    })

    test('should not show auth error when switching to timeline in a public share', async ({ page }) => {
        if (!shareUrl) {
            test.skip()
            return
        }

        // Log out to become an unauthenticated user
        await page.goto('/settings')
        await page.waitForURL('**/settings/account')
        await page.getByRole('button', { name: 'Log Out' }).click()
        await page.waitForURL('**/login')

        // Navigate to the public share URL
        await page.goto(shareUrl)

        // Wait for the shared folder to load
        await expect(page.locator('h3').filter({ hasText: 'ShareTimelineFolder' })).toBeVisible({ timeout: 15000 })

        // Verify we're unauthenticated — "Log In" button should be visible
        await expect(page.getByRole('button', { name: 'Log In' })).toBeVisible()

        // Switch to timeline mode using the timeline toggle button
        const timelineToggle = page.locator('.tabler-icon-photo')
        await timelineToggle.last().click()

        // Wait for the URL to include timeline=true
        await page.waitForURL('**?timeline=true', { timeout: 5000 })

        // The search placeholder should switch to "Search Media..." indicating timeline mode is active
        await expect(page.getByPlaceholder('Search Media...')).toBeVisible({ timeout: 15000 })

        // CRITICAL ASSERTION: No authentication error should be shown.
        // The bug causes a 401 from GET /api/v1/media because RequireSignIn blocks
        // unauthenticated requests even when a valid shareID is provided.
        const authError = page.getByText('not authenticated')
        const unknownError = page.getByRole('heading', { name: 'An unknown error occurred.' })
        const forbiddenError = page.getByRole('heading', { name: 'Access Forbidden' })

        // None of these error messages should be visible
        await expect(authError).not.toBeVisible({ timeout: 5000 })
        await expect(unknownError).not.toBeVisible()
        await expect(forbiddenError).not.toBeVisible()

        // Instead, the timeline should show either media content or the "No media found" empty state.
        // In the test environment with no processed media, we expect "No media found".
        await expect(page.getByRole('heading', { name: 'No media found' })).toBeVisible({ timeout: 15000 })
    })
})
