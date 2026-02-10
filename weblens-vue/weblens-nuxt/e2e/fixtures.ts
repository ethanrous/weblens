import { test as base, expect } from '@playwright/test'
import { addCoverageReport } from 'monocart-reporter'

const test = base.extend<{ autoTestFixture: unknown }>({
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

export { test, expect }
