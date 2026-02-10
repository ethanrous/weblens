import { defineConfig } from '@playwright/test'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

const port = process.env.WEBLENS_TEST_PORT ?? '14100'
const baseURL = `http://localhost:${port}`

const buildDir = path.resolve(__dirname, '../../_build/playwright')

export default defineConfig({
    testDir: './e2e/',
    outputDir: path.join(buildDir, 'test-results'),

    fullyParallel: false,
    workers: 1,
    retries: 3,

    reporter: [
        ['list'],
        [
            'monocart-reporter',
            {
                name: 'Weblens E2E Report',
                outputFile: path.join(buildDir, 'report/index.html'),
                coverage: {
                    reports: ['v8', 'console-details'],
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

    projects: [
        {
            name: 'setup-flow',
            testMatch: 'setup.spec.ts',
        },
        {
            name: 'authenticated',
            testMatch: [
                'login.spec.ts',
                'files.spec.ts',
                'settings.spec.ts',
                'navigation.spec.ts',
                'file-operations.spec.ts',
                'media-timeline.spec.ts',
                'dev-actions.spec.ts',
                'multi-user.spec.ts',
                'presentation.spec.ts',
                'context-menu.spec.ts',
                'sort-and-view.spec.ts',
                'share-browsing.spec.ts',
                'search-filters.spec.ts',
                'share-interactions.spec.ts',
                'password-change.spec.ts',
                'file-history.spec.ts',
                'empty-states.spec.ts',
                'download.spec.ts',
                'keyboard-shortcuts.spec.ts',
                'file-preview.spec.ts',
                'upload-flow.spec.ts',
            ],
            dependencies: ['setup-flow'],
        },
    ],

    globalSetup: './e2e/global-setup.ts',
    globalTeardown: './e2e/global-teardown.ts',
})
