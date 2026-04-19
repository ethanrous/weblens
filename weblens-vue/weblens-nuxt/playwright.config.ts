import { defineConfig } from '@playwright/test'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

const baseURL = 'http://localhost:14100'

const buildDir = path.resolve(__dirname, '../../_build/playwright')

const defaultWorkers = 8

export default defineConfig({
    testDir: './e2e/',
    testIgnore: ['setup.spec.ts'],
    outputDir: path.join(buildDir, 'test-results'),

    fullyParallel: true,
    workers: process.env.PW_WORKERS ? parseInt(process.env.PW_WORKERS) : defaultWorkers,
    retries: process.env.CI ? 2 : 1,
    maxFailures: process.env.CI ? undefined : 1,

    globalSetup: './e2e/global-setup',

    reporter: [['line']],
    timeout: 10_000,

    use: {
        baseURL,
        trace: 'on-first-retry',
        screenshot: 'only-on-failure',
    },
})
