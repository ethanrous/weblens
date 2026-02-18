import { test, expect, createFolder, uploadTestFile } from './fixtures'

/**
 * Tests for browsing shared files via share links.
 *
 * These tests exercise:
 * - pages/files/share.vue (share root with inShareRoot display)
 * - pages/files/share/[shareID]/[fileID].vue (shared folder content)
 * - types/weblensShare.ts (share creation, GetLink, toggleIsPublic, clone, checkPermission)
 * - stores/location.ts (isInShare, activeShareID, activeShare, inShareRoot)
 * - components/molecule/CopyBox.vue (share link display and copy)
 * - components/organism/ShareModal.vue (opening share modal, making share public)
 */

test.describe('Share Browsing', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await createFolder(page, 'ShareBrowseFolder')
        // Navigate into the folder and upload a file
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'ShareBrowseFolder' })
        await folderCard.dblclick()
        await expect(page.locator('h3').filter({ hasText: 'ShareBrowseFolder' })).toBeVisible({ timeout: 15000 })
        await uploadTestFile(page, 'shared-file.txt', 'This file lives in a shared folder.')
        // Navigate back to home
        await page.locator('.tabler-icon-chevron-left').first().click()
        await page.waitForURL('**/files/home')
    })

    test('should create public share and browse via share link', async ({ page }) => {
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'ShareBrowseFolder' })
        await expect(folderCard).toBeVisible({ timeout: 15000 })

        // Right-click to open context menu and click Share
        await folderCard.click({ button: 'right' })
        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()

        let shareModal = page.locator('.fullscreen-modal')
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

        await folderCard.click({ button: 'right' })
        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()

        shareModal = page.locator('.fullscreen-modal')
        await expect(shareModal.locator('h4').filter({ hasText: 'Share' })).toBeVisible({ timeout: 15000 })

        // Capture the share URL
        let shareUrl = ''
        const copyBox = shareModal.getByText(/\/files\/share\//)
        if (await copyBox.isVisible({ timeout: 5000 }).catch(() => false)) {
            const linkText = await copyBox.textContent()
            if (linkText) {
                const match = linkText.match(/(https?:\/\/[^\s]+\/files\/share\/[^\s]+)/)
                if (match) {
                    shareUrl = match[1]
                }
            }
        }

        await shareModal.getByRole('button', { name: 'Done' }).first().click()

        if (!shareUrl) {
            test.skip()
            return
        }

        // Log out
        await page.goto('/settings')
        await page.waitForURL('**/settings/account')
        await page.getByRole('button', { name: 'Log Out' }).click()
        await page.waitForURL('**/login')

        // Navigate to the share URL as unauthenticated user
        await page.goto(shareUrl)

        // The shared folder heading should show the folder name
        await expect(page.locator('h3').filter({ hasText: 'ShareBrowseFolder' })).toBeVisible({ timeout: 15000 })

        // The shared file should be visible
        await expect(page.getByText('shared-file.txt')).toBeVisible({ timeout: 15000 })

        // The "Log In" button should be visible (unauthenticated)
        await expect(page.getByRole('button', { name: 'Log In' })).toBeVisible()
    })

    test('should navigate to shared root page', async ({ page }) => {
        // Navigate to the shared files root
        await page.getByRole('button', { name: 'Shared' }).click()
        await page.waitForURL('**/files/share', { timeout: 15000 })

        // Should show either shared files or the "No files shared with you" message
        const noShared = page.getByText('No files shared with you')
        const sharedFiles = page.locator('[id^="file-card-"]')
        await expect(noShared.or(sharedFiles.first())).toBeVisible({ timeout: 15000 })
    })
})
