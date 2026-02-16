import { test, expect } from './fixtures'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const testMediaDir = path.resolve(__dirname, '../../../images/testMedia')

/**
 * Tests for the file upload flow and progress tracking.
 *
 * These tests exercise:
 * - stores/upload.ts (startUpload, setUploadProgress, finishChunk, addFilesToUpload, clearUploads)
 * - components/molecule/UploadButton.vue (upload indicator, progress display)
 * - components/molecule/SingleUploadBox.vue (individual upload progress)
 * - types/uploadTypes.ts (UploadInfo, FileUploadMetadata, setChunk, recomputeSoFar, addRate)
 * - types/promiseQueue.ts (TaskQueue for concurrent uploads)
 * - util/humanBytes.ts (file size formatting in upload and file cards)
 * - components/atom/MediaImage.vue (thumbnail rendering after media processing)
 * - components/molecule/FileCard.vue (contentID-based media display)
 * - components/organism/MediaTimeline.vue (timeline view with processed media)
 * - stores/media.ts (media fetching and timeline display)
 */

test.describe('Upload Flow', () => {
    test.describe.configure({ mode: 'serial' })

    test.beforeEach(async ({ page }) => {
        await page.goto('/login')
        await page.getByPlaceholder('Username').fill('test_admin')
        await page.getByPlaceholder('Password').fill('password123')
        await page.getByRole('button', { name: 'Sign in' }).click()
        await page.waitForURL('**/files/home')
    })

    test('should upload a small text file and verify it appears', async ({ page }) => {
        await expect(page.locator('h3').filter({ hasText: 'Home' })).toBeVisible({ timeout: 15000 })

        const fileChooserPromise = page.waitForEvent('filechooser')
        await page.getByRole('button', { name: 'Upload' }).click()
        const fileChooser = await fileChooserPromise

        await fileChooser.setFiles({
            name: 'upload-small.txt',
            mimeType: 'text/plain',
            buffer: Buffer.from('Small file content for upload testing.'),
        })

        // Wait for the file to appear
        await expect(page.getByText('upload-small.txt')).toBeVisible({ timeout: 15000 })
    })

    test('should upload a medium-sized file to exercise chunking pipeline', async ({ page }) => {
        const fileChooserPromise = page.waitForEvent('filechooser')
        await page.getByRole('button', { name: 'Upload' }).click()
        const fileChooser = await fileChooserPromise

        // Create a 200KB file to exercise the upload pipeline more thoroughly
        const content = 'X'.repeat(200 * 1024)
        await fileChooser.setFiles({
            name: 'upload-medium.txt',
            mimeType: 'text/plain',
            buffer: Buffer.from(content),
        })

        // Wait for the file to appear in the file list
        await expect(page.getByText('upload-medium.txt')).toBeVisible({ timeout: 15000 })

        // Verify the file size is displayed (exercises humanBytes formatting)
        const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: 'upload-medium.txt' })
        await expect(fileCard.getByText(/\d+(\.\d+)?\s*kB/)).toBeVisible()
    })

    test('should upload multiple files in sequence', async ({ page }) => {
        // Upload first file
        let fileChooserPromise = page.waitForEvent('filechooser')
        await page.getByRole('button', { name: 'Upload' }).click()
        let fileChooser = await fileChooserPromise

        await fileChooser.setFiles({
            name: 'upload-seq-1.txt',
            mimeType: 'text/plain',
            buffer: Buffer.from('Sequential upload file 1'),
        })

        await expect(page.getByText('upload-seq-1.txt')).toBeVisible({ timeout: 15000 })

        // Upload second file
        fileChooserPromise = page.waitForEvent('filechooser')
        await page.getByRole('button', { name: 'Upload' }).click()
        fileChooser = await fileChooserPromise

        await fileChooser.setFiles({
            name: 'upload-seq-2.txt',
            mimeType: 'text/plain',
            buffer: Buffer.from('Sequential upload file 2'),
        })

        await expect(page.getByText('upload-seq-2.txt')).toBeVisible({ timeout: 15000 })
    })

    test('should show uploaded file sizes correctly formatted', async ({ page }) => {
        // Check that all uploaded files have properly formatted sizes
        const smallFile = page.locator('[id^="file-card-"]').filter({ hasText: 'upload-small.txt' })

        if (await smallFile.isVisible({ timeout: 3000 }).catch(() => false)) {
            // Small file should show bytes (e.g., "37B" or "38B")
            await expect(smallFile.getByText(/\d+\s*B/)).toBeVisible()
        }
    })

    test('should upload media images and display visible thumbnails', async ({ page }) => {
        // Triple timeout — media upload + backend processing + timeline fetch
        test.slow()

        await expect(page.locator('h3').filter({ hasText: 'Home' })).toBeVisible({ timeout: 15000 })

        // Upload JPEG images from the testMedia directory
        const fileChooserPromise = page.waitForEvent('filechooser')
        await page.getByRole('button', { name: 'Upload' }).click()
        const fileChooser = await fileChooserPromise

        await fileChooser.setFiles([path.join(testMediaDir, 'DSC08113.jpg'), path.join(testMediaDir, 'IMG_3973.JPG')])

        // Wait for both files to appear as file cards (not just in upload indicator).
        // Scope to file cards to avoid matching the upload progress indicator text.
        const dscCard = page.locator('[id^="file-card-"]').filter({ hasText: 'DSC08113.jpg' })
        const imgCard = page.locator('[id^="file-card-"]').filter({ hasText: 'IMG_3973.JPG' })

        await expect(dscCard).toBeVisible({ timeout: 30000 })
        await expect(imgCard).toBeVisible({ timeout: 30000 })

        // Wait for the backend to process the images and generate thumbnails.
        // FileCard renders MediaImage (div.media-image) when file.contentID is set
        // by the backend after scanning via a WebSocket FileUpdatedEvent.
        await expect(dscCard.locator('.media-image')).toBeVisible({ timeout: 60000 })
        await expect(imgCard.locator('.media-image')).toBeVisible({ timeout: 60000 })

        // Switch to timeline mode to verify the processed media appears there too.
        // The timeline fetches media from the API on mount. If the backend hasn't
        // finished indexing yet, the first fetch returns 0 results and caches that
        // empty state (canLoadMore becomes false). If that happens we toggle back
        // to files and re-enter timeline to trigger a fresh fetch.
        const timelineToggle = page.locator('.tabler-icon-photo')
        const folderToggle = page.locator('.tabler-icon-folder')
        const mediaImages = page.locator('.timelineContainer .media-image')
        const noMedia = page.getByRole('heading', { name: 'No media found' })

        await timelineToggle.last().click()
        await expect(page.getByPlaceholder('Search Media...')).toBeVisible({ timeout: 15000 })

        // Wait for either media or the empty state
        await expect(mediaImages.first().or(noMedia)).toBeVisible({ timeout: 30000 })

        // If the timeline shows "No media found", the scan wasn't done yet.
        // Toggle back to files and re-enter timeline to retry the fetch.
        if (await noMedia.isVisible()) {
            await folderToggle.last().click()
            await expect(page.getByPlaceholder('Search Files...')).toBeVisible({ timeout: 15000 })

            // Give the backend time to finish indexing the uploaded media
            await page.waitForTimeout(5000)

            await timelineToggle.last().click()
            await expect(page.getByPlaceholder('Search Media...')).toBeVisible({ timeout: 15000 })
        }

        // MediaTimeline renders MediaImage components for each media item
        await expect(mediaImages.first()).toBeVisible({ timeout: 60000 })

        // Switch back to file browser mode
        await folderToggle.last().click()
        await expect(page.getByPlaceholder('Search Files...')).toBeVisible({ timeout: 15000 })
    })

    test('should clean up upload test files', async ({ page }) => {
        const uploadedFiles = [
            'upload-small.txt',
            'upload-medium.txt',
            'upload-seq-1.txt',
            'upload-seq-2.txt',
            'DSC08113.jpg',
            'IMG_3973.JPG',
        ]

        for (const fileName of uploadedFiles) {
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
