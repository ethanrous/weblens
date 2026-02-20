import { defineConfig } from '@playwright/test'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

const baseURL = 'http://localhost:14100'

const buildDir = path.resolve(__dirname, '../../_build/playwright')

const defaultWorkers = 4

export default defineConfig({
    testDir: './e2e/',
    testIgnore: ['setup.spec.ts'],
    outputDir: path.join(buildDir, 'test-results'),

    fullyParallel: true,
    workers: process.env.PW_WORKERS ? parseInt(process.env.PW_WORKERS) : defaultWorkers,
    // retries: process.env.CI ? 2 : 0,
    retries: 0,
    // maxFailures: process.env.CI ? undefined : 1,
    maxFailures: 1,

    globalSetup: './e2e/global-setup',

    reporter: [
        ['line'],
        [
            'monocart-reporter',
            {
                name: 'Weblens E2E Report',
                outputFile: path.join(buildDir, 'report/index.html'),
                coverage: {
                    reports: ['v8'],
                    entryFilter: (entry: { url: string }) => {
                        return entry.url.includes('/_nuxt/')
                    },
                    sourceFilter: (sourcePath: string) => {
                        return (
                            !sourcePath.includes('node_modules') &&
                            !sourcePath.includes('\x00') &&
                            !sourcePath.includes('virtual:') &&
                            !sourcePath.match('^api/')
                        )
                    },
                    sourcePath: (filePath: string) => {
                        // Strip the nuxt project root prefix to show clean relative paths
                        // like "pages/login.vue" instead of full absolute paths
                        const nuxtRoot = path.resolve(__dirname) + '/'
                        if (filePath.startsWith(nuxtRoot)) {
                            return filePath.slice(nuxtRoot.length)
                        }
                        return filePath
                    },
                },
            },
        ],
    ],
    timeout: 30_000,

    // Folder creation relies on WebSocket updates which can take variable time.
    // Set a generous default assertion timeout to handle this.
    expect: {
        timeout: 15_000,
    },

    use: {
        baseURL,
        trace: 'on-first-retry',
        screenshot: 'only-on-failure',
    },
})
