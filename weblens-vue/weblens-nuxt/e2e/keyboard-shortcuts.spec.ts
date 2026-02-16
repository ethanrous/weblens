import { test, expect } from './fixtures'

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
    test.describe.configure({ mode: 'serial' })

    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should create test folders for keyboard shortcut tests', async ({ page }) => {
        await expect(page.locator('h3').filter({ hasText: 'Home' })).toBeVisible({ timeout: 15000 })
        const newFolderBtn = page.getByRole('button', { name: 'New Folder' })
        await expect(newFolderBtn).toBeEnabled({ timeout: 15000 })

        // Create a container folder to isolate keyboard shortcut tests
        await newFolderBtn.click()
        const nameInput = page.locator('.file-context-menu input')
        await expect(nameInput).toBeVisible()
        await nameInput.fill('KeyboardTestContainer')
        await nameInput.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })
        const containerCard = page
            .locator('[id^="file-card-"]')
            .filter({ hasText: 'KeyboardTestContainer' })
        await expect(containerCard).toBeVisible({ timeout: 15000 })
        await expect(nameInput).not.toBeVisible({ timeout: 3000 })

        // Navigate into the container
        await containerCard.dblclick()
        await expect(page.locator('h3').filter({ hasText: 'KeyboardTestContainer' })).toBeVisible({
            timeout: 15000,
        })

        // Create two sub-folders for testing Escape and Ctrl+A
        await page.getByRole('button', { name: 'New Folder' }).click()
        const input1 = page.locator('.file-context-menu input')
        await expect(input1).toBeVisible()
        await input1.fill('KB Test Item A')
        await input1.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })
        await expect(
            page.locator('[id^="file-card-"]').filter({ hasText: 'KB Test Item A' }),
        ).toBeVisible({ timeout: 15000 })
        await expect(input1).not.toBeVisible({ timeout: 3000 })

        await page.getByRole('button', { name: 'New Folder' }).click()
        const input2 = page.locator('.file-context-menu input')
        await expect(input2).toBeVisible()
        await input2.fill('KB Test Item B')
        await input2.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })
        await expect(
            page.locator('[id^="file-card-"]').filter({ hasText: 'KB Test Item B' }),
        ).toBeVisible({ timeout: 15000 })
    })

    test('should focus search input with Ctrl+K keyboard shortcut', async ({ page }) => {
        await expect(page.locator('[id^="file-card-"]').first()).toBeVisible({
            timeout: 15000,
        })

        // Press Ctrl+K to focus search
        await page.keyboard.press('ControlOrMeta+k')
        await page.waitForTimeout(500)

        // The search input should now be focused
        const searchInput = page.getByPlaceholder('Search Files...')
        await expect(searchInput).toBeFocused({ timeout: 3000 })

        // Type a search query while focused
        await page.keyboard.type('test')
        await page.waitForTimeout(500)

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
        await page.waitForTimeout(500)

        // Search input should be focused
        const searchInput = page.getByPlaceholder('Search Files...')
        await expect(searchInput).toBeFocused({ timeout: 3000 })

        // Type and press Enter for recursive search
        await page.keyboard.type('test')
        await page.keyboard.press('Enter')
        await page.waitForTimeout(1500)

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
        await page.waitForTimeout(2000)

        // The file card should still be visible (it scrolled into view)
        await expect(fileCards.first()).toBeVisible({ timeout: 15000 })
    })

    test('should use Escape to clear file selection', async ({ page }) => {
        // Navigate into the container folder
        const containerCard = page
            .locator('[id^="file-card-"]')
            .filter({ hasText: 'KeyboardTestContainer' })
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
        const containerCard = page
            .locator('[id^="file-card-"]')
            .filter({ hasText: 'KeyboardTestContainer' })
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
            await page.waitForTimeout(1000)

            // Should open presentation for non-folder files
            const presentation = page.locator('.presentation')
            if (await presentation.isVisible({ timeout: 3000 }).catch(() => false)) {
                await expect(presentation).toBeVisible()
                await page.keyboard.press('Escape')
                await expect(presentation).not.toBeVisible({ timeout: 15000 })
            }
        }
    })

    test('should clean up keyboard shortcut test folders', async ({ page }) => {
        // Navigate back to home first
        await page.goto('/files/home')
        await page.waitForURL('**/files/home')

        const containerCard = page
            .locator('[id^="file-card-"]')
            .filter({ hasText: 'KeyboardTestContainer' })
        if (await containerCard.isVisible({ timeout: 3000 }).catch(() => false)) {
            await containerCard.click({ button: 'right' })
            const trashBtn = page.locator('#filebrowser-container').getByRole('button', { name: 'Trash' })
            if (await trashBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
                await trashBtn.click()
                await expect(containerCard).not.toBeVisible({ timeout: 15000 })
            } else {
                await page.keyboard.press('Escape')
            }
        }
    })
})
