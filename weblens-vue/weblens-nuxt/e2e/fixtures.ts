import { test as base, expect, type Page } from '@playwright/test'
import { addCoverageReport } from 'monocart-reporter'
import {
    startWorkerMongo,
    stopWorkerMongo,
    startTestBackend,
    stopTestBackend,
    type WorkerMongo,
    type TestBackend,
} from './backend-manager'

const DEFAULT_ADMIN_USERNAME = 'admin'
const DEFAULT_ADMIN_PASSWORD = 'adminadmin1'

const test = base.extend<
    { autoTestFixture: unknown; testBackend: TestBackend; login: unknown },
    { workerMongo: WorkerMongo }
>({
    workerMongo: [
        // eslint-disable-next-line no-empty-pattern
        async ({}, use, workerInfo) => {
            const mongo = await startWorkerMongo(workerInfo.parallelIndex)
            await use(mongo)
            await stopWorkerMongo(mongo)
        },
        { scope: 'worker' },
    ],

    testBackend: async ({ workerMongo }, use, testInfo) => {
        const backend = await startTestBackend(testInfo.parallelIndex, workerMongo)
        // await initializeServer(backend.baseURL)
        await use(backend)
        await stopTestBackend(backend)
    },

    baseURL: async ({ testBackend }, use) => {
        await use(testBackend.baseURL)
    },

    login: [
        async ({ page, baseURL }: { page: Page; baseURL: string }, use: () => Promise<void>) => {
            await page.goto(baseURL)
            await login(page)
            await use()
        },
        { box: true },
    ],

    autoTestFixture: [
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        async ({ page }: { page: any }, use: () => unknown) => {
            await page.coverage.startJSCoverage({
                resetOnNavigation: false,
            })

            await use()

            const coverage = await page.coverage.stopJSCoverage()
            const testInfo = test.info()
            if (coverage.length > 0) {
                await addCoverageReport(coverage, testInfo)
            }
        },
        { auto: true },
    ],
})

test.beforeEach(async ({ page }) => {
    page.on('console', (msg) => {
        if (msg.type() === 'error') {
            console.error('[Browser ERROR]', msg.text())
        } else if (msg.type() === 'warning') {
            console.warn('[Browser WARNING]', msg.text())
        }
    })
})

async function login(
    page: import('@playwright/test').Page,
    username = DEFAULT_ADMIN_USERNAME,
    password = DEFAULT_ADMIN_PASSWORD,
) {
    await page.goto('/login')
    await page.getByPlaceholder('Username').fill(username)
    await page.getByPlaceholder('Password').fill(password)
    await page.getByRole('button', { name: 'Sign in' }).click()
    await page.waitForURL('**/files/home')
}

async function createFolder(page: import('@playwright/test').Page, name: string) {
    await page.waitForTimeout(500) // Wait briefly for UI to stabilize (e.g. after login or navigation)

    await page.getByRole('button', { name: 'New Folder' }).click()
    const nameInput = page.locator('.file-context-menu input')
    await expect(nameInput).toBeVisible()
    await nameInput.fill(name)
    await nameInput.dispatchEvent('keydown', {
        key: 'Enter',
        code: 'Enter',
        bubbles: true,
    })
    await expect(page.locator('[id^="file-card-"]').filter({ hasText: name })).toBeVisible({ timeout: 15000 })
    await expect(nameInput).not.toBeVisible({ timeout: 3000 })
}

async function uploadTestFile(page: import('@playwright/test').Page, name: string, content: string) {
    const fileChooserPromise = page.waitForEvent('filechooser')
    await page.getByRole('button', { name: 'Upload' }).click()
    const fileChooser = await fileChooserPromise
    await fileChooser.setFiles({
        name,
        mimeType: 'text/plain',
        buffer: Buffer.from(content),
    })
    await expect(page.locator('[id^="file-card-"]').filter({ hasText: name })).toBeVisible({ timeout: 15000 })
}

async function createUser(page: import('@playwright/test').Page, username: string, password: string) {
    await page.goto('/settings/users')
    await page.waitForURL('**/settings/users')
    await page.getByPlaceholder('Username').fill(username)
    await page.getByPlaceholder('Password').fill(password)
    await page.getByRole('button', { name: 'Create User' }).click()
    await expect(page.getByText(username)).toBeVisible({ timeout: 15000 })
}

export { test, expect, login, createFolder, uploadTestFile, createUser, DEFAULT_ADMIN_USERNAME, DEFAULT_ADMIN_PASSWORD }
