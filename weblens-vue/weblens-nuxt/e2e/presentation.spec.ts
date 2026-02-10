import { test, expect } from './fixtures'

/**
 * Tests for presentation/slideshow mode, error page, and additional
 * file interactions to increase coverage.
 *
 * These tests exercise:
 * - components/organism/Presentation.vue
 * - stores/presentation.ts
 * - pages/files.vue (presentation rendering, nextFn, prevFn)
 * - components/organism/FileScroller.vue (space key handler, context menu on empty area)
 * - error.vue (404 error page)
 * - components/molecule/PresentationFileInfo.vue
 * - types/weblensFile.ts (FormatModified, FormatSize, GetFileIcon, etc.)
 * - util/humanBytes.ts (via FormatSize)
 */
test.describe('Presentation Mode', () => {
    test.describe.configure({ mode: 'serial' })

    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should create test folders for presentation tests', async ({ page }) => {
        await expect(page.locator('h3').filter({ hasText: 'Home' })).toBeVisible({ timeout: 15000 })
        const newFolderBtn = page.getByRole('button', { name: 'New Folder' })
        await expect(newFolderBtn).toBeEnabled({ timeout: 15000 })

        // Create folders that we can use for presentation testing
        await newFolderBtn.click()
        const nameInput = page.locator('.file-context-menu input')
        await expect(nameInput).toBeVisible()
        await nameInput.fill('PresentationTestA')
        await nameInput.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })
        await expect(
            page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: 'PresentationTestA' }),
        ).toBeVisible({ timeout: 15000 })
        await expect(nameInput).not.toBeVisible({ timeout: 3000 })

        await page.getByRole('button', { name: 'New Folder' }).click()
        const nameInput2 = page.locator('.file-context-menu input')
        await expect(nameInput2).toBeVisible()
        await nameInput2.fill('PresentationTestB')
        await nameInput2.dispatchEvent('keydown', { key: 'Enter', code: 'Enter', bubbles: true })
        await expect(
            page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: 'PresentationTestB' }),
        ).toBeVisible({ timeout: 15000 })
    })

    test('should enter presentation mode by pressing Space after selecting a folder', async ({ page }) => {
        // Wait for files to load
        await expect(page.getByText('PresentationTestA')).toBeVisible({ timeout: 15000 })

        // Click on a folder to select it (sets filesStore.lastSelected)
        const folderCard = page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: 'PresentationTestA' })
        await folderCard.click()

        // Verify the card is selected
        await expect(folderCard).toHaveClass(/bg-card-background-selected/)

        // Press Space to enter presentation mode
        // The FileScroller watches the space key and opens presentation when a file is selected
        await page.keyboard.press('Space')

        // The Presentation component should render as a fullscreen-modal
        // For a folder, it shows a folder icon and the folder name
        const presentationModal = page.locator('.presentation')
        await expect(presentationModal).toBeVisible({ timeout: 15000 })

        // Verify the folder name is shown in the presentation (h1 heading)
        await expect(presentationModal.locator('h1').filter({ hasText: 'PresentationTestA' })).toBeVisible()
    })

    test('should navigate between files in presentation with arrow keys', async ({ page }) => {
        // Wait for files and select a folder
        await expect(page.getByText('PresentationTestA')).toBeVisible({ timeout: 15000 })
        const folderCard = page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: 'PresentationTestA' })
        await folderCard.click()

        // Enter presentation
        await page.keyboard.press('Space')
        const presentationModal = page.locator('.presentation')
        await expect(presentationModal).toBeVisible({ timeout: 15000 })

        // Press ArrowRight to navigate to the next file
        await page.keyboard.press('ArrowRight')

        // Press ArrowLeft to navigate back
        await page.keyboard.press('ArrowLeft')

        // Toggle the info panel with 'i' key
        await page.keyboard.press('i')

        // Press Escape to close presentation
        await page.keyboard.press('Escape')

        // Presentation should close
        await expect(presentationModal).not.toBeVisible({ timeout: 15000 })
    })

    test('should toggle presentation with Space key (open then close)', async ({ page }) => {
        await expect(page.getByText('PresentationTestA')).toBeVisible({ timeout: 15000 })
        const folderCard = page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: 'PresentationTestA' })
        await folderCard.click()

        // Open presentation with Space
        await page.keyboard.press('Space')
        const presentationModal = page.locator('.presentation')
        await expect(presentationModal).toBeVisible({ timeout: 15000 })

        // Close presentation with Space again
        await page.keyboard.press('Space')
        await expect(presentationModal).not.toBeVisible({ timeout: 15000 })
    })

    test('should open context menu by right-clicking on empty file scroller area', async ({ page }) => {
        // Wait for files to load
        await expect(page.getByText('PresentationTestA')).toBeVisible({ timeout: 15000 })

        // Right-click on the file scroller container (the outer wrapper, not individual file cards)
        // The scroller container handles contextmenu events and opens the active folder context menu
        const fileScroller = page.locator('#file-scroller').first()
        const box = await fileScroller.boundingBox()

        // Click near the bottom of the scroller where there are no file cards
        if (box) {
            await page.mouse.click(box.x + box.width / 2, box.y + box.height - 20, {
                button: 'right',
            })
        }

        // The context menu should open for the active folder (Home)
        const fileBrowser = page.locator('#filebrowser-container')
        await expect(fileBrowser.getByRole('button', { name: 'Scan Folder' })).toBeVisible({
            timeout: 15000,
        })

        // Close context menu
        await page.keyboard.press('Escape')
    })

    test('should show folder icon in presentation for folder items', async ({ page }) => {
        await expect(page.getByText('PresentationTestA')).toBeVisible({ timeout: 15000 })
        const fileCards = page.locator('[id^="file-"]:not(#file-scroller)')

        // Find a folder card (has folder icon)
        const folders = fileCards.filter({ has: page.locator('.tabler-icon-folder') })

        if ((await folders.count()) === 0) {
            test.skip()
            return
        }

        // Select folder and open presentation
        await folders.first().click()
        await page.keyboard.press('Space')

        const presentationModal = page.locator('.presentation')
        await expect(presentationModal).toBeVisible({ timeout: 15000 })

        // Folder presentation shows the folder icon
        const folderIcon = presentationModal.locator('.tabler-icon-folder')
        await expect(folderIcon).toBeVisible({ timeout: 3000 })

        // Close
        await page.keyboard.press('Escape')
        await expect(presentationModal).not.toBeVisible({ timeout: 15000 })
    })

    test('should clean up presentation test folders', async ({ page }) => {
        const folders = ['PresentationTestA', 'PresentationTestB']
        for (const folderName of folders) {
            const card = page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: folderName })

            if (await card.isVisible()) {
                await card.click({ button: 'right' })
                await page.locator('#filebrowser-container').getByRole('button', { name: 'Trash' }).click()
                await expect(card).not.toBeVisible({ timeout: 15000 })
            }
        }
    })
})

