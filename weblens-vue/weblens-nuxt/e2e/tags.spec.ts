import { test, expect, createFolder } from './fixtures'

/**
 * Tests for the file tagging system.
 *
 * Tags allow users to organize files across folders without moving them.
 * Tags have a name and color, and files can have multiple tags.
 *
 * Components under test:
 * - TagManager.vue: CRUD modal for managing tags
 * - TagSelector.vue: Tag assignment popover in context menu
 * - FileSearchFilters.vue: Tag filter checkboxes
 * - FileSidebar.vue: Tags button in sidebar
 * - ContextMenuActions.vue: Tags button in context menu
 * - FileSquare.vue / FileRow.vue: Tag dots on file cards
 * - stores/tags.ts: Tag state management
 * - api/TagApi.ts: Tag API client
 */

test.describe('Tag Manager', () => {
    test.beforeEach(async ({ login: _login }) => {})

    test('should open tag manager from sidebar', async ({ page }) => {
        const sidebar = page.locator('#global-left-sidebar')
        await sidebar.getByRole('button', { name: 'Tags' }).click()

        // Tag Manager modal should appear — scope to the fullscreen-modal overlay
        const modal = page.locator('.fullscreen-modal')
        await expect(modal.getByRole('heading', { name: 'Tags' })).toBeVisible({ timeout: 5000 })
    })

    test('should create a new tag', async ({ page }) => {
        const sidebar = page.locator('#global-left-sidebar')
        await sidebar.getByRole('button', { name: 'Tags' }).click()

        const modal = page.locator('.fullscreen-modal')
        await expect(modal.getByRole('heading', { name: 'Tags' })).toBeVisible({ timeout: 5000 })

        // Click "New Tag" button within the modal
        await modal.getByRole('button', { name: 'New Tag' }).click()

        // Fill in the tag name
        await modal.getByPlaceholder('Tag name').fill('Favorites')

        // Click Create
        await modal.getByRole('button', { name: 'Create' }).click()

        // The tag should appear in the modal list
        await expect(modal.getByText('Favorites')).toBeVisible({ timeout: 5000 })
        await expect(modal.getByText('0 files')).toBeVisible()
    })

    test('should close tag manager with Escape', async ({ page }) => {
        const sidebar = page.locator('#global-left-sidebar')
        await sidebar.getByRole('button', { name: 'Tags' }).click()

        const modal = page.locator('.fullscreen-modal')
        await expect(modal.getByRole('heading', { name: 'Tags' })).toBeVisible({ timeout: 5000 })

        await page.keyboard.press('Escape')

        // Modal should become hidden
        await expect(modal.getByRole('heading', { name: 'Tags' })).not.toBeVisible({ timeout: 3000 })
    })

    test('should delete a tag', async ({ page }) => {
        const sidebar = page.locator('#global-left-sidebar')
        await sidebar.getByRole('button', { name: 'Tags' }).click()

        const modal = page.locator('.fullscreen-modal')
        await expect(modal.getByRole('heading', { name: 'Tags' })).toBeVisible({ timeout: 5000 })

        // Create a tag first
        await modal.getByRole('button', { name: 'New Tag' }).click()
        await modal.getByPlaceholder('Tag name').fill('ToDelete')
        await modal.getByRole('button', { name: 'Create' }).click()
        await expect(modal.getByText('ToDelete')).toBeVisible({ timeout: 5000 })

        // Hover over the tag row to reveal delete icon
        await modal.getByText('ToDelete').hover()

        // Click the trash icon (delete button) within the modal
        const trashIcon = modal.locator('.tabler-icon-trash').first()
        await trashIcon.click()

        // Tag should be removed from the modal
        await expect(modal.getByText('ToDelete')).not.toBeVisible({ timeout: 5000 })
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

        // Tag selector should replace the actions in the context menu
        await expect(contextMenu.getByText('TAGS')).toBeVisible({ timeout: 5000 })
        await expect(contextMenu.getByRole('button', { name: 'New Tag' })).toBeVisible()

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
        await expect(contextMenu.getByText('Important')).toBeVisible({ timeout: 5000 })

        // Close context menu
        await page.keyboard.press('Escape')
    })
})

test.describe('Tag Display on File Cards', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await createFolder(page, 'TagDisplayFolder')
    })

    test('should show tag indicator on tagged files in grid view', async ({ page }) => {
        // First, tag a file via context menu
        const fileCards = page.locator('[id^="file-card-"]')
        await fileCards.first().click({ button: 'right' })

        const contextMenu = page.locator('#file-context-menu')
        await contextMenu.getByRole('button', { name: 'Tags' }).click()

        await contextMenu.getByRole('button', { name: 'New Tag' }).click()
        await contextMenu.getByPlaceholder('Tag name').fill('Starred')
        await contextMenu.getByRole('button', { name: 'Create' }).click()
        await expect(contextMenu.getByText('Starred')).toBeVisible({ timeout: 5000 })

        // Close context menu
        await page.keyboard.press('Escape')

        // The file card should now show a colored tag dot (via the title attribute on the dot container)
        const taggedCard = fileCards.first()
        const tagDot = taggedCard.locator('[title="Starred"]')
        await expect(tagDot).toBeVisible({ timeout: 5000 })
    })
})
