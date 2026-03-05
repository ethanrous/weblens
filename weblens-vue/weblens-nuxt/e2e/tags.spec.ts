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
        // Click the Tags button in the left sidebar
        const sidebar = page.locator('#global-left-sidebar')
        await sidebar.getByRole('button', { name: 'Tags' }).click()

        // Tag Manager modal should appear
        await expect(page.getByRole('heading', { name: 'Tags' })).toBeVisible({ timeout: 5000 })
    })

    test('should create a new tag', async ({ page }) => {
        const sidebar = page.locator('#global-left-sidebar')
        await sidebar.getByRole('button', { name: 'Tags' }).click()

        await expect(page.getByRole('heading', { name: 'Tags' })).toBeVisible({ timeout: 5000 })

        // Click "New Tag" button
        await page.getByRole('button', { name: 'New Tag' }).click()

        // Fill in the tag name
        await page.getByPlaceholder('Tag name').fill('Favorites')

        // Click Create
        await page.getByRole('button', { name: 'Create' }).click()

        // The tag should appear in the list
        await expect(page.getByText('Favorites')).toBeVisible({ timeout: 5000 })
        await expect(page.getByText('0 files')).toBeVisible()
    })

    test('should close tag manager with Escape', async ({ page }) => {
        const sidebar = page.locator('#global-left-sidebar')
        await sidebar.getByRole('button', { name: 'Tags' }).click()

        await expect(page.getByRole('heading', { name: 'Tags' })).toBeVisible({ timeout: 5000 })

        await page.keyboard.press('Escape')

        // Modal should be hidden (opacity-0)
        // The heading may still be in DOM but invisible
        await expect(page.getByRole('heading', { name: 'Tags' })).not.toBeVisible({ timeout: 3000 })
    })

    test('should delete a tag', async ({ page }) => {
        const sidebar = page.locator('#global-left-sidebar')
        await sidebar.getByRole('button', { name: 'Tags' }).click()
        await expect(page.getByRole('heading', { name: 'Tags' })).toBeVisible({ timeout: 5000 })

        // Create a tag first
        await page.getByRole('button', { name: 'New Tag' }).click()
        await page.getByPlaceholder('Tag name').fill('ToDelete')
        await page.getByRole('button', { name: 'Create' }).click()
        await expect(page.getByText('ToDelete')).toBeVisible({ timeout: 5000 })

        // Hover over the tag row to reveal delete icon
        await page.getByText('ToDelete').hover()

        // Click the trash icon (delete button)
        const trashIcon = page.locator('.tabler-icon-trash').first()
        await trashIcon.click()

        // Tag should be removed
        await expect(page.getByText('ToDelete')).not.toBeVisible({ timeout: 5000 })
    })
})

test.describe('Tag Assignment via Context Menu', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await createFolder(page, 'TagTestFolder')
    })

    test('should show Tags button in context menu', async ({ page }) => {
        const fileCards = page.locator('[id^="file-card-"]')
        await fileCards.first().click({ button: 'right' })

        const fileBrowser = page.locator('#filebrowser-container')
        const tagsBtn = fileBrowser.getByRole('button', { name: 'Tags' }).first()
        await expect(tagsBtn).toBeVisible({ timeout: 3000 })

        await page.keyboard.press('Escape')
    })

    test('should open tag selector from context menu', async ({ page }) => {
        const fileCards = page.locator('[id^="file-card-"]')
        await fileCards.first().click({ button: 'right' })

        const fileBrowser = page.locator('#filebrowser-container')
        await fileBrowser.getByRole('button', { name: 'Tags' }).first().click()

        // Tag selector should replace the actions in the context menu
        // It should show the "TAGS" header and "New Tag" button
        await expect(page.getByText('TAGS')).toBeVisible({ timeout: 5000 })
        await expect(page.getByRole('button', { name: 'New Tag' })).toBeVisible()

        await page.keyboard.press('Escape')
    })

    test('should create and assign tag from context menu', async ({ page }) => {
        const fileCards = page.locator('[id^="file-card-"]')
        await fileCards.first().click({ button: 'right' })

        const fileBrowser = page.locator('#filebrowser-container')
        await fileBrowser.getByRole('button', { name: 'Tags' }).first().click()

        // Create a new tag from inside the selector
        await page.getByRole('button', { name: 'New Tag' }).click()
        await page.getByPlaceholder('Tag name').fill('Important')
        await page.getByRole('button', { name: 'Create' }).click()

        // The tag should now appear in the selector with a check mark (auto-applied)
        await expect(page.getByText('Important')).toBeVisible({ timeout: 5000 })

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

        const fileBrowser = page.locator('#filebrowser-container')
        await fileBrowser.getByRole('button', { name: 'Tags' }).first().click()

        await page.getByRole('button', { name: 'New Tag' }).click()
        await page.getByPlaceholder('Tag name').fill('Starred')
        await page.getByRole('button', { name: 'Create' }).click()
        await expect(page.getByText('Starred')).toBeVisible({ timeout: 5000 })

        // Close context menu
        await page.keyboard.press('Escape')

        // The file card should now show a colored tag dot
        // Tag dots are small spans with rounded-full class inside the file card
        const taggedCard = fileCards.first()
        const tagDot = taggedCard.locator('.rounded-full')
        await expect(tagDot.first()).toBeVisible({ timeout: 5000 })
    })
})
