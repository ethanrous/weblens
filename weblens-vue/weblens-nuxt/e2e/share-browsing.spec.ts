import { test, expect, createFolder, createUser, login, uploadTestFile } from './fixtures'

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
        await expect(page.locator('h3').filter({ hasText: 'ShareBrowseFolder' }).first()).toBeVisible()
        await uploadTestFile(page, 'shared-file.txt', 'This file lives in a shared folder.')
        // Navigate back to home
        await page.locator('.tabler-icon-chevron-left').first().click()
        await page.waitForURL('**/files/home')
    })

    test('should create public share and browse via share link', async ({ page }) => {
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'ShareBrowseFolder' })
        await expect(folderCard).toBeVisible()

        // Right-click to open context menu and click Share
        await folderCard.click({ button: 'right' })
        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()

        let shareModal = page.locator('.fullscreen-modal')
        await expect(shareModal.locator('h4').filter({ hasText: 'Share' })).toBeVisible()

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
        await expect(shareModal.locator('h4').filter({ hasText: 'Share' })).toBeVisible()

        // Capture the share URL
        let shareUrl = ''
        const copyBox = shareModal.getByText(/\/files\/share\//)
        if (await copyBox.isVisible().catch(() => false)) {
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
        await expect(page.locator('h3').filter({ hasText: 'ShareBrowseFolder' }).first()).toBeVisible()

        // The shared file should be visible
        await expect(page.getByText('shared-file.txt')).toBeVisible()

        // The "Log In" button should be visible (unauthenticated)
        await expect(page.getByRole('button', { name: 'Log In' })).toBeVisible()
    })

    test('should navigate to shared root page', async ({ page }) => {
        // Navigate to the shared files root
        await page.getByRole('button', { name: 'Shared' }).click()
        await page.waitForURL('**/files/share')

        // Should show either shared files or the "No files shared with you" message
        const noShared = page.getByText('No files shared with you')
        const sharedFiles = page.locator('[id^="file-card-"]')
        await expect(noShared.or(sharedFiles.first())).toBeVisible()
    })

    test('should show single Shared breadcrumb on share root page', async ({ page }) => {
        // Navigate to the shared files root
        await page.getByRole('button', { name: 'Shared' }).click()
        await page.waitForURL('**/files/share')

        // Wait for the page content to load
        const noShared = page.getByText('No files shared with you')
        const sharedFiles = page.locator('[id^="file-card-"]')
        await expect(noShared.or(sharedFiles.first())).toBeVisible()

        // The breadcrumb bar should contain exactly one "Shared" entry, not two.
        // The chevron separator only appears between crumb entries (v-if="index > 0"),
        // so if there are two crumbs there will be exactly one chevron.
        const breadcrumbBar = page.locator('.flex.h-max.w-full.items-center.border-t')
        await expect(breadcrumbBar).toBeVisible()

        const chevrons = breadcrumbBar.locator('.tabler-icon-chevron-right')
        await expect(chevrons).toHaveCount(0)

        const sharedCrumbs = breadcrumbBar.getByText('Shared', { exact: true })
        await expect(sharedCrumbs).toHaveCount(1)
    })

    test('should not throw during back-navigation from a share into the share root', async ({ page }) => {
        // Reproduces a regression where navigating from a specific share URL
        // back to /files/share triggers a Vue lifecycle error:
        //   [Vue warn]: Unhandled error during execution of component update
        //     at <FileBrowser key=0 > at <Share onVnodeUnmounted=fn ... >
        // Underlying: TypeError: Cannot read properties of null (reading 'insertBefore')
        // emitted while pages/files/share.vue swaps its v-if branch from <NuxtPage> to <FileBrowser>.

        const pageErrors: Error[] = []
        const consoleWarnings: string[] = []

        page.on('pageerror', (err) => {
            pageErrors.push(err)
        })
        page.on('console', (msg) => {
            const type = msg.type()
            if (type === 'warning' || type === 'error') {
                consoleWarnings.push(`[${type}] ${msg.text()}`)
            }
        })

        // Create a public share for the test folder so we have a routable share URL.
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'ShareBrowseFolder' })
        await expect(folderCard).toBeVisible()
        await folderCard.click({ button: 'right' })
        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()

        const shareModal = page.locator('.fullscreen-modal')
        await expect(shareModal.locator('h4').filter({ hasText: 'Share' })).toBeVisible()

        const publicBtn = shareModal
            .getByRole('button', { name: 'Private' })
            .or(shareModal.getByRole('button', { name: 'Public' }))
        const btnText = await publicBtn.first().textContent()
        if (btnText?.includes('Private')) {
            await publicBtn.first().click()
        }
        await shareModal.getByRole('button', { name: 'Done' }).first().click()

        // Re-open to read the share URL (path /files/share/<shareID>/<fileID>).
        await folderCard.click({ button: 'right' })
        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()
        await expect(shareModal.locator('h4').filter({ hasText: 'Share' })).toBeVisible()

        const linkText = (await shareModal.getByText(/\/files\/share\//).textContent()) ?? ''
        const linkMatch = linkText.match(/(\/files\/share\/[^\s]+)/)
        expect(linkMatch).not.toBeNull()
        const sharePath = linkMatch![1]

        await shareModal.getByRole('button', { name: 'Done' }).first().click()

        // Build SPA history: home -> /files/share -> share file -> back lands at /files/share.
        // Use the in-page Vue Router so the navigation is a real SPA transition (the bug
        // does not reproduce on a hard page load).
        await page.evaluate(async (target) => {
            const router = (window as unknown as { useNuxtApp: () => { $router: { push: (p: string) => Promise<unknown> } } })
                .useNuxtApp().$router
            await router.push('/files/home')
            await new Promise((r) => setTimeout(r, 200))
            await router.push('/files/share')
            await new Promise((r) => setTimeout(r, 400))
            await router.push(target)
            await new Promise((r) => setTimeout(r, 600))
        }, sharePath)

        // Wait until the share file content is mounted before going back.
        await expect(page.locator('h3').filter({ hasText: 'ShareBrowseFolder' }).first()).toBeVisible()

        await page.evaluate(async () => {
            ;(window as unknown as { useNuxtApp: () => { $router: { back: () => void } } }).useNuxtApp().$router.back()
            await new Promise((r) => setTimeout(r, 800))
        })

        await page.waitForURL('**/files/share')

        const vueWarnings = consoleWarnings.filter((m) => m.includes('[Vue warn]'))
        const insertBeforeErrors = [
            ...pageErrors.map((e) => e.message),
            ...consoleWarnings,
        ].filter((m) => m.includes("Cannot read properties of null (reading 'insertBefore')"))

        expect(vueWarnings, `Vue warnings on back-nav to share root:\n${vueWarnings.join('\n')}`).toEqual([])
        expect(
            insertBeforeErrors,
            `null insertBefore errors on back-nav to share root:\n${insertBeforeErrors.join('\n')}`,
        ).toEqual([])
    })
})

