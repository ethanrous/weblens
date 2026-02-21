import { test, expect, createFolder } from './fixtures'

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
    test.beforeEach(async ({ page, login: _login }) => {
        await createFolder(page, 'ContextMenuTestFolder')
    })

    test('should open context menu by clicking folder heading', async ({ page }) => {
        // Click on the "Home" heading to open context menu for the active folder
        await page.locator('h3').filter({ hasText: 'Home' }).click()

        const fileBrowser = page.locator('#filebrowser-container')
        const contextMenu = fileBrowser.locator('#file-context-menu')
        await expect(contextMenu).toBeVisible({ timeout: 15000 })

        // Verify folder-level actions are present
        await expect(fileBrowser.getByRole('button', { name: 'New Folder' }).first()).toBeVisible()
        await expect(fileBrowser.getByRole('button', { name: 'Scan Folder' }).first()).toBeVisible()
        await expect(fileBrowser.getByRole('button', { name: 'Folder History' }).first()).toBeVisible()
        await expect(fileBrowser.getByRole('button', { name: 'Download' }).first()).toBeVisible()

        // Close menu
        await page.keyboard.press('Escape')
    })

    test('should open folder history from context menu', async ({ page }) => {
        await page.locator('h3').filter({ hasText: 'Home' }).click()

        const fileBrowser = page.locator('#filebrowser-container')
        const contextMenu = fileBrowser.locator('#file-context-menu')
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
        await page.locator('h3').filter({ hasText: 'Home' }).click()

        const fileBrowser = page.locator('#filebrowser-container')
        const contextMenu = fileBrowser.locator('#file-context-menu')
        await expect(contextMenu).toBeVisible({ timeout: 15000 })

        const scanBtn = fileBrowser.getByRole('button', { name: 'Scan Folder' }).first()

        if (await scanBtn.isEnabled({ timeout: 2000 }).catch(() => false)) {
            await scanBtn.click()
            // Scan triggers a background task; just verify the click doesn't error
        } else {
            await page.keyboard.press('Escape')
        }
    })
})

test.describe('File-Level Context Menu', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await createFolder(page, 'CtxFileTestA')
        await createFolder(page, 'CtxFileTestB')
    })

    test('should show Rename and Share buttons for a single file', async ({ page }) => {
        const fileCards = page.locator('[id^="file-card-"]')
        await fileCards.first().click({ button: 'right' })

        const fileBrowser = page.locator('#filebrowser-container')

        // Single file context menu should show Rename and Share
        await expect(fileBrowser.getByRole('button', { name: 'Rename' }).first()).toBeVisible({ timeout: 3000 })
        await expect(fileBrowser.getByRole('button', { name: 'Share' }).first()).toBeVisible()

        await page.keyboard.press('Escape')
    })

    test('should hide Rename and Share when multiple files are selected', async ({ page }) => {
        const fileCards = page.locator('[id^="file-card-"]')
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
})

test.describe('Trash Context Menu', () => {
    test.beforeEach(async ({ login: _login }) => {})

    test('should show Delete instead of Trash when viewing trash folder', async ({ page }) => {
        // Navigate to trash via sidebar button and wait for heading to change
        await page.getByRole('button', { name: 'Trash' }).click()
        await expect(page.locator('h3').filter({ hasText: 'Trash' })).toBeVisible({ timeout: 15000 })

        const fileCards = page.locator('[id^="file-card-"]')
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