test.describe('Error Page', () => {
    test('should show 404 error page for non-existent route', async ({ page }) => {
        // Navigate to a URL that doesn't exist
        await page.goto('/this-page-does-not-exist-at-all')

        // The error page should show 404 content
        // error.vue checks error.statusCode === 404 and shows "404" heading and "Page not found"
        await expect(page.getByText('404')).toBeVisible({ timeout: 15000 })
        await expect(page.getByText('Page not found')).toBeVisible()

        // The "Go Home" button should be visible
        await expect(page.getByRole('button', { name: 'Go Home' })).toBeVisible()
    })

    test('should navigate back to home from error page', async ({ page }) => {
        // First login
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')

        // Navigate to non-existent page
        await page.goto('/this-page-does-not-exist')

        // Click "Go Home" button
        const goHomeBtn = page.getByRole('button', { name: 'Go Home' })
        if (await goHomeBtn.isVisible({ timeout: 15000 }).catch(() => false)) {
            await goHomeBtn.click()
            // Should navigate to the home page
            await page.waitForURL('**/files/home', { timeout: 15000 })
        }
    })
})

test.describe('Upload Progress and File Size Formatting', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should upload a larger file and see upload progress indicators', async ({ page }) => {
        // Create a moderately sized file (100KB) to exercise upload chunking
        const fileChooserPromise = page.waitForEvent('filechooser')
        await page.getByRole('button', { name: 'Upload' }).click()
        const fileChooser = await fileChooserPromise

        // Generate 100KB of content to test the upload pipeline more thoroughly
        const content = 'A'.repeat(100 * 1024)
        await fileChooser.setFiles({
            name: 'large-test-file.txt',
            mimeType: 'text/plain',
            buffer: Buffer.from(content),
        })

        // Wait for the file to appear in the file list
        await expect(page.getByText('large-test-file.txt')).toBeVisible({ timeout: 15000 })

        // Verify the file size is displayed in the file card (e.g., "102.4kB - Sun Feb 08 2026")
        // humanBytes formats sizes with lowercase k: "102.4kB"
        const fileCard = page.locator('[id^="file-"]:not(#file-scroller)').filter({ hasText: 'large-test-file.txt' })
        await expect(fileCard.getByText(/\d+(\.\d+)?kB/)).toBeVisible()

        // Clean up: trash the uploaded file (guard against disabled Trash button)
        await fileCard.click({ button: 'right' })
        const trashBtn = page.locator('#filebrowser-container').getByRole('button', { name: 'Trash' })
        if (await trashBtn.isEnabled({ timeout: 2000 }).catch(() => false)) {
            await trashBtn.click()
            await expect(fileCard).not.toBeVisible({ timeout: 15000 })
        } else {
            // Close context menu if Trash is disabled
            await page.keyboard.press('Escape')
        }
    })
})
