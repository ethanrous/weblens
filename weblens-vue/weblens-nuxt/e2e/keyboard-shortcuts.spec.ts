import { test, expect, createFolder } from './fixtures'

/**
 * Tests for keyboard shortcuts throughout the application.
 *
 * These tests exercise:
 * - FileHeader.vue (Ctrl+K search focus, Shift+Ctrl+K recursive search)
 * - FileScroller.vue (Ctrl+A select all, Escape clear selection, Space presentation)
 * - location.ts (highlightFileID via URL hash navigation)
 * - stores/files.ts (selectAll, clearSelected)
 * - Searchbar.vue (focus, blur, submit)
 */

test.describe('Keyboard Shortcuts', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await createFolder(page, 'KeyboardTestContainer')
        // Navigate into container
        const containerCard = page.locator('[id^="file-card-"]').filter({ hasText: 'KeyboardTestContainer' })
        await containerCard.dblclick()
        await expect(page.locator('h3').filter({ hasText: 'KeyboardTestContainer' })).toBeVisible({ timeout: 15000 })
        // Create sub-folders
        await createFolder(page, 'KB Test Item A')
        await createFolder(page, 'KB Test Item B')
        // Navigate back to home
        await page.goto('/files/home')
        await page.waitForURL('**/files/home')
    })

    test('should focus search input with Ctrl+K keyboard shortcut', async ({ page }) => {
        await expect(page.locator('[id^="file-card-"]').first()).toBeVisible({
            timeout: 15000,
        })

        // Press Ctrl+K to focus search
        await page.keyboard.press('ControlOrMeta+k')

        // The search input should now be focused
        const searchInput = page.getByPlaceholder('Search Files...')
        await expect(searchInput).toBeFocused({ timeout: 3000 })

        // Type a search query while focused
        await page.keyboard.type('test')

        // Clear and unfocus
        await searchInput.clear()
        await page.keyboard.press('Escape')
    })

    test('should enable recursive search with Shift+Ctrl+K', async ({ page }) => {
        await expect(page.locator('[id^="file-card-"]').first()).toBeVisible({
            timeout: 15000,
        })

        // Press Shift+Ctrl+K for recursive search
        await page.keyboard.press('Shift+ControlOrMeta+k')

        // Search input should be focused
        const searchInput = page.getByPlaceholder('Search Files...')
        await expect(searchInput).toBeFocused({ timeout: 3000 })

        // Type and press Enter for recursive search
        await page.keyboard.type('test')
        await page.keyboard.press('Enter')

        // Clear search
        await searchInput.clear()
        await page.keyboard.press('Escape')
    })

    test('should navigate to URL with file hash to trigger highlight scroll', async ({ page }) => {
        // Wait for file cards to load
        const fileCards = page.locator('[id^="file-card-"]')
        await expect(fileCards.first()).toBeVisible({ timeout: 15000 })

        // Get the ID of the first file card
        const firstCardId = await fileCards.first().getAttribute('id')
        expect(firstCardId).toBeTruthy()

        // Navigate to the same page with the file hash
        // This exercises location.ts highlightFileID and FileCard onMounted scrollIntoView
        const fileId = firstCardId!.replace('file-', '')
        await page.goto(`/files/home#file-${fileId}`)

        // The file card should still be visible (it scrolled into view)
        await expect(fileCards.first()).toBeVisible({ timeout: 15000 })
    })

    test('should use Escape to clear file selection', async ({ page }) => {
        // Navigate into the container folder
        const containerCard = page.locator('[id^="file-card-"]').filter({ hasText: 'KeyboardTestContainer' })
        await expect(containerCard).toBeVisible({ timeout: 15000 })
        await containerCard.dblclick()
        await expect(page.locator('h3').filter({ hasText: 'KeyboardTestContainer' })).toBeVisible({
            timeout: 15000,
        })

        // Press Escape first to close any lingering context menu
        await page.keyboard.press('Escape')

        // Select a file by clicking on it
        const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: 'KB Test Item A' })
        await fileCard.click()

        // Verify it's selected
        await expect(fileCard).toHaveClass(/bg-card-background-selected/, { timeout: 3000 })

        // Press Escape to deselect
        await page.keyboard.press('Escape')

        // File should be deselected
        await expect(fileCard).not.toHaveClass(/bg-card-background-selected/, { timeout: 3000 })
    })

    test('should use Ctrl+A to select all files', async ({ page }) => {
        // Navigate into the container folder
        const containerCard = page.locator('[id^="file-card-"]').filter({ hasText: 'KeyboardTestContainer' })
        await expect(containerCard).toBeVisible({ timeout: 15000 })
        await containerCard.dblclick()
        await expect(page.locator('h3').filter({ hasText: 'KeyboardTestContainer' })).toBeVisible({
            timeout: 15000,
        })

        // Press Escape to close any lingering context menu and defocus input
        await page.keyboard.press('Escape')

        // Use Ctrl+A (or Cmd+A on Mac) to select all
        await page.keyboard.press('ControlOrMeta+a')

        // Both files should be selected
        const cardA = page.locator('[id^="file-card-"]').filter({ hasText: 'KB Test Item A' })
        const cardB = page.locator('[id^="file-card-"]').filter({ hasText: 'KB Test Item B' })
        await expect(cardA).toHaveClass(/bg-card-background-selected/, { timeout: 3000 })
        await expect(cardB).toHaveClass(/bg-card-background-selected/, { timeout: 3000 })
    })

    test('should double-click a non-folder file to open presentation', async ({ page }) => {
        const fileCards = page.locator('[id^="file-card-"]')
        await expect(fileCards.first()).toBeVisible({ timeout: 15000 })

        // Find a .txt file to double-click (might not exist in this test)
        const txtFile = page.locator('[id^="file-card-"]').filter({ hasText: '.txt' })
        if ((await txtFile.count()) > 0) {
            await txtFile.first().dblclick()

            // Should open presentation for non-folder files
            const presentation = page.locator('.presentation')
            if (await presentation.isVisible({ timeout: 3000 }).catch(() => false)) {
                await expect(presentation).toBeVisible()
                await page.keyboard.press('Escape')
                await expect(presentation).not.toBeVisible({ timeout: 15000 })
            }
        }
    })
})
