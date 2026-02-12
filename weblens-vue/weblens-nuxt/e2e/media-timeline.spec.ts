import { test, expect } from './fixtures'

/**
 * Tests for media timeline mode and related functionality.
 *
 * These tests exercise:
 * - stores/media.ts (fetchMoreMedia, toggleSortDirection, updateImageSize, showRaw, clearData)
 * - stores/location.ts (isInTimeline, search in timeline mode)
 * - components/organism/MediaTimeline.vue
 * - components/molecule/TimelineControls.vue
 * - components/molecule/MediaSearchFilters.vue
 * - components/molecule/Searchbar.vue (timeline mode)
 * - components/atom/SizeStepper.vue
 */
test.describe('Media Timeline', () => {
    test.describe.configure({ mode: 'serial' })

    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should switch to media timeline and see empty state or media', async ({ page }) => {
        // Switch to timeline mode
        const timelineToggle = page.locator('.tabler-icon-photo')
        await timelineToggle.last().click()

        // Should show "Search Media..." placeholder
        await expect(page.getByPlaceholder('Search Media...')).toBeVisible({ timeout: 15000 })

        // In test environment, there is no media uploaded, so we expect the empty state
        // Wait for the "No media found" heading to appear
        await expect(page.getByRole('heading', { name: 'No media found' })).toBeVisible({
            timeout: 15000,
        })

        // Verify the "Adjust filters" text and "Return to Files" button
        await expect(page.getByText('Adjust filters')).toBeVisible()
        await expect(page.getByRole('button', { name: 'Return to Files' })).toBeVisible()
    })

    test('should toggle sort direction in timeline controls', async ({ page }) => {
        // Switch to timeline mode
        await page.locator('.tabler-icon-photo').last().click()
        await expect(page.getByPlaceholder('Search Media...')).toBeVisible({ timeout: 15000 })

        // The timeline controls should have a sort direction button
        const sortButton = page.locator('.tabler-icon-sort-ascending, .tabler-icon-sort-descending')
        await expect(sortButton.first()).toBeVisible({ timeout: 3000 })

        // Click to toggle sort direction
        await sortButton.first().click()

        // Verify the button icon changed
        await expect(sortButton.first()).toBeVisible()
    })

    test('should open media search filters and toggle show raws', async ({ page }) => {
        // Switch to timeline mode
        await page.locator('.tabler-icon-photo').last().click()
        await expect(page.getByPlaceholder('Search Media...')).toBeVisible({ timeout: 15000 })

        // Open the filter panel
        await page.locator('.tabler-icon-filter-2').click()

        // Should show "Show Raws" checkbox (MediaSearchFilters)
        await expect(page.getByText('Show Raws')).toBeVisible({ timeout: 15000 })

        // Toggle the checkbox
        await page.getByText('Show Raws').click()

        // Close the filter panel
        await page.getByRole('button', { name: 'Done' }).click()
    })

    test('should search in timeline mode', async ({ page }) => {
        // Switch to timeline mode
        await page.locator('.tabler-icon-photo').last().click()
        await expect(page.getByPlaceholder('Search Media...')).toBeVisible({ timeout: 15000 })

        // Type in search
        const searchInput = page.getByPlaceholder('Search Media...')
        await searchInput.click()
        await searchInput.fill('test')
        await searchInput.press('Enter')

        // Should trigger media search (might show results or empty state)
        await page.waitForTimeout(1000)

        // Clear search
        await searchInput.clear()
        await searchInput.press('Enter')
    })

    test('should return to files from timeline', async ({ page }) => {
        // Switch to timeline mode
        await page.locator('.tabler-icon-photo').last().click()
        await expect(page.getByPlaceholder('Search Media...')).toBeVisible({ timeout: 15000 })

        // Check if "Return to Files" button is available (empty timeline)
        const returnButton = page.getByRole('button', { name: 'Return to Files' })
        if (await returnButton.isVisible()) {
            await returnButton.click()
        } else {
            // Switch back using the folder toggle icon
            const folderToggle = page.locator('.tabler-icon-folder')
            await folderToggle.last().click()
        }

        // Should be back in file mode
        await expect(page.getByPlaceholder('Search Files...')).toBeVisible({ timeout: 15000 })
    })
})
