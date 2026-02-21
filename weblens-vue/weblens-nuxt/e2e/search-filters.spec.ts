import { test, expect, createFolder } from './fixtures'

/**
 * Tests for the file search and filtering system.
 *
 * The search system has two modes:
 * 1. Local filtering (default): Filters currently loaded files by name as the user types
 * 2. Recursive search: Makes an API call to search through all nested files
 *
 * Additionally, users can toggle regex mode and open a filter panel
 * to configure search behavior.
 *
 * Components under test:
 * - Searchbar.vue: Search input with filter panel
 * - FileSearchFilters.vue: Recursive and regex toggles
 * - FileHeader.vue: Ctrl+K keyboard shortcut to focus search
 * - stores/files.ts: doSearch, setSearchRecurively, setSearchWithRegex
 * - stores/location.ts: search custom ref
 */

test.describe('File Search', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await createFolder(page, 'SearchTestFolder')
    })

    test('should filter files locally as user types without pressing Enter', async ({ page }) => {
        // Wait for files to load
        await expect(page.getByText('SearchTestFolder')).toBeVisible({ timeout: 15000 })

        // Local filtering filters currently loaded files by name as the user types
        const searchInput = page.getByPlaceholder('Search Files...')
        await searchInput.click()
        await searchInput.fill('zzz_no_match_file_xyz')

        // Give Vue time to recompute the filtered list

        // With no matching files, file cards should disappear
        const fileCards = page.locator('[id^="file-card-"]')
        const noResults = page.getByText('No files found')
        const emptyFolder = page.getByText('This folder is empty')

        const count = await fileCards.count()
        const noResultsVisible = await noResults.isVisible({ timeout: 2000 }).catch(() => false)
        const emptyVisible = await emptyFolder.isVisible({ timeout: 1000 }).catch(() => false)

        expect(count === 0 || noResultsVisible || emptyVisible).toBeTruthy()

        // Clear search to restore file list
        await searchInput.clear()
        await expect(fileCards.first()).toBeVisible({ timeout: 15000 })
    })

    test('should perform recursive search when enabled and Enter is pressed', async ({ page }) => {
        await expect(page.getByText('SearchTestFolder')).toBeVisible({ timeout: 15000 })
        const searchInput = page.getByPlaceholder('Search Files...')

        // Open search filter panel via the filter icon
        await page.locator('.tabler-icon-filter-2').click()

        // Enable recursive search
        await expect(page.getByText('Search Recursively')).toBeVisible({ timeout: 3000 })
        await page.getByText('Search Recursively').click()

        // Close filter panel
        await page.getByRole('button', { name: 'Done' }).first().click()

        // Type search
        await searchInput.click()
        await searchInput.fill('SearchTest')

        // Wait for API response
        const searchRequest = page.waitForResponse((res) => res.url().includes('/search') && res.status() === 200, {
            timeout: 10000,
        })

        // Press Enter to trigger recursive search
        await searchInput.press('Enter')

        await searchRequest

        // Results should appear (the folder we created should match)
        const fileCards = page.locator('[id^="file-card-"]')
        expect(await fileCards.count()).toBe(1)

        // Clear search
        await searchInput.clear()
        await searchInput.press('Enter')

        // Disable recursive search to clean up
        await page.locator('.tabler-icon-filter-2').click()
        await page.getByText('Search Recursively').click()
        await page.getByRole('button', { name: 'Done' }).first().click()
    })

    test('should toggle regex mode in search filters', async ({ page }) => {
        // Open filter panel
        await page.locator('.tabler-icon-filter-2').click()

        // Toggle regex ON
        const regexToggle = page.getByText('Search using Regular Expressions')
        await expect(regexToggle).toBeVisible({ timeout: 3000 })
        await regexToggle.click()

        // Toggle regex OFF
        await regexToggle.click()

        // Close filter panel
        await page.getByRole('button', { name: 'Done' }).first().click()
    })

    test('should clear search results when input is emptied', async ({ page }) => {
        await expect(page.getByText('SearchTestFolder')).toBeVisible({ timeout: 15000 })
        const searchInput = page.getByPlaceholder('Search Files...')

        // Type a search term
        await searchInput.click()
        await searchInput.fill('SomeQuery')
        await searchInput.press('Enter')

        // Clear the input and submit empty search
        await searchInput.clear()
        await searchInput.press('Enter')

        // Files should be back to normal listing
        await expect(page.locator('[id^="file-card-"]').first()).toBeVisible({ timeout: 15000 })
    })
})
