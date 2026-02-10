import { test, expect } from './fixtures'

test.describe('File Browser', () => {
    test.describe.configure({ mode: 'serial' })

    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should display sidebar with navigation buttons', async ({ page }) => {
        await expect(page.getByRole('button', { name: 'Home' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'Shared' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'Trash' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'New Folder' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'Settings' })).toBeVisible()
    })

    test('should create a new folder via sidebar', async ({ page }) => {
        await expect(page.getByRole('button', { name: 'New Folder' })).toBeEnabled({ timeout: 15000 })
        await page.getByRole('button', { name: 'New Folder' }).click()

        // The context menu opens with ContextNameFile input (auto-focused)
        const nameInput = page.locator('.file-context-menu input')
        await expect(nameInput).toBeVisible()
        await nameInput.fill('Test Folder')
        // Re-focus after fill to ensure Vue's useFocusWithin has settled —
        // fill() on search-type inputs triggers select() which can briefly
        // disrupt focus tracking in the full suite with a complex DOM.
        await nameInput.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })

        // Wait for the folder to appear in the file browser (use file card locator to avoid matching input text)
        await expect(page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: 'Test Folder' })).toBeVisible({
            timeout: 15000,
        })
    })

    test('should navigate into folder and see breadcrumbs', async ({ page }) => {
        // Double-click the folder to navigate into it - use exact match to avoid
        // matching other folders that contain "Test Folder" as a substring
        await page.getByText('Test Folder', { exact: true }).dblclick()

        // Wait for navigation — the header should now show "Test Folder"
        await expect(page.locator('h3').filter({ hasText: 'Test Folder' })).toBeVisible({
            timeout: 15000,
        })
    })

    test('should navigate back using the back chevron', async ({ page }) => {
        // First navigate into the folder
        await page.getByText('Test Folder', { exact: true }).dblclick()
        await expect(page.locator('h3').filter({ hasText: 'Test Folder' })).toBeVisible({
            timeout: 15000,
        })

        // Click the back chevron (IconChevronLeft in FileHeader)
        await page.locator('.tabler-icon-chevron-left').first().click()

        // Should be back at home — header shows "Home" or the home folder name
        await page.waitForURL('**/files/home')
    })

    test('should rename a folder via context menu', async ({ page }) => {
        // Right-click the folder to open context menu
        await page.getByText('Test Folder', { exact: true }).click({ button: 'right' })

        // Click Rename in context menu
        await page.getByRole('button', { name: 'Rename' }).click()

        // The name input should appear pre-filled with current name
        const nameInput = page.locator('.file-context-menu input')
        await expect(nameInput).toBeVisible()
        await nameInput.clear()
        await nameInput.fill('Renamed Folder')
        await nameInput.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })

        // Assert the renamed folder appears
        await expect(page.getByText('Renamed Folder')).toBeVisible({ timeout: 15000 })
        await expect(page.getByText('Test Folder')).not.toBeVisible()
    })

    test('should toggle file view between grid and row layout', async ({ page }) => {
        // The FileSortControls has a WeblensOptions for shape with Grid/Rows options.
        // The currently selected option's label is visible. Click the dropdown to open,
        // then select the other option.

        // Find the shape selector — it shows the current shape label (e.g., "Grid")
        // WeblensOptions has both a visible label and a hidden measurement span,
        // so we use .first() to target the visible one.
        const gridLabel = page.getByText('Grid').first()
        const rowsLabel = page.getByText('Rows').first()

        // If Grid is currently selected, switch to Rows
        if (await gridLabel.isVisible()) {
            await gridLabel.click()
            await rowsLabel.click()
            // File cards should now be in row layout (h-20 w-full)
            await expect(rowsLabel).toBeVisible()
        } else {
            // Switch back to Grid
            await rowsLabel.click()
            await gridLabel.click()
            await expect(gridLabel).toBeVisible()
        }
    })

    test('should change sort order', async ({ page }) => {
        // The sort condition dropdown shows Filename/Size/Date options
        // WeblensOptions has hidden measurement spans, so use .first()
        const filenameLabel = page.getByText('Filename').first()

        // Click the current sort option to open dropdown
        if (await filenameLabel.isVisible()) {
            await filenameLabel.click()
            // Select Date sort
            await page.getByText('Date').first().click()
        } else {
            // Some other sort is active, click it and switch to Filename
            const sizeLabel = page.getByText('Size').first()
            const dateLabel = page.getByText('Date').first()
            if (await sizeLabel.isVisible()) {
                await sizeLabel.click()
            } else {
                await dateLabel.click()
            }
            await filenameLabel.click()
        }

        // Toggle sort direction by clicking the sort direction button
        // (has IconSortAscending or IconSortDescending)
        const sortDirButton = page.locator('.tabler-icon-sort-ascending, .tabler-icon-sort-descending')
        await sortDirButton.first().click()
    })

    test('should search for files', async ({ page }) => {
        // Click the searchbar to focus it
        const searchInput = page.getByPlaceholder('Search Files...')
        await searchInput.click()
        await searchInput.fill('Renamed Folder')
        await searchInput.press('Enter')

        // Should show search results (the renamed folder should appear)
        await expect(page.getByText('Renamed Folder')).toBeVisible({ timeout: 15000 })

        // Clear the search
        await searchInput.click()
        await searchInput.clear()
        await searchInput.press('Enter')
    })

    test('should move folder to trash', async ({ page }) => {
        // Right-click the folder card to open context menu
        // Use :not(#file-scroller) to avoid matching the scroller container
        const folderCard = page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: 'Renamed Folder' })
        await folderCard.click({ button: 'right' })

        // Click Trash in context menu (scoped to filebrowser to avoid sidebar Trash button)
        await page.locator('#filebrowser-container').getByRole('button', { name: 'Trash' }).click()

        // The folder card should disappear from the file browser
        await expect(folderCard).not.toBeVisible({ timeout: 15000 })
    })

    test('should navigate to trash via sidebar', async ({ page }) => {
        // Click Trash button in sidebar
        await page.getByRole('button', { name: 'Trash' }).click()

        // Should see trashed items (the folder we just trashed)
        await expect(page.getByText('Renamed Folder')).toBeVisible({ timeout: 15000 })

        // Navigate back to Home
        await page.getByRole('button', { name: 'Home' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should toggle between files and media timeline', async ({ page }) => {
        // The toggle button in FileHeader has IconPhoto (when in file mode)
        // or IconFolder (when in timeline mode). Click it to switch.
        const timelineToggle = page.locator('.tabler-icon-photo, .tabler-icon-folder')
        await timelineToggle.last().click()

        // In timeline mode, the search placeholder changes to "Search Media..."
        await expect(page.getByPlaceholder('Search Media...')).toBeVisible({ timeout: 15000 })

        // Toggle back to file mode
        const folderIcon = page.locator('.tabler-icon-folder')
        await folderIcon.last().click()

        await expect(page.getByPlaceholder('Search Files...')).toBeVisible({ timeout: 15000 })
    })

    test('should navigate to settings via sidebar', async ({ page }) => {
        await page.getByRole('button', { name: 'Settings' }).click()
        await page.waitForURL('**/settings/account')
    })
})
