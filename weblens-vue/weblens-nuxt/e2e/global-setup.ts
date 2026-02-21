async function globalSetup(): Promise<void> {
    // Record the start time of the entire test run in a global environment variable
    process.env.WEBLENS_PLAYWRIGHT_TEST_START_TIME = Date.now().toString()
}

export default globalSetup
