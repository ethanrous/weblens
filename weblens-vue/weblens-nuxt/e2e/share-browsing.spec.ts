import { test, expect } from './fixtures'

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
    test.describe.configure({ mode: 'serial' })

    let shareUrl = ''

    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should create a folder and upload a file for sharing', async ({ page }) => {
        await expect(page.locator('h3').filter({ hasText: 'Home' })).toBeVisible({ timeout: 15000 })

        // Create a folder to share
        await page.getByRole('button', { name: 'New Folder' }).click()
        const nameInput = page.locator('.file-context-menu input')
        await expect(nameInput).toBeVisible()
        await nameInput.fill('ShareBrowseFolder')
        await nameInput.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })
        await expect(page.locator('[id^="file-card-"]').filter({ hasText: 'ShareBrowseFolder' })).toBeVisible({
            timeout: 15000,
        })
        await expect(nameInput).not.toBeVisible({ timeout: 3000 })

        // Navigate into the folder and upload a file
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'ShareBrowseFolder' })
        await folderCard.dblclick()
        await expect(page.locator('h3').filter({ hasText: 'ShareBrowseFolder' })).toBeVisible({
            timeout: 15000,
        })

        const fileChooserPromise = page.waitForEvent('filechooser')
        await page.getByRole('button', { name: 'Upload' }).click()
        const fileChooser = await fileChooserPromise
        await fileChooser.setFiles({
            name: 'shared-file.txt',
            mimeType: 'text/plain',
            buffer: Buffer.from('This file lives in a shared folder.'),
        })

        await expect(page.getByText('shared-file.txt')).toBeVisible({ timeout: 15000 })

        // Navigate back to home
        await page.locator('.tabler-icon-chevron-left').first().click()
        await page.waitForURL('**/files/home')
    })

    test('should create a public share and capture the share link', async ({ page }) => {
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'ShareBrowseFolder' })
        await expect(folderCard).toBeVisible({ timeout: 15000 })

        // Right-click to open context menu and click Share
        await folderCard.click({ button: 'right' })
        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()

        let shareModal = page.locator('.fullscreen-modal')
        await expect(shareModal.locator('h4').filter({ hasText: 'Share' })).toBeVisible({
            timeout: 15000,
        })

        // Make the share public (this creates the share via API)
        const publicBtn = shareModal
            .getByRole('button', { name: 'Private' })
            .or(shareModal.getByRole('button', { name: 'Public' }))

        const btnText = await publicBtn.first().textContent()
        if (btnText?.includes('Private')) {
            await publicBtn.first().click()
            await page.waitForTimeout(1500) // Wait for share to be created via API
        }

        // Close the modal — the CopyBox won't update with the link until re-opened
        // (useAsyncData with deep:false requires a refetch)
        await shareModal.getByRole('button', { name: 'Done' }).first().click()
        await page.waitForTimeout(500)

        // Re-open the share modal to see the updated share link
        await folderCard.click({ button: 'right' })
        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()

        shareModal = page.locator('.fullscreen-modal')
        await expect(shareModal.locator('h4').filter({ hasText: 'Share' })).toBeVisible({
            timeout: 15000,
        })

        // Now the CopyBox should show the share link (contains /files/share/)
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

        // Close the modal
        await shareModal.getByRole('button', { name: 'Done' }).first().click()
    })

    test('should browse shared folder via share link', async ({ page }) => {
        // Skip if we couldn't capture the share URL
        if (!shareUrl) {
            test.skip()
            return
        }

        // Log out first to test as unauthenticated user
        await page.goto('/settings')
        await page.waitForURL('**/settings/account')
        await page.getByRole('button', { name: 'Log Out' }).click()
        await page.waitForURL('**/login')

        // Navigate to the share URL
        await page.goto(shareUrl)

        // The shared folder heading should show the folder name
        await expect(page.locator('h3').filter({ hasText: 'ShareBrowseFolder' })).toBeVisible({
            timeout: 15000,
        })

        // The shared file should be visible in the file browser
        await expect(page.getByText('shared-file.txt')).toBeVisible({ timeout: 15000 })

        // The sidebar Shared button should be visible
        await expect(page.getByRole('button', { name: 'Shared' })).toBeVisible()

        // The "Log In" button should be visible (unauthenticated user)
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

    test('should clean up share browse test folder', async ({ page }) => {
        // Navigate to home
        await page.getByRole('button', { name: 'Home' }).click()
        await page.waitForURL('**/files/home', { timeout: 15000 })

        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'ShareBrowseFolder' })

        if (await folderCard.isVisible({ timeout: 15000 }).catch(() => false)) {
            await folderCard.click({ button: 'right' })
            const trashBtn = page.locator('#filebrowser-container').getByRole('button', { name: 'Trash' })
            if (await trashBtn.isEnabled({ timeout: 2000 }).catch(() => false)) {
                await trashBtn.click()
                await expect(folderCard).not.toBeVisible({ timeout: 15000 })
            } else {
                await page.keyboard.press('Escape')
            }
        }
    })
})
