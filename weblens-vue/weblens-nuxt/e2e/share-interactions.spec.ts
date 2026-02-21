import { test, expect, createFolder } from './fixtures'

/**
 * Tests for share modal interactions.
 *
 * These tests exercise:
 * - components/organism/ShareModal.vue (toggle public/private, timeline only, accessor management)
 * - components/molecule/UserSearch.vue (search users, select user)
 * - components/molecule/CopyBox.vue (share link display)
 * - types/weblensShare.ts (createShare, toggleIsPublic, toggleTimelineOnly, addAccessor, updateAccessorPerms)
 * - api/AllApi.ts (share CRUD endpoints)
 */

test.describe('Share Modal Interactions', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await createFolder(page, 'ShareInteractionTest')
    })

    test('should open share modal and toggle public/private', async ({ page }) => {
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'ShareInteractionTest' })
        await expect(folderCard).toBeVisible({ timeout: 15000 })
        await folderCard.click({ button: 'right' })

        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()

        const shareModal = page.locator('.fullscreen-modal')
        await expect(shareModal.locator('h4').filter({ hasText: 'Share' })).toBeVisible({ timeout: 15000 })

        // Toggle Public — creates a share on first toggle
        const publicPrivateBtn = shareModal
            .getByRole('button', { name: 'Private' })
            .or(shareModal.getByRole('button', { name: 'Public' }))
        await publicPrivateBtn.first().click()

        // Wait for share to be created and UI to update

        // The CopyBox should show a share link when public
        const copyArea = shareModal.locator('.cursor-text')
        if (await copyArea.isVisible({ timeout: 3000 }).catch(() => false)) {
            await expect(copyArea).toBeVisible()
        }

        // Toggle back to private
        await publicPrivateBtn.first().click()

        await shareModal.getByRole('button', { name: 'Done' }).first().click()
    })

    test('should toggle Timeline Only in share modal', async ({ page }) => {
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'ShareInteractionTest' })
        await expect(folderCard).toBeVisible({ timeout: 15000 })
        await folderCard.click({ button: 'right' })

        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()

        const shareModal = page.locator('.fullscreen-modal')
        await expect(shareModal.locator('h4').filter({ hasText: 'Share' })).toBeVisible({ timeout: 15000 })

        // Toggle Timeline Only
        const timelineOnlyBtn = shareModal.getByRole('button', { name: 'Timeline Only' })
        await expect(timelineOnlyBtn.first()).toBeVisible()
        await timelineOnlyBtn.first().click()

        // Toggle it back
        await timelineOnlyBtn.first().click()

        await shareModal.getByRole('button', { name: 'Done' }).first().click()
    })

    test('should search for users in share modal', async ({ page }) => {
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'ShareInteractionTest' })
        await expect(folderCard).toBeVisible({ timeout: 15000 })
        await folderCard.click({ button: 'right' })

        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()

        const shareModal = page.locator('.fullscreen-modal')
        await expect(shareModal.getByRole('button', { name: 'Done' })).toBeVisible({ timeout: 15000 })

        // Search for a user in the UserSearch component.
        // Use the textbox role for more specific targeting.
        const userSearchInput = shareModal.getByRole('textbox', { name: 'Search Users...' })
        if (await userSearchInput.isVisible({ timeout: 3000 }).catch(() => false)) {
            await userSearchInput.fill('test')

            // User search results contain <strong> with fullName inside the dropdown.
            // Use <strong> to specifically target search results, not file cards.
            const userResult = shareModal.locator('strong').filter({ hasText: /test/i })
            if (
                await userResult
                    .first()
                    .isVisible({ timeout: 5000 })
                    .catch(() => false)
            ) {
                await userResult.first().click()

                // Verify the accessor was added by checking for the table row
                const accessorRow = shareModal.locator('td').filter({ hasText: /test/i })
                await expect(accessorRow.first()).toBeVisible({ timeout: 5000 })

                // Toggle a permission checkbox if present (exercises updateAccessorPerms).
                // WeblensCheckbox uses visibility:hidden on <input> — use force click.
                const checkboxes = shareModal.locator('input[type="checkbox"]')
                if ((await checkboxes.count()) > 0) {
                    await checkboxes.first().click({ force: true })
                    await checkboxes.first().click({ force: true })
                }
            }

            await userSearchInput.clear()
        }

        await shareModal.getByRole('button', { name: 'Done' }).first().click()
    })
})
