import { test, expect } from './fixtures'

test.describe('Setup Page', () => {
    test('should redirect to setup page when server is uninitialized', async ({ page }) => {
        await page.goto('/')
        await page.waitForURL('**/setup')
        await expect(page).toHaveURL(/\/setup$/)
    })

    test('should complete server setup and redirect to login', async ({ page }) => {
        await page.goto('/setup')

        // Fill in the server name
        await page.getByPlaceholder('Server Name').fill('Test Server')

        // Fill in the owner display name
        await page.getByPlaceholder('Owner Display Name').fill('Test Admin')

        // Verify the owner username was auto-populated from the display name
        await expect(page.getByPlaceholder('Owner Username')).toHaveValue('test_admin')

        // Fill in the owner password
        await page.getByPlaceholder('Owner Password').fill('password123')

        // Click the setup button
        await page.getByRole('button', { name: 'Set up as a Core Server' }).click()

        // Wait for redirect to login page
        await page.waitForURL('**/login')
        await expect(page).toHaveURL(/\/login$/)
    })
})
