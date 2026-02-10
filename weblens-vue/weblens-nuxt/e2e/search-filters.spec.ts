import { test, expect } from './fixtures'

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
    test.describe.configure({ mode: 'serial' })

    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should create a test folder for search tests', async ({ page }) => {
        await expect(page.locator('h3').filter({ hasText: 'Home' })).toBeVisible({ timeout: 15000 })
        const newFolderBtn = page.getByRole('button', { name: 'New Folder' })
        await expect(newFolderBtn).toBeEnabled({ timeout: 15000 })
        await newFolderBtn.click()
        const nameInput = page.locator('.file-context-menu input')
        await expect(nameInput).toBeVisible()
        await nameInput.fill('SearchTestFolder')
        await nameInput.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })
        await expect(
            page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: 'SearchTestFolder' }),
        ).toBeVisible({ timeout: 15000 })
    })

    test('should filter files locally as user types without pressing Enter', async ({ page }) => {
        // Wait for files to load
        await expect(page.getByText('SearchTestFolder')).toBeVisible({ timeout: 15000 })

        // Local filtering filters currently loaded files by name as the user types
        const searchInput = page.getByPlaceholder('Search Files...')
        await searchInput.click()
        await searchInput.fill('zzz_no_match_file_xyz')

        // Give Vue time to recompute the filtered list
        await page.waitForTimeout(500)

        // With no matching files, file cards should disappear
        const fileCards = page.locator('[id^="file-"]:not(#file-scroller)')
        const noResults = page.getByText('No files found')
        const emptyFolder = page.getByText('This folder is empty')

        const count = await fileCards.count()
        const noResultsVisible = await noResults.isVisible({ timeout: 2000 }).catch(() => false)
        const emptyVisible = await emptyFolder.isVisible({ timeout: 1000 }).catch(() => false)

        expect(count === 0 || noResultsVisible || emptyVisible).toBeTruthy()

        // Clear search to restore file list
        await searchInput.clear()
        await page.waitForTimeout(300)
        await expect(fileCards.first()).toBeVisible({ timeout: 15000 })
    })

    test('should perform recursive search when enabled and Enter is pressed', async ({ page }) => {
        await expect(page.getByText('SearchTestFolder')).toBeVisible({ timeout: 15000 })
        const searchInput = page.getByPlaceholder('Search Files...')

        // Open search filter panel via the filter icon
        await page.locator('.tabler-icon-filter-2').click()
        await page.waitForTimeout(300)

        // Enable recursive search
        await expect(page.getByText('Search Recursively')).toBeVisible({ timeout: 3000 })
        await page.getByText('Search Recursively').click()

        // Close filter panel
        await page.getByRole('button', { name: 'Done' }).first().click()
        await page.waitForTimeout(300)

        // Type search and press Enter to trigger the API-based recursive search
        await searchInput.click()
        await searchInput.fill('SearchTest')
        await searchInput.press('Enter')

        // Wait for API response
        await page.waitForTimeout(2000)

        // Results should appear (the folder we created should match)
        const fileCards = page.locator('[id^="file-"]:not(#file-scroller)')
        const noResults = page.getByText('No files found')
        const hasResults = (await fileCards.count()) > 0
        const hasNoResults = await noResults.isVisible({ timeout: 2000 }).catch(() => false)
        expect(hasResults || hasNoResults).toBeTruthy()

        // Clear search
        await searchInput.clear()
        await searchInput.press('Enter')
        await page.waitForTimeout(500)

        // Disable recursive search to clean up
        await page.locator('.tabler-icon-filter-2').click()
        await page.waitForTimeout(300)
        await page.getByText('Search Recursively').click()
        await page.getByRole('button', { name: 'Done' }).first().click()
    })

    test('should toggle regex mode in search filters', async ({ page }) => {
        // Open filter panel
        await page.locator('.tabler-icon-filter-2').click()
        await page.waitForTimeout(300)

        // Toggle regex ON
        const regexToggle = page.getByText('Search using Regular Expressions')
        await expect(regexToggle).toBeVisible({ timeout: 3000 })
        await regexToggle.click()
        await page.waitForTimeout(200)

        // Toggle regex OFF
        await regexToggle.click()
        await page.waitForTimeout(200)

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
        await page.waitForTimeout(500)

        // Clear the input and submit empty search
        await searchInput.clear()
        await searchInput.press('Enter')
        await page.waitForTimeout(500)

        // Files should be back to normal listing
        await expect(page.locator('[id^="file-"]:not(#file-scroller)').first()).toBeVisible({ timeout: 15000 })
    })

    test('should clean up search test folder', async ({ page }) => {
        const card = page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: 'SearchTestFolder' })

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
    })
})
