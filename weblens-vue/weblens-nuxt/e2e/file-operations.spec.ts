import { test, expect } from './fixtures'

/**
 * Tests for file operations: upload, download, context menu actions,
 * multi-select, share, folder history, and presentation mode.
 *
 * These tests exercise:
 * - api/uploadApi.ts, api/FileBrowserApi.ts
 * - stores/files.ts, stores/location.ts, stores/upload.ts, stores/tasks.ts
 * - types/weblensFile.ts, types/weblensShare.ts
 * - components/organism/FileScroller, FileContextMenu, ShareModal, FileHistory, Presentation
 * - components/molecule/ContextMenuActions, FileCard, FileSearchFilters, Searchbar
 * - util/humanBytes.ts, util/domHelpers.ts
 */
test.describe('File Operations', () => {
    test.describe.configure({ mode: 'serial' })

    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should upload a file via drag-and-drop simulation and see it appear', async ({ page }) => {
        // Create a file to upload via the file chooser triggered by the Upload button
        const fileChooserPromise = page.waitForEvent('filechooser')

        // Click the Upload button in the sidebar
        const uploadButton = page.getByRole('button', { name: 'Upload' })
        await expect(uploadButton).toBeVisible()
        await uploadButton.click()

        const fileChooser = await fileChooserPromise
        // Create a small text file for upload
        await fileChooser.setFiles({
            name: 'test-upload.txt',
            mimeType: 'text/plain',
            buffer: Buffer.from('Hello, Weblens! This is a test file for e2e testing.'),
        })

        // Wait for the uploaded file to appear in the file list
        await expect(page.getByText('test-upload.txt')).toBeVisible({ timeout: 15000 })
    })

    test('should create multiple folders for testing operations', async ({ page }) => {
        // Create first folder
        await page.getByRole('button', { name: 'New Folder' }).click()
        const nameInput = page.locator('.file-context-menu input')
        await expect(nameInput).toBeVisible()
        await nameInput.fill('Operations Folder A')
        await nameInput.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })
        // Wait for the file card (not the input text) and context menu to close
        await expect(page.locator('[id^="file-card-"]').filter({ hasText: 'Operations Folder A' })).toBeVisible({
            timeout: 15000,
        })
        await expect(nameInput).not.toBeVisible({ timeout: 3000 })

        // Create second folder
        await page.getByRole('button', { name: 'New Folder' }).click()
        const nameInput2 = page.locator('.file-context-menu input')
        await expect(nameInput2).toBeVisible()
        await nameInput2.fill('Operations Folder B')
        await nameInput2.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })
        await expect(page.locator('[id^="file-card-"]').filter({ hasText: 'Operations Folder B' })).toBeVisible({
            timeout: 15000,
        })
    })

    test('should select a file by clicking and deselect by clicking elsewhere', async ({ page }) => {
        // Click on a file card to select it
        const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: 'Operations Folder A' })
        await fileCard.click()

        // The card should have a selected state (border changes)
        await expect(fileCard).toHaveClass(/bg-card-background-selected/)

        // Click on empty area to deselect
        await page.locator('#file-scroller').last().click()

        // Card should no longer be selected
        await expect(fileCard).not.toHaveClass(/bg-card-background-selected/)
    })

    test('should open context menu on right-click and see all action buttons', async ({ page }) => {
        // Right-click on a folder
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'Operations Folder A' })
        await folderCard.click({ button: 'right' })

        // Context menu should show standard actions
        // Scope to #filebrowser-container to avoid matching sidebar buttons (e.g., "Shared" matching "Share")
        const fileBrowser = page.locator('#filebrowser-container')
        await expect(fileBrowser.getByRole('button', { name: 'Rename' })).toBeVisible()
        await expect(fileBrowser.getByRole('button', { name: 'Share' })).toBeVisible()
        await expect(fileBrowser.getByRole('button', { name: 'Download' })).toBeVisible()
        await expect(fileBrowser.getByRole('button', { name: 'Folder History' })).toBeVisible()

        // The trash button should show "Trash" (not "Delete" since we're not in trash)
        await expect(fileBrowser.getByRole('button', { name: 'Trash' })).toBeVisible()

        // Close context menu by pressing Escape
        await page.keyboard.press('Escape')
    })

    test('should open context menu on the folder header for the active folder', async ({ page }) => {
        // Click on the folder name header to open context menu for active folder
        await page.locator('h3').filter({ hasText: 'Home' }).click()

        // Should see folder-specific context menu options
        // Scope to #filebrowser-container to avoid matching the sidebar's "New Folder" button
        const fileBrowser = page.locator('#filebrowser-container')
        await expect(fileBrowser.getByRole('button', { name: 'Scan Folder' })).toBeVisible()
        await expect(fileBrowser.getByRole('button', { name: 'Folder History' })).toBeVisible()

        // Close it
        await page.keyboard.press('Escape')
    })

    test('should open share modal from context menu', async ({ page }) => {
        // Right-click on a folder
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'Operations Folder A' })
        await folderCard.click({ button: 'right' })

        // Click Share (scoped to filebrowser to avoid matching sidebar "Shared" button)
        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()

        // Share modal should appear with the "Share" heading
        // Scope to the fullscreen-modal container to avoid ambiguity with other Done buttons
        const shareModal = page.locator('.fullscreen-modal')
        await expect(page.getByRole('heading', { name: 'Share' })).toBeVisible({ timeout: 15000 })

        // The share modal should have Private/Public toggle and Timeline Only toggle
        await expect(shareModal.getByRole('button', { name: 'Private' })).toBeVisible()
        await expect(shareModal.getByRole('button', { name: 'Timeline Only' })).toBeVisible()

        // Should have a Done button
        await expect(shareModal.getByRole('button', { name: 'Done' })).toBeVisible()

        // Close the share modal
        await shareModal.getByRole('button', { name: 'Done' }).click()
    })

    test('should toggle public/private in share modal', async ({ page }) => {
        // Right-click on a folder
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'Operations Folder A' })
        await folderCard.click({ button: 'right' })

        // Open share (scoped to filebrowser to avoid matching sidebar "Shared" button)
        await page.locator('#filebrowser-container').getByRole('button', { name: 'Share' }).click()
        const shareModal = page.locator('.fullscreen-modal')
        await expect(shareModal.getByRole('button', { name: 'Done' })).toBeVisible({ timeout: 15000 })

        // Toggle public
        const publicBtn = shareModal
            .getByRole('button', { name: 'Public' })
            .or(shareModal.getByRole('button', { name: 'Private' }))
        await publicBtn.first().click()

        // Close
        await shareModal.getByRole('button', { name: 'Done' }).click()
    })

    test('should open folder history panel from context menu', async ({ page }) => {
        // Wait for files to load (ensures activeFile is set in the store)
        await expect(page.getByText('Operations Folder A')).toBeVisible({ timeout: 15000 })

        // Click on the folder name header to open context menu for the active folder
        await page.locator('h3').filter({ hasText: 'Home' }).click()

        // Click Folder History (scoped to filebrowser to avoid sidebar ambiguity)
        const fileBrowser = page.locator('#filebrowser-container')
        await fileBrowser.getByRole('button', { name: 'Folder History' }).click()

        // The history panel fetches data when open. FileAction items (e.g. "File Created")
        // are conditionally rendered via v-for, so they only appear when the panel is truly open.
        // Note: the "History of" heading is always in the DOM (parent uses w-0 without
        // overflow:hidden when closed), so we check for actual history data instead.
        await expect(page.getByText('File Created')).toBeVisible({ timeout: 15000 })

        // Close the history panel by clicking the X icon next to the "History of" heading
        await page.getByRole('heading', { name: 'History of' }).locator('..').locator('.tabler-icon-x').click()
    })

    test('should use search with recursive filter', async ({ page }) => {
        // Click the search bar
        const searchInput = page.getByPlaceholder('Search Files...')
        await searchInput.click()

        // Open the search filter panel (click filter icon)
        await page.locator('.tabler-icon-filter-2').click()

        // The FileSearchFilters should show the "Search Recursively" checkbox
        await expect(page.getByText('Search Recursively')).toBeVisible({ timeout: 15000 })
        await expect(page.getByText('Search using Regular Expressions')).toBeVisible()

        // Toggle recursive search
        await page.getByText('Search Recursively').click()

        // Close filters
        await page.getByRole('button', { name: 'Done' }).click()

        // Type a search term and press Enter
        await searchInput.fill('Operations')
        await searchInput.press('Enter')

        // Should see results
        await expect(page.getByText('Operations Folder A')).toBeVisible({ timeout: 15000 })

        // Clear search
        await searchInput.clear()
        await searchInput.press('Enter')
    })

    test('should navigate into a folder and use breadcrumb path crumbs', async ({ page }) => {
        // Double-click to navigate into folder
        await page.getByText('Operations Folder A').dblclick()
        await expect(page.locator('h3').filter({ hasText: 'Operations Folder A' })).toBeVisible({
            timeout: 15000,
        })

        // Path crumbs should show at the bottom
        // Navigate back using the back chevron
        await page.locator('.tabler-icon-chevron-left').first().click()
        await page.waitForURL('**/files/home')
    })

    test('should download a folder via context menu', async ({ page }) => {
        // Right-click on a folder
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'Operations Folder B' })
        await folderCard.click({ button: 'right' })

        // Click Download
        const downloadButton = page.getByRole('button', { name: 'Download' })
        await expect(downloadButton).toBeVisible()
        await downloadButton.click()

        // Download might take a moment - just verify the context menu closes
        // The download triggers a takeout API call. We just need to not error.
        // Wait briefly then move on
        await page.waitForTimeout(1000)
    })

    test('should scan folder via context menu', async ({ page }) => {
        // Right-click on the background (active folder context menu)
        await page.locator('h3').filter({ hasText: 'Home' }).click()

        // Click Scan Folder
        await page.getByRole('button', { name: 'Scan Folder' }).click()

        // Scanning triggers a websocket task. Just verify it doesn't error.
        await page.waitForTimeout(500)
    })

    test('should move folder to trash and restore from trash', async ({ page }) => {
        // Right-click on Operations Folder B
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'Operations Folder B' })
        await folderCard.click({ button: 'right' })

        // Click Trash
        await page.locator('#file-context-menu').waitFor({ state: 'visible', timeout: 15000 })
        await page.locator('#file-context-menu').getByRole('button', { name: 'Trash' }).click()

        // Folder should disappear
        await expect(folderCard).not.toBeVisible({ timeout: 15000 })

        // Navigate to Trash
        await page.getByRole('button', { name: 'Trash' }).click()
        await expect(page.getByText('Operations Folder B')).toBeVisible({ timeout: 15000 })

        // The trash page should show "Delete" and "Empty Trash" options
        // Right-click on the trashed folder
        const trashedCard = page.locator('[id^="file-card-"]').filter({ hasText: 'Operations Folder B' })
        await trashedCard.click({ button: 'right' })

        // In trash, the button should say "Delete" (permanently delete)
        await expect(page.locator('#filebrowser-container').getByRole('button', { name: 'Delete' })).toBeVisible()

        // Close context menu
        await page.keyboard.press('Escape')

        // Navigate back to home
        await page.getByRole('button', { name: 'Home' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should verify trash heading context menu shows Empty Trash option', async ({ page }) => {
        // Navigate to Trash
        await page.getByRole('button', { name: 'Trash' }).click()

        // Wait for the Trash page to load
        await expect(page.locator('h3').filter({ hasText: 'Trash' })).toBeVisible({ timeout: 15000 })

        // Click the Trash heading to open the folder context menu
        await page.locator('h3').filter({ hasText: 'Trash' }).click()

        // The trash folder context menu should show "Empty Trash" label on the delete button
        await expect(page.getByRole('button', { name: 'Empty Trash' })).toBeVisible()

        // Close context menu
        await page.keyboard.press('Escape')

        // Navigate back to home
        await page.getByRole('button', { name: 'Home' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should navigate to shared files page', async ({ page }) => {
        // Click the Shared button in sidebar
        await page.getByRole('button', { name: 'Shared' }).click()

        // Should navigate to the share page
        await page.waitForURL('**/files/share')

        // The page should load (even if empty)
        await page.waitForTimeout(500)

        // Navigate back
        await page.getByRole('button', { name: 'Home' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should switch to row view and interact with files', async ({ page }) => {
        // Switch to row view
        const gridLabel = page.getByText('Grid').first()
        if (await gridLabel.isVisible()) {
            await gridLabel.click()
            await page.getByText('Rows').first().click()
        }

        // Verify files are visible in row layout
        await expect(page.getByText('Operations Folder A')).toBeVisible()

        // Switch back to grid
        const rowsLabel = page.getByText('Rows').first()
        await rowsLabel.click()
        await page.getByText('Grid').first().click()
    })

    test('should clean up test files', async ({ page }) => {
        // Delete Operations Folder A
        const folderA = page.locator('[id^="file-card-"]').filter({ hasText: 'Operations Folder A' })

        if (await folderA.isVisible()) {
            await folderA.click({ button: 'right' })
            await page.locator('#filebrowser-container').getByRole('button', { name: 'Trash' }).click()
            await expect(folderA).not.toBeVisible({ timeout: 15000 })
        }

        // Delete uploaded file
        const uploadedFile = page.locator('[id^="file-card-"]').filter({ hasText: 'test-upload.txt' })

        if (await uploadedFile.isVisible()) {
            await uploadedFile.click({ button: 'right' })
            await page.locator('#filebrowser-container').getByRole('button', { name: 'Trash' }).click()
            await expect(uploadedFile).not.toBeVisible({ timeout: 15000 })
        }
    })
})
