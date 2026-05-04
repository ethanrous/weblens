import { test, expect, createFolder } from './fixtures'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const testMediaDir = path.resolve(__dirname, '../../../images/testMedia')

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
        await expect(page.getByText('History of')).toBeVisible()

        // Should show at least one history action (folder was just created)
        await expect(page.getByText('ContextMenuTestFolder Created0 sec.ago')).toBeVisible()

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

test.describe('FolderPickerModal via Set as Cover', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        // Need a folder so we have a non-home target to pick as the cover destination.
        await createFolder(page, 'CoverTargetFolder')
    })

    test('opens picker at the image parent and sets the folder cover', async ({ page }) => {
        test.slow() // media upload + scan + autocomplete

        // Upload a real image so the file ends up with hasMedia + contentID,
        // which is what enables the "Set as Cover" action.
        const fileChooserPromise = page.waitForEvent('filechooser')
        await page.getByRole('button', { name: 'Upload' }).click()
        const fileChooser = await fileChooserPromise
        await fileChooser.setFiles([path.join(testMediaDir, 'DSC08113.jpg')])

        const imageCard = page.locator('[id^="file-card-"]').filter({ hasText: 'DSC08113.jpg' })
        await expect(imageCard).toBeVisible({ timeout: 15_000 })
        // Wait for the backend to scan the media — the thumbnail rendering means
        // contentID has been set on the file via WebSocket FileUpdatedEvent.
        await expect(imageCard.locator('.media-image-lowres')).toBeVisible({ timeout: 15_000 })

        // Reload so the folder is re-fetched via the GET endpoint, which populates
        // hasMedia from the media map. Without this, the in-memory file keeps
        // hasMedia=false (the WS update doesn't carry hasMedia) and the
        // "Set as Cover" button stays hidden.
        await page.reload()
        await expect(page.locator('h3').filter({ hasText: 'Home' })).toBeVisible({ timeout: 15_000 })

        const reloadedImageCard = page.locator('[id^="file-card-"]').filter({ hasText: 'DSC08113.jpg' })
        await expect(reloadedImageCard).toBeVisible({ timeout: 15_000 })
        await reloadedImageCard.click({ button: 'right' })

        const fileBrowser = page.locator('#filebrowser-container')
        const setCoverBtn = fileBrowser.getByRole('button', { name: 'Set as Cover' })
        await expect(setCoverBtn).toBeVisible({ timeout: 15_000 })
        await setCoverBtn.click()

        // FolderPickerModal opens. The input is a WeblensInput, so the actual <input>
        // is nested. Match by placeholder text the modal sets.
        const pickerInput = page.getByPlaceholder('File path...')
        await expect(pickerInput).toBeVisible({ timeout: 5_000 })

        // The image lives at USERS:<user>/DSC08113.jpg, so its parent is the home
        // directory. The picker should open showing "~/" as the suggested path.
        await expect(pickerInput).toHaveValue('~/')

        // The newly-created folder should be listed as a child of home (i.e. the
        // initial autocomplete call must have actually been made for the home dir,
        // not skipped or pointed at the wrong path).
        // Scope to the modal's scrollable folder list and match cursor-pointer rows.
        const folderRow = page
            .locator('div.cursor-pointer')
            .filter({ hasText: /^CoverTargetFolder$/ })
        await expect(folderRow.first()).toBeVisible({ timeout: 10_000 })

        // Navigate into the folder by clicking it. This exercises navigateInto(),
        // which currently throws ReferenceError on toDisplayPath.
        await folderRow.first().click()
        await expect(pickerInput).toHaveValue('~/CoverTargetFolder/')

        // The "Select" button on the current-folder row commits the cover.
        const selectBtn = page.getByRole('button', { name: 'Select' })
        await expect(selectBtn).toBeVisible()
        await selectBtn.click()

        // Modal closes after a successful set.
        await expect(pickerInput).not.toBeVisible({ timeout: 10_000 })
    })
})
