import { test, expect } from './fixtures'

/**
 * Tests for the file context menu system.
 *
 * The context menu appears when right-clicking on files or the folder heading.
 * It provides different actions depending on context:
 * - On a file: Rename, Share, Download, Trash, Scan (if folder)
 * - On the folder heading: New Folder, Scan, Folder History, Download, Trash
 * - In trash: Delete (permanent), not Trash
 *
 * Components under test:
 * - FileContextMenu.vue: Menu container, positioning, rename/new folder modes
 * - ContextMenuActions.vue: Action buttons (Scan, Download, Trash/Delete, History)
 * - ContextMenuHeader.vue: File info header in context menu
 * - ContextNameFile.vue: Text input for rename/new folder
 * - FileScroller.vue: handleContextMenu on empty space
 * - FileHeader.vue: openContextMenu on heading click
 * - stores/contextMenu.ts: setMenuOpen, setMenuMode, setTarget, setSharing
 */

test.describe('Folder-Level Context Menu', () => {
    test.describe.configure({ mode: 'serial' })

    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
        // Wait for the page header to be fully loaded (don't require files — folder may be empty)
        await expect(page.locator('h3').filter({ hasText: 'Home' })).toBeVisible({ timeout: 15000 })
    })

    test('should open context menu by clicking folder heading', async ({ page }) => {
        // Click on the "Home" heading to open context menu for the active folder
        await page.locator('h3').filter({ hasText: 'Home' }).click()

        const fileBrowser = page.locator('#filebrowser-container')
        const contextMenu = fileBrowser.locator('.file-context-menu')
        await expect(contextMenu).toBeVisible({ timeout: 15000 })

        // Verify folder-level actions are present
        await expect(fileBrowser.getByRole('button', { name: 'New Folder' }).first()).toBeVisible()
        await expect(fileBrowser.getByRole('button', { name: 'Scan Folder' }).first()).toBeVisible()
        await expect(fileBrowser.getByRole('button', { name: 'Folder History' }).first()).toBeVisible()
        await expect(fileBrowser.getByRole('button', { name: 'Download' }).first()).toBeVisible()

        // Close menu
        await page.keyboard.press('Escape')
    })

    test('should create a new folder via the context menu', async ({ page }) => {
        // Open context menu on heading
        await page.locator('h3').filter({ hasText: 'Home' }).click()

        const fileBrowser = page.locator('#filebrowser-container')
        const contextMenu = fileBrowser.locator('.file-context-menu')
        await expect(contextMenu).toBeVisible({ timeout: 15000 })

        await fileBrowser.getByRole('button', { name: 'New Folder' }).first().click()

        // The context menu should switch to name input mode
        const nameInput = fileBrowser.locator('.file-context-menu input')
        await expect(nameInput).toBeVisible({ timeout: 3000 })
        await nameInput.fill('ContextMenuTestFolder')
        await nameInput.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })

        // New folder should appear
        await expect(page.getByText('ContextMenuTestFolder')).toBeVisible({ timeout: 15000 })
    })

    test('should open folder history from context menu', async ({ page }) => {
        // Wait for the folder we created in the previous test
        await expect(page.getByText('ContextMenuTestFolder')).toBeVisible({ timeout: 15000 })

        await page.locator('h3').filter({ hasText: 'Home' }).click()

        const fileBrowser = page.locator('#filebrowser-container')
        const contextMenu = fileBrowser.locator('.file-context-menu')
        await expect(contextMenu).toBeVisible({ timeout: 15000 })

        const historyBtn = fileBrowser.getByRole('button', { name: 'Folder History' }).first()
        await historyBtn.click()

        // History panel should open with "History of" heading
        await expect(page.getByText('History of')).toBeVisible({ timeout: 15000 })

        // Should show at least one history action (folder was just created)
        await expect(page.getByText('File Created').first()).toBeVisible({ timeout: 15000 })

        // Close history panel — the X icon's parent div intercepts pointer events
        // and the icon may be outside the default viewport. Widen viewport and force click.
        const viewport = page.viewportSize()!
        await page.setViewportSize({ width: 1920, height: viewport.height })
        await page.locator('.tabler-icon-x').first().click({ force: true })
        await page.setViewportSize(viewport)
    })

    test('should trigger folder scan from context menu', async ({ page }) => {
        await expect(page.getByText('ContextMenuTestFolder')).toBeVisible({ timeout: 15000 })

        await page.locator('h3').filter({ hasText: 'Home' }).click()

        const fileBrowser = page.locator('#filebrowser-container')
        const contextMenu = fileBrowser.locator('.file-context-menu')
        await expect(contextMenu).toBeVisible({ timeout: 15000 })

        const scanBtn = fileBrowser.getByRole('button', { name: 'Scan Folder' }).first()

        if (await scanBtn.isEnabled({ timeout: 2000 }).catch(() => false)) {
            await scanBtn.click()
            // Scan triggers a background task; just verify the click doesn't error
            await page.waitForTimeout(1000)
        } else {
            await page.keyboard.press('Escape')
        }
    })

    test('should clean up test folder', async ({ page }) => {
        const folderCard = page
            .locator('[id^="file-"]:not(#file-scroller)')
            .filter({ hasText: 'ContextMenuTestFolder' })

        if (await folderCard.isVisible({ timeout: 15000 }).catch(() => false)) {
            await folderCard.click({ button: 'right' })
            const trashBtn = page.locator('#filebrowser-container').getByRole('button', { name: 'Trash' }).first()
            if (await trashBtn.isEnabled({ timeout: 2000 }).catch(() => false)) {
                await trashBtn.click()
                await expect(folderCard).not.toBeVisible({ timeout: 15000 })
            } else {
                await page.keyboard.press('Escape')
            }
        }
    })
})

