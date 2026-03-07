import { test, expect, createFolder } from './fixtures'

/**
 * Tests for the file tagging system.
 *
 * Components under test:
 * - TagManager.vue, TagSelector.vue, FileSearchFilters.vue
 * - FileSidebar.vue, ContextMenuActions.vue, FileSquare.vue
 * - stores/tags.ts (uses generated @ethanrous/weblens-api)
 */

test.describe('Tag Manager', () => {
    test.beforeEach(async ({ login: _login }) => {})

    test('should open tag manager from sidebar', async ({ page }) => {
        const sidebar = page.locator('#global-left-sidebar')
        await sidebar.getByRole('button', { name: 'Tags' }).click()

        // Tag Manager heading should appear in the fullscreen-modal overlay
        const modal = page.locator('.fullscreen-modal')
        await expect(modal.locator('h4', { hasText: 'Tags' })).toBeVisible({ timeout: 5000 })
    })

    test('should create a new tag', async ({ page }) => {
        const sidebar = page.locator('#global-left-sidebar')
        await sidebar.getByRole('button', { name: 'Tags' }).click()

        const modal = page.locator('.fullscreen-modal')
        await expect(modal.locator('h4', { hasText: 'Tags' })).toBeVisible({ timeout: 5000 })

        await modal.getByRole('button', { name: 'New Tag' }).click()
        await modal.getByPlaceholder('Tag name').fill('Favorites')
        await modal.getByRole('button', { name: 'Create' }).click()

        // The tag should appear in the modal list with a file count
        await expect(modal.getByText('Favorites', { exact: true })).toBeVisible({ timeout: 5000 })
        await expect(modal.getByText('0 files')).toBeVisible()
    })

    test('should close tag manager with Escape', async ({ page }) => {
        const sidebar = page.locator('#global-left-sidebar')
        await sidebar.getByRole('button', { name: 'Tags' }).click()

        const modal = page.locator('.fullscreen-modal')
        await expect(modal.locator('h4', { hasText: 'Tags' })).toBeVisible({ timeout: 5000 })

        await page.keyboard.press('Escape')

        // The fullscreen-modal gets pointer-events-none + opacity-0 when closed
        await expect(modal).toHaveClass(/pointer-events-none/, { timeout: 3000 })
    })

    test('should delete a tag', async ({ page }) => {
        const sidebar = page.locator('#global-left-sidebar')
        await sidebar.getByRole('button', { name: 'Tags' }).click()

        const modal = page.locator('.fullscreen-modal')
        await expect(modal.locator('h4', { hasText: 'Tags' })).toBeVisible({ timeout: 5000 })

        // Create a tag first
        await modal.getByRole('button', { name: 'New Tag' }).click()
        await modal.getByPlaceholder('Tag name').fill('ToDelete')
        await modal.getByRole('button', { name: 'Create' }).click()
        await expect(modal.getByText('ToDelete', { exact: true })).toBeVisible({ timeout: 5000 })

        // Hover over the tag row to reveal delete icon
        await modal.getByText('ToDelete', { exact: true }).hover()

        // Click the trash icon within the modal
        await modal.locator('.tabler-icon-trash').first().click()

        // Tag should be removed from the modal
        await expect(modal.getByText('ToDelete', { exact: true })).not.toBeVisible({ timeout: 5000 })
    })
})

test.describe('Tag Assignment via Context Menu', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await createFolder(page, 'TagTestFolder')
    })

    test('should show Tags button in context menu', async ({ page }) => {
        const fileCards = page.locator('[id^="file-card-"]')
        await fileCards.first().click({ button: 'right' })

        const contextMenu = page.locator('#file-context-menu')
        await expect(contextMenu.getByRole('button', { name: 'Tags' })).toBeVisible({ timeout: 3000 })

        await page.keyboard.press('Escape')
    })

    test('should open tag selector from context menu', async ({ page }) => {
        const fileCards = page.locator('[id^="file-card-"]')
        await fileCards.first().click({ button: 'right' })

        const contextMenu = page.locator('#file-context-menu')
        await contextMenu.getByRole('button', { name: 'Tags' }).click()

        // Tag selector replaces actions — check for the "New Tag" button which is unique to the selector
        await expect(contextMenu.getByRole('button', { name: 'New Tag' })).toBeVisible({ timeout: 5000 })

        await page.keyboard.press('Escape')
    })

    test('should create and assign tag from context menu', async ({ page }) => {
        const fileCards = page.locator('[id^="file-card-"]')
        await fileCards.first().click({ button: 'right' })

        const contextMenu = page.locator('#file-context-menu')
        await contextMenu.getByRole('button', { name: 'Tags' }).click()

        // Create a new tag from inside the selector
        await contextMenu.getByRole('button', { name: 'New Tag' }).click()
        await contextMenu.getByPlaceholder('Tag name').fill('Important')
        await contextMenu.getByRole('button', { name: 'Create' }).click()

        // The tag should now appear in the selector
        await expect(contextMenu.getByText('Important', { exact: true })).toBeVisible({ timeout: 5000 })

        await page.keyboard.press('Escape')
    })
})

test.describe('Tag Display on File Cards', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await createFolder(page, 'TagDisplayFolder')
    })

    test('should show tag indicator on tagged files in grid view', async ({ page }) => {
        const fileCards = page.locator('[id^="file-card-"]')
        await fileCards.first().click({ button: 'right' })

        const contextMenu = page.locator('#file-context-menu')
        await contextMenu.getByRole('button', { name: 'Tags' }).click()

        await contextMenu.getByRole('button', { name: 'New Tag' }).click()
        await contextMenu.getByPlaceholder('Tag name').fill('Starred')
        await contextMenu.getByRole('button', { name: 'Create' }).click()
        await expect(contextMenu.getByText('Starred', { exact: true })).toBeVisible({ timeout: 5000 })

        await page.keyboard.press('Escape')

        // The file card should show a colored tag dot (via title attribute on the dot container)
        const taggedCard = fileCards.first()
        const tagDot = taggedCard.locator('[title="Starred"]')
        await expect(tagDot).toBeVisible({ timeout: 5000 })
    })
})
