import { test, expect, createFolder, createUser, login, uploadTestFile } from './fixtures'

/**
 * Tests for timeline view in shared folders.
 *
 * These tests exercise:
 * - pages/files/share/[shareID]/[fileID].vue (shared folder content with timeline toggle)
 * - pages/files/share.vue (inShareRoot gating of NuxtPage vs FileScroller)
 * - stores/media.ts (fetchMoreMedia with shareID)
 * - stores/location.ts (isInTimeline, activeShare, inShareRoot in share context)
 * - components/organism/MediaTimeline.vue (timeline rendering in share context)
 * - routers/api/v1/api.go (media batch endpoint auth — GET /media requires share-based access)
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

test.describe('Share Timeline - Authenticated User', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        // Create a second user who will be the share recipient
        await createUser(page, 'share_timeline_user', 'sharetimeline1')

        // Create a folder and upload a file as admin
        await page.goto('/files/home')
        await page.waitForURL('**/files/home')
        await page.locator('h3').filter({ hasText: 'Home' }).waitFor({ state: 'visible', timeout: 15000 })
        await createFolder(page, 'SharedTimelineTest')

        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'SharedTimelineTest' })
        await folderCard.dblclick()
        await expect(page.locator('h3').filter({ hasText: 'SharedTimelineTest' })).toBeVisible({ timeout: 15000 })
        await uploadTestFile(page, 'share-timeline-file.txt', 'Content for share timeline test.')

        // Navigate back to home
        await page.locator('.tabler-icon-chevron-left').first().click()
        await page.waitForURL('**/files/home')

        // Create a share and add the second user as accessor
        const card = page.locator('[id^="file-card-"]').filter({ hasText: 'SharedTimelineTest' })
        await expect(card).toBeVisible({ timeout: 15000 })
        await card.click({ button: 'right' })
        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()

        const shareModal = page.locator('.fullscreen-modal')
        await expect(shareModal.locator('h4').filter({ hasText: 'Share' })).toBeVisible({ timeout: 15000 })

        // Add the second user as accessor
        const userSearchInput = shareModal.getByRole('textbox', { name: 'Search Users...' })
        if (await userSearchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
            await userSearchInput.fill('share_timeline_user')
            const userResult = shareModal
                .locator('div')
                .filter({ hasText: /\(share_timeline_user\)/i })
                .last()
            await userResult.click({ timeout: 5000 })

            // Wait for the accessor to appear
            const accessorRow = shareModal.locator('td').filter({ hasText: /share_timeline_user/i })
            await expect(accessorRow.first()).toBeVisible({ timeout: 5000 })
        }

        await shareModal.getByRole('button', { name: 'Done' }).first().click()
    })

    test('should show timeline instead of "No files shared with you" when switching to timeline in a shared folder', async ({
        page,
    }) => {
        // Log out of admin and log in as the share recipient
        await page.goto('/settings')
        await page.waitForURL('**/settings/account')
        await page.getByRole('button', { name: 'Log Out' }).click()
        await page.waitForURL('**/login')
        await login(page, 'share_timeline_user', 'sharetimeline1')

        // Navigate to the "Shared" section via the sidebar
        await page.getByRole('button', { name: 'Shared' }).click()
        await page.waitForURL('**/files/share', { timeout: 15000 })

        // The shared folder should be listed
        const sharedFolderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'SharedTimelineTest' })
        await expect(sharedFolderCard).toBeVisible({ timeout: 15000 })

        // Navigate into the shared folder (in-app navigation)
        await sharedFolderCard.dblclick()
        await expect(page.locator('h3').filter({ hasText: 'SharedTimelineTest' })).toBeVisible({ timeout: 15000 })

        // Verify the shared file is visible
        await expect(page.getByText('share-timeline-file.txt')).toBeVisible({ timeout: 15000 })

        // Switch to timeline mode
        const timelineToggle = page.locator('.tabler-icon-photo')
        await timelineToggle.last().click()

        // Wait for the URL to include timeline=true
        await page.waitForURL('**timeline=true', { timeout: 5000 })

        // The search placeholder should switch to "Search Media..." indicating timeline mode
        await expect(page.getByPlaceholder('Search Media...')).toBeVisible({ timeout: 15000 })

        const noSharedMsg = page.getByText('No files shared with you')
        await expect(noSharedMsg).not.toBeVisible({ timeout: 5000 })

        // The timeline should show either media content or the "No media found" empty state.
        await expect(page.getByRole('heading', { name: 'No media found' })).toBeVisible({ timeout: 15000 })
    })
})