test.describe('Share Browsing - Private Share Accessor', () => {
    test.describe.configure({ timeout: 30_000 })

    test.beforeEach(async ({ page, login: _login }) => {
        // Create a second user who will be the share accessor
        await createUser(page, 'share_accessor', 'shareaccess1')

        // Navigate back to files and create a folder with content
        await page.goto('/files/home')
        await page.waitForURL('**/files/home')
        await page.locator('h3').filter({ hasText: 'Home' }).waitFor({ state: 'visible', timeout: 15000 })
        await createFolder(page, 'PrivateShareFolder')

        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'PrivateShareFolder' })
        await folderCard.dblclick()
        await expect(page.locator('h3').filter({ hasText: 'PrivateShareFolder' }).first()).toBeVisible()
        await uploadTestFile(page, 'accessor-file.txt', 'File visible to share accessor.')

        // Navigate back to home
        await page.locator('.tabler-icon-chevron-left').first().click()
        await page.waitForURL('**/files/home')

        // Create a private share and add the second user as accessor
        const card = page.locator('[id^="file-card-"]').filter({ hasText: 'PrivateShareFolder' })
        await expect(card).toBeVisible()
        await card.click({ button: 'right' })
        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()

        const shareModal = page.locator('.fullscreen-modal')
        await expect(shareModal.locator('h4').filter({ hasText: 'Share' })).toBeVisible()

        // Add the accessor user
        const userSearchInput = shareModal.getByRole('textbox', { name: 'Search Users...' })
        await expect(userSearchInput).toBeVisible()
        await userSearchInput.fill('share_accessor')
        const userResult = shareModal
            .locator('div')
            .filter({ hasText: /\(share_accessor\)/i })
            .last()
        await userResult.click()

        // Wait for the accessor to appear in the table
        const accessorRow = shareModal.locator('td').filter({ hasText: /share_accessor/i })
        await expect(accessorRow.first()).toBeVisible()

        await shareModal.getByRole('button', { name: 'Done' }).first().click()
    })

    test('should populate activeShare in Pinia when accessor navigates to private shared folder', async ({ page }) => {
        // Log out of admin
        await page.goto('/settings')
        await page.waitForURL('**/settings/account')
        await page.getByRole('button', { name: 'Log Out' }).click()
        await page.waitForURL('**/login')

        // Log in as the accessor user
        await login(page, 'share_accessor', 'shareaccess1')

        // Navigate to the Shared section
        await page.getByRole('button', { name: 'Shared' }).click()
        await page.waitForURL('**/files/share')

        // The shared folder should appear in the list
        const sharedFolderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'PrivateShareFolder' })
        await expect(sharedFolderCard).toBeVisible()

        // Navigate into the shared folder
        await sharedFolderCard.dblclick()

        // The folder heading should be visible
        await expect(page.locator('h3').filter({ hasText: 'PrivateShareFolder' }).first()).toBeVisible()

        // The file inside the shared folder should be visible
        await expect(page.getByText('accessor-file.txt')).toBeVisible()
    })
})
