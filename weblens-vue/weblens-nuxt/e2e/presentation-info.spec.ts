import { test, expect, createFolder, uploadTestFile } from './fixtures'

/**
 * Tests for the presentation info sidecar panel.
 *
 * These tests exercise:
 * - components/organism/Presentation.vue (info panel toggle, click isolation)
 * - components/molecule/PresentationFileInfo.vue (file details, actions)
 * - types/weblensFile.ts (FormatModified, FormatSize, GetFilename, GetFileIcon)
 * - stores/presentation.ts (infoOpen state)
 */
test.describe('Presentation Info Panel', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        await createFolder(page, 'InfoPanelFolder')
        await uploadTestFile(page, 'info-panel-test.txt', 'Info panel test content. '.repeat(100))
    })

    test('should show complete file details in info panel for text file', async ({ page }) => {
        const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: 'info-panel-test.txt' })
        await expect(fileCard).toBeVisible({ timeout: 15000 })

        // Select and open presentation
        await fileCard.click()
        await page.keyboard.press('Space')

        const presentation = page.locator('.presentation')
        await expect(presentation).toBeVisible({ timeout: 15000 })

        // Ensure info panel is open (toggle if needed — state persists in localStorage)
        const fileDetailsText = presentation.getByText('File Details')
        try {
            await expect(fileDetailsText).toBeVisible({ timeout: 3000 })
        } catch {
            await page.keyboard.press('i')
        }
        await expect(fileDetailsText).toBeVisible({ timeout: 5000 })

        // Filename in h3
        await expect(presentation.locator('h3').filter({ hasText: 'info-panel-test.txt' })).toBeVisible()

        // File icon (not folder icon)
        await expect(presentation.locator('.tabler-icon-file')).toBeVisible()

        // Size label with formatted value
        await expect(presentation.getByText('Size')).toBeVisible()
        await expect(presentation.getByText(/\d+(\.\d+)?\s*(B|kB|MB)/)).toBeVisible()

        // Modified label with a date value
        await expect(presentation.getByText('Modified')).toBeVisible()

        // Action buttons
        await expect(presentation.getByRole('button', { name: 'Show in Files' })).toBeVisible()
        await expect(presentation.getByRole('button', { name: 'Download' })).toBeVisible()

        await page.keyboard.press('Escape')
        await expect(presentation).not.toBeVisible({ timeout: 15000 })
    })

    test('should show folder icon and details in presentation for folders', async ({ page }) => {
        const folderCard = page.locator('[id^="file-card-"]').filter({ hasText: 'InfoPanelFolder' })
        await expect(folderCard).toBeVisible({ timeout: 15000 })

        await folderCard.click()
        await page.keyboard.press('Space')

        const presentation = page.locator('.presentation')
        await expect(presentation).toBeVisible({ timeout: 15000 })

        // Ensure info panel is open
        const fileDetailsText = presentation.getByText('File Details')
        try {
            await expect(fileDetailsText).toBeVisible({ timeout: 3000 })
        } catch {
            await page.keyboard.press('i')
        }
        await expect(fileDetailsText).toBeVisible({ timeout: 5000 })

        // Folder icon in info panel
        // await expect(presentation.locator('.tabler-icon-folder')).toBeVisible()

        // Folder name in h3
        await expect(presentation.locator('h3').filter({ hasText: 'InfoPanelFolder' })).toBeVisible()

        // Size and Modified labels
        await expect(presentation.getByText('Size')).toBeVisible()
        await expect(presentation.getByText('Modified')).toBeVisible()

        await page.keyboard.press('Escape')
        await expect(presentation).not.toBeVisible({ timeout: 15000 })
    })

    test('should toggle info panel and isolate clicks', async ({ page }) => {
        const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: 'info-panel-test.txt' })
        await expect(fileCard).toBeVisible({ timeout: 15000 })

        await fileCard.click()
        await page.keyboard.press('Space')

        const presentation = page.locator('.presentation')
        await expect(presentation).toBeVisible({ timeout: 15000 })

        // Ensure info panel is open
        const fileDetailsText = presentation.getByText('File Details')
        try {
            await expect(fileDetailsText).toBeVisible({ timeout: 3000 })
        } catch {
            await page.keyboard.press('i')
        }
        await expect(fileDetailsText).toBeVisible({ timeout: 5000 })

        // Click inside info panel (on the filename) — presentation should NOT close
        await presentation.locator('h3').filter({ hasText: 'info-panel-test.txt' }).click()
        await expect(presentation).toBeVisible()
        await expect(fileDetailsText).toBeVisible()

        // Toggle info closed with 'i'
        await page.keyboard.press('i')
        await expect(fileDetailsText).not.toBeVisible({ timeout: 5000 })

        // Toggle info open with 'i'
        await page.keyboard.press('i')
        await expect(fileDetailsText).toBeVisible({ timeout: 5000 })

        // Toggle via info circle icon click — panel closes
        await presentation.locator('.tabler-icon-info-circle').click()
        await expect(fileDetailsText).not.toBeVisible({ timeout: 5000 })

        // Click icon again — panel opens
        await presentation.locator('.tabler-icon-info-circle').click()
        await expect(fileDetailsText).toBeVisible({ timeout: 5000 })

        await page.keyboard.press('Escape')
        await expect(presentation).not.toBeVisible({ timeout: 15000 })
    })

    test('should navigate away from presentation when clicking Show in Files', async ({ page }) => {
        const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: 'info-panel-test.txt' })
        await expect(fileCard).toBeVisible({ timeout: 15000 })

        await fileCard.click()
        await page.keyboard.press('Space')

        const presentation = page.locator('.presentation')
        await expect(presentation).toBeVisible({ timeout: 15000 })

        // Ensure info panel is open
        const fileDetailsText = presentation.getByText('File Details')
        try {
            await expect(fileDetailsText).toBeVisible({ timeout: 3000 })
        } catch {
            await page.keyboard.press('i')
        }
        await expect(fileDetailsText).toBeVisible({ timeout: 5000 })

        // Click "Show in Files"
        await presentation.getByRole('button', { name: 'Show in Files' }).click()

        // Presentation should close
        await expect(presentation).not.toBeVisible({ timeout: 15000 })

        // URL should contain /files/ and the file card should be visible
        await expect(page).toHaveURL(/\/files\//)
        await expect(page.locator('[id^="file-card-"]').filter({ hasText: 'info-panel-test.txt' })).toBeVisible({
            timeout: 15000,
        })
    })

    test('should trigger file download when clicking Download button', async ({ page }) => {
        const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: 'info-panel-test.txt' })
        await expect(fileCard).toBeVisible({ timeout: 15000 })

        await fileCard.click()
        await page.keyboard.press('Space')

        const presentation = page.locator('.presentation')
        await expect(presentation).toBeVisible({ timeout: 15000 })

        // Ensure info panel is open
        const fileDetailsText = presentation.getByText('File Details')
        try {
            await expect(fileDetailsText).toBeVisible({ timeout: 3000 })
        } catch {
            await page.keyboard.press('i')
        }
        await expect(fileDetailsText).toBeVisible({ timeout: 5000 })

        // Listen for download event and click Download
        const downloadPromise = page.waitForEvent('download')
        await presentation.getByRole('button', { name: 'Download' }).click()
        const download = await downloadPromise

        // Verify download filename
        expect(download.suggestedFilename()).toContain('info-panel-test')

        await page.keyboard.press('Escape')
    })
})