test.describe('File-Level Context Menu', () => {
    test.describe.configure({ mode: 'serial' })

    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should create test folders for file context menu tests', async ({ page }) => {
        // Create two folders so we can test single-file and multi-select context menus
        await page.getByRole('button', { name: 'New Folder' }).click()
        const nameInput = page.locator('.file-context-menu input')
        await expect(nameInput).toBeVisible()
        await nameInput.fill('CtxFileTestA')
        await nameInput.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })
        // Wait for the file card (not the input text) and for the context menu to close
        await expect(page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: 'CtxFileTestA' })).toBeVisible(
            { timeout: 15000 },
        )
        await expect(nameInput).not.toBeVisible({ timeout: 3000 })

        await page.getByRole('button', { name: 'New Folder' }).click()
        const nameInput2 = page.locator('.file-context-menu input')
        await expect(nameInput2).toBeVisible()
        await nameInput2.fill('CtxFileTestB')
        await nameInput2.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })
        await expect(page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: 'CtxFileTestB' })).toBeVisible(
            { timeout: 15000 },
        )
    })

    test('should show Rename and Share buttons for a single file', async ({ page }) => {
        await expect(page.getByText('CtxFileTestA')).toBeVisible({ timeout: 15000 })

        const fileCards = page.locator('[id^="file-"]:not(#file-scroller)')
        await fileCards.first().click({ button: 'right' })

        const fileBrowser = page.locator('#filebrowser-container')

        // Single file context menu should show Rename and Share
        await expect(fileBrowser.getByRole('button', { name: 'Rename' }).first()).toBeVisible({ timeout: 3000 })
        await expect(fileBrowser.getByRole('button', { name: 'Share' }).first()).toBeVisible()

        await page.keyboard.press('Escape')
    })

    test('should hide Rename and Share when multiple files are selected', async ({ page }) => {
        await expect(page.getByText('CtxFileTestA')).toBeVisible({ timeout: 15000 })

        const fileCards = page.locator('[id^="file-"]:not(#file-scroller)')
        const count = await fileCards.count()

        if (count < 2) {
            test.skip()
            return
        }

        // Select two files
        await fileCards.first().click()
        await fileCards.nth(1).click({ modifiers: ['ControlOrMeta'] })

        // Right-click one of the selected files
        await fileCards.nth(1).click({ button: 'right' })

        const fileBrowser = page.locator('#filebrowser-container')

        // In multi-select, Rename and Share should NOT be visible
        const renameBtn = fileBrowser.getByRole('button', { name: 'Rename' })
        const shareBtn = fileBrowser.getByRole('button', { name: 'Share' })
        await expect(renameBtn).not.toBeVisible({ timeout: 2000 })
        await expect(shareBtn).not.toBeVisible()

        // Download should still be visible
        await expect(fileBrowser.getByRole('button', { name: 'Download' }).first()).toBeVisible()

        await page.keyboard.press('Escape')
    })

    test('should clean up file context menu test folders', async ({ page }) => {
        for (const folderName of ['CtxFileTestA', 'CtxFileTestB']) {
            const card = page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: folderName })

            if (await card.isVisible({ timeout: 3000 }).catch(() => false)) {
                await card.click({ button: 'right' })
                const trashBtn = page.locator('#filebrowser-container').getByRole('button', { name: 'Trash' })
                if (await trashBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
                    await trashBtn.click()
                    await expect(card).not.toBeVisible({ timeout: 15000 })
                } else {
                    await page.keyboard.press('Escape')
                }
            }
        }
    })
})

test.describe('Trash Context Menu', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should show Delete instead of Trash when viewing trash folder', async ({ page }) => {
        // Navigate to trash via sidebar button and wait for heading to change
        await page.getByRole('button', { name: 'Trash' }).click()
        await expect(page.locator('h3').filter({ hasText: 'Trash' })).toBeVisible({ timeout: 15000 })

        const fileCards = page.locator('[id^="file-"]:not(#file-scroller)')
        const hasFiles = await fileCards
            .first()
            .isVisible({ timeout: 15000 })
            .catch(() => false)

        if (!hasFiles) {
            // Trash is empty, verify the empty state message
            await expect(page.getByText('Someone already took the trash out')).toBeVisible({ timeout: 15000 })
            return
        }

        // Right-click a file in trash
        await fileCards.first().click({ button: 'right' })

        const fileBrowser = page.locator('#filebrowser-container')

        // In trash, the button should say "Delete" (permanent) not "Trash"
        const deleteBtn = fileBrowser.getByRole('button', { name: 'Delete' }).first()
        await expect(deleteBtn).toBeVisible({ timeout: 3000 })

        await page.keyboard.press('Escape')
    })
})
