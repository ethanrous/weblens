import { test, expect, createFolder } from './fixtures'

/**
 * Tests for file sorting controls and view mode switching.
 *
 * Users can:
 * - Sort files by name, date, or size
 * - Toggle sort direction (ascending/descending)
 * - Switch view between Grid and Rows layout
 *
 * Components under test:
 * - FileSortControls.vue: Sort condition dropdown, sort direction button, file shape dropdown
 * - WeblensOptions.vue: Dropdown component for selecting options
 * - stores/files.ts: setSortCondition, toggleSortDirection, setFileShape, saveFoldersSettings
 *
 * NOTE: Sort controls live in FileHeader.vue which is a SIBLING of #filebrowser-container,
 * so they must be located with page-level locators (not scoped to #filebrowser-container).
 *
 * WeblensOptions enters "iconOnly" mode when the control is too narrow for the label text.
 * In iconOnly mode, only the icon is visible (not the label text). We must account for this
 * by clicking sort icons rather than text labels when re-opening a dropdown after a change.
 */

test.describe('Sort and View Controls', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await createFolder(page, 'SortTestAlpha')
        await createFolder(page, 'SortTestBeta')
    })

    test('should change sort condition to Filename', async ({ page }) => {
        await expect(page.getByText('SortTestAlpha')).toBeVisible({ timeout: 15000 })

        // Sort controls are in FileHeader (outside #filebrowser-container).
        // WeblensOptions renders a visible label + a hidden measurement span.
        // In narrow layouts, iconOnly mode hides the label and only shows the icon.
        // Click the sort dropdown icon to open it — the icon is always visible.
        const sortIcon = page.locator('.tabler-icon-calendar, .tabler-icon-sort-a-z, .tabler-icon-file-analytics')
        await expect(sortIcon.first()).toBeVisible({ timeout: 3000 })
        await sortIcon.first().click()

        // When the dropdown is open, all option labels are rendered (isOpen overrides iconOnly)
        const filenameOption = page.getByText('Filename').first()
        await expect(filenameOption).toBeVisible({ timeout: 2000 })
        await filenameOption.click()

        // Files should still be visible after sort change
        await expect(page.locator('[id^="file-card-"]').first()).toBeVisible()

        // Verify sort changed — the sort-a-z icon should now be the active sort icon
        await expect(page.locator('.tabler-icon-sort-a-z').first()).toBeVisible()
    })

    test('should change sort condition to Size', async ({ page }) => {
        await expect(page.getByText('SortTestAlpha')).toBeVisible({ timeout: 15000 })

        // Click the current sort icon to open the dropdown
        const sortIcon = page.locator('.tabler-icon-calendar, .tabler-icon-sort-a-z, .tabler-icon-file-analytics')
        await expect(sortIcon.first()).toBeVisible({ timeout: 3000 })
        await sortIcon.first().click()

        // Select Size from the open dropdown
        const sizeOption = page.getByText('Size').first()
        await expect(sizeOption).toBeVisible({ timeout: 2000 })
        await sizeOption.click()

        // Files should still be visible
        await expect(page.locator('[id^="file-card-"]').first()).toBeVisible()

        // Verify sort changed — the file-analytics icon should now be active
        await expect(page.locator('.tabler-icon-file-analytics').first()).toBeVisible()
    })

    test('should change sort condition back to Date', async ({ page }) => {
        await expect(page.getByText('SortTestAlpha')).toBeVisible({ timeout: 15000 })

        // Click the current sort icon to open the dropdown
        const sortIcon = page.locator('.tabler-icon-calendar, .tabler-icon-sort-a-z, .tabler-icon-file-analytics')
        await expect(sortIcon.first()).toBeVisible({ timeout: 3000 })
        await sortIcon.first().click()

        // Select Date from the open dropdown
        const dateOption = page.getByText('Date').first()
        await expect(dateOption).toBeVisible({ timeout: 2000 })
        await dateOption.click()

        // Files should still be visible
        await expect(page.locator('[id^="file-card-"]').first()).toBeVisible()

        // Verify sort changed — the calendar icon should now be active
        await expect(page.locator('.tabler-icon-calendar').first()).toBeVisible()
    })

    test('should toggle sort direction between ascending and descending', async ({ page }) => {
        await expect(page.getByText('SortTestAlpha')).toBeVisible({ timeout: 15000 })

        // Sort direction button is in FileHeader, use page-level selector
        const sortDirIcon = page.locator('.tabler-icon-sort-ascending, .tabler-icon-sort-descending')
        await expect(sortDirIcon.first()).toBeVisible({ timeout: 3000 })

        // Click to toggle direction
        await sortDirIcon.first().click()

        // Files should still be visible after re-sort
        await expect(page.locator('[id^="file-card-"]').first()).toBeVisible()

        // Toggle back to original direction
        const sortDirIcon2 = page.locator('.tabler-icon-sort-ascending, .tabler-icon-sort-descending')
        await sortDirIcon2.first().click()
    })

    test('should switch from Grid to Rows view and back', async ({ page }) => {
        await expect(page.getByText('SortTestAlpha')).toBeVisible({ timeout: 15000 })

        // View mode controls are in FileHeader (outside #filebrowser-container).
        // The shape icons: IconLayoutGrid (.tabler-icon-layout-grid),
        // IconLayoutRows (.tabler-icon-layout-rows)
        const gridLabel = page.getByText('Grid').first()
        const rowsLabel = page.getByText('Rows').first()

        // Switch to Rows view
        if (await gridLabel.isVisible()) {
            await gridLabel.click()
            await rowsLabel.click()
        } else {
            // If currently Rows, first verify, then switch to Grid and back
            await rowsLabel.click()
            await gridLabel.click()
            await gridLabel.click()
            await rowsLabel.click()
        }

        // Files should still be visible in row layout
        await expect(page.locator('[id^="file-card-"]').first()).toBeVisible()

        // Switch back to Grid using the icon (Rows icon is always visible)
        const shapeIcon = page.locator('.tabler-icon-layout-grid, .tabler-icon-layout-rows')
        await shapeIcon.first().click()

        // Select Grid from dropdown (when open, label text is always shown)
        await page.getByText('Grid').first().click()
    })

    test('should show Columns option as disabled in shape dropdown', async ({ page }) => {
        await expect(page.getByText('SortTestAlpha')).toBeVisible({ timeout: 15000 })

        // Open the shape dropdown by clicking the current shape icon
        const shapeIcon = page.locator('.tabler-icon-layout-grid, .tabler-icon-layout-rows')
        await expect(shapeIcon.first()).toBeVisible({ timeout: 3000 })
        await shapeIcon.first().click()

        // Columns option should be visible in the dropdown but disabled (pointer-events-none)
        const columnsOption = page.getByText('Columns').first()
        await expect(columnsOption).toBeVisible({ timeout: 2000 })

        // Verify the Columns option has the disabled styling (pointer-events-none)
        const columnsWrapper = columnsOption.locator('..')
        await expect(columnsWrapper).toHaveClass(/pointer-events-none/)

        // Close the dropdown by clicking the shape icon again
        await shapeIcon.first().click()
    })
})
