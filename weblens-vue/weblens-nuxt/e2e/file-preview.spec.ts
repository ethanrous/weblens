import { test, expect } from './fixtures'

/**
 * Tests for file preview/presentation with real uploaded files.
 *
 * These tests exercise:
 * - components/molecule/PresentationFileInfo.vue (file name display in presentation)
 * - components/organism/Presentation.vue (fullscreen presentation overlay)
 * - types/weblensFile.ts (GetFilename, FormatModified, FormatSize, IsFolder, GetFileIcon)
 * - util/humanBytes.ts (formatBytes for file sizes)
 * - stores/presentation.ts (setPresentationFileID, clearPresentation)
 * - stores/files.ts (getNextFileID, getPreviousFileID)
 * - components/molecule/FileCard.vue (double-click navigateToFile for non-folder)
 */

test.describe('File Preview and Presentation', () => {
    test.describe.configure({ mode: 'serial' })

    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should upload test files for preview', async ({ page }) => {
        await expect(page.locator('h3').filter({ hasText: 'Home' })).toBeVisible({ timeout: 15000 })

        // Upload a text file
        const fileChooserPromise = page.waitForEvent('filechooser')
        await page.getByRole('button', { name: 'Upload' }).click()
        const fileChooser = await fileChooserPromise

        // Create a file with known content to test size display
        const content = 'Hello from preview test! '.repeat(100)
        await fileChooser.setFiles({
            name: 'preview-test.txt',
            mimeType: 'text/plain',
            buffer: Buffer.from(content),
        })

        await expect(page.getByText('preview-test.txt')).toBeVisible({ timeout: 15000 })

        // Upload a second file for navigation testing
        const fileChooserPromise2 = page.waitForEvent('filechooser')
        await page.getByRole('button', { name: 'Upload' }).click()
        const fileChooser2 = await fileChooserPromise2

        await fileChooser2.setFiles({
            name: 'preview-test-2.txt',
            mimeType: 'text/plain',
            buffer: Buffer.from('Second preview file content.'),
        })

        await expect(page.getByText('preview-test-2.txt')).toBeVisible({ timeout: 15000 })
    })

    test('should open presentation by double-clicking a text file', async ({ page }) => {
        const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: 'preview-test.txt' })
        await expect(fileCard.first()).toBeVisible({ timeout: 15000 })

        // Double-click opens presentation for non-folder files
        await fileCard.first().dblclick()

        const presentation = page.locator('.presentation')
        await expect(presentation).toBeVisible({ timeout: 15000 })

        // The presentation should show the file name
        await expect(presentation.getByText('preview-test.txt')).toBeVisible({ timeout: 5000 })

        // Close presentation
        await page.keyboard.press('Escape')
        await expect(presentation).not.toBeVisible({ timeout: 15000 })
    })

    test('should open presentation with Space and show file info via i key', async ({ page }) => {
        const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: 'preview-test.txt' })
        await expect(fileCard.first()).toBeVisible({ timeout: 15000 })

        // Select the file
        await fileCard.first().click()
        await expect(fileCard.first()).toHaveClass(/bg-card-background-selected/)

        // Open presentation with Space
        await page.keyboard.press('Space')
        const presentation = page.locator('.presentation')
        await expect(presentation).toBeVisible({ timeout: 15000 })

        // Press 'i' to toggle the info panel
        await page.keyboard.press('i')

        // The file name should be visible in the presentation info
        await expect(presentation.getByText('preview-test')).toBeVisible({ timeout: 5000 })

        // Press 'i' again to close info panel
        await page.keyboard.press('i')

        // Close presentation
        await page.keyboard.press('Escape')
        await expect(presentation).not.toBeVisible({ timeout: 15000 })
    })

    test('should navigate between files in presentation with arrow keys', async ({ page }) => {
        // Select the first text file
        const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: 'preview-test.txt' })
        await expect(fileCard.first()).toBeVisible({ timeout: 15000 })
        await fileCard.first().click()

        // Open presentation
        await page.keyboard.press('Space')
        const presentation = page.locator('.presentation')
        await expect(presentation).toBeVisible({ timeout: 15000 })

        // Navigate to the next file
        await page.keyboard.press('ArrowRight')

        // Navigate back
        await page.keyboard.press('ArrowLeft')

        // Close with Escape
        await page.keyboard.press('Escape')
        await expect(presentation).not.toBeVisible({ timeout: 15000 })
    })

    test('should show file size in file card', async ({ page }) => {
        // The file card should display the file size (exercises humanBytes)
        const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: 'preview-test.txt' })
        await expect(fileCard.first()).toBeVisible({ timeout: 15000 })

        // The file size should be visible (formatted by humanBytes, e.g. "2.4kB")
        await expect(fileCard.first().getByText(/\d+(\.\d+)?\s*(B|kB|MB)/)).toBeVisible()
    })

    test('should clean up preview test files', async ({ page }) => {
        for (const fileName of ['preview-test.txt', 'preview-test-2.txt']) {
            const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: fileName })

            if (await fileCard.isVisible({ timeout: 3000 }).catch(() => false)) {
                await fileCard.click({ button: 'right' })
                const trashBtn = page.locator('#filebrowser-container').getByRole('button', { name: 'Trash' })
                if (await trashBtn.isEnabled({ timeout: 2000 }).catch(() => false)) {
                    await trashBtn.click()
                    await expect(fileCard).not.toBeVisible({ timeout: 15000 })
                } else {
                    await page.keyboard.press('Escape')
                }
            }
        }
    })
})
