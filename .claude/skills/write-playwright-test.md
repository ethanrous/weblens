---
name: write-playwright-test
description: Write Playwright E2E tests for the weblens Nuxt frontend — emphasizes extending existing spec files over creating new ones
---

# Write Playwright Test

## CRITICAL: Extend existing specs, don't create new files

Before writing any test, **search for an existing spec file** that covers the feature area. Almost every user flow already has a spec — add new `test()` blocks or `test.describe()` blocks to those files. Test slop (scattered thin spec files) makes the suite brittle and slow (each file can spawn its own backend worker).

Existing spec files in `weblens-vue/weblens-nuxt/e2e/`:

| File | Covers |
|------|--------|
| `files.spec.ts` | File browser core: navigation, breadcrumbs, file display |
| `file-operations.spec.ts` | Upload, download, context menus, share modal, trash, search, folder history |
| `tags.spec.ts` | Tag manager, tag assignment via context menu, tag display on cards |
| `upload-flow.spec.ts` | Upload with progress tracking, media uploads |
| `search-filters.spec.ts` | Local/recursive search, regex mode |
| `sort-and-view.spec.ts` | Sort controls, grid/row view switching |
| `download.spec.ts` | Single/multi/directory downloads |
| `context-menu.spec.ts` | File and folder context menu actions |
| `keyboard-shortcuts.spec.ts` | Ctrl+K, Shift+Ctrl+K, Ctrl+A, Escape, Space |
| `login.spec.ts` | Authentication, login page behavior |
| `navigation.spec.ts` | Route protection, redirects |
| `settings.spec.ts` | Account, appearance, users, developer settings |
| `multi-user.spec.ts` | Non-admin user scenarios |
| `password-change.spec.ts` | Password change form validation |
| `file-history.spec.ts` | File history, folder rewind |
| `dev-actions.spec.ts` | Developer page actions, user management |
| `presentation-info.spec.ts` | File presentation mode, info panel |
| `presentation.spec.ts` | Presentation/lightbox navigation and controls |
| `file-preview.spec.ts` | File content preview (text, image, video) |
| `media-timeline.spec.ts` | Media timeline view, thumbnail loading |
| `empty-states.spec.ts` | Empty folder, no media, no shares states |
| `setup.spec.ts` | Initial server setup flow |
| `share-browsing.spec.ts` | Browsing shared files as recipient |
| `share-interactions.spec.ts` | Share creation, permission editing, link copying |

**Add new tests to the matching file.** For example, if you added a new context menu action, add a `test()` inside the relevant `test.describe()` in `context-menu.spec.ts` or `file-operations.spec.ts` — don't create a new file.

Only create a new spec file if the feature is genuinely new and doesn't fit any existing file.

---

## Test infrastructure

### Fixtures (`e2e/fixtures.ts`)

All tests import from `fixtures.ts`, not from `@playwright/test` directly:

```typescript
import { test, expect, createFolder, uploadTestFile, login, createUser } from './fixtures'
```

Available fixtures:

- **`testBackend`** — auto-starts an isolated weblens binary per worker (unique port, DB, filesystem). Auto-cleaned after test.
- **`baseURL`** — derived from `testBackend`, points to the test server
- **`login`** — logs in as admin (`admin`/`adminadmin1`) and waits for Home to load. Use as fixture parameter: `{ page, login: _login }`
- **`autoTestFixture`** — auto-applied to all tests. Captures browser console logs and JS coverage.

### Helper functions from `fixtures.ts`

```typescript
// Login as specific user (default: admin/adminadmin1)
await login(page, username?, password?)

// Create a folder in current directory, waits for it to appear
await createFolder(page, 'My Folder')

// Upload an inline text file, waits for card to appear
await uploadTestFile(page, 'readme.txt', 'file content here')

// Create a new user via settings page
await createUser(page, 'newuser', 'password123')
```

---

## Test structure

### Basic pattern

```typescript
test.describe('Feature Name', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        // Common setup: create test data
        await createFolder(page, 'Test Folder')
        // Reload to get full permissions if needed
        await page.reload()
        await expect(page.locator('[id^="file-card-"]').first()).toBeVisible({ timeout: 15000 })
    })

    test('should do the thing', async ({ page }) => {
        // Arrange, Act, Assert
    })
})
```

### Adding to an existing describe block

If the existing file has a `test.describe('File Operations', ...)`, add your test inside that block. If your test needs different setup, add a new `test.describe()` block in the same file:

```typescript
// In file-operations.spec.ts — ADD this block, don't create a new file
test.describe('New Feature in File Operations', () => {
    test.beforeEach(async ({ page, login: _login }) => {
        // setup specific to these tests
    })

    test('should handle the new feature', async ({ page }) => {
        // ...
    })
})
```

---

## Locator patterns

Use the same locator patterns the existing tests use — consistency prevents flaky selectors.

### File cards
```typescript
const fileCard = page.locator('[id^="file-card-"]').filter({ hasText: 'filename' })
```

### Context menu (scope to filebrowser to avoid matching sidebar buttons)
```typescript
const fileBrowser = page.locator('#filebrowser-container')
await fileBrowser.getByRole('button', { name: 'Download' }).click()

// Or the explicit context menu element
const contextMenu = page.locator('#file-context-menu')
await contextMenu.getByRole('button', { name: 'Tags' }).click()
```

### Modal windows
```typescript
const modal = page.locator('.fullscreen-modal')
await modal.getByRole('button', { name: 'Done' }).click()
```

### Sidebar (scope to avoid ambiguity with filebrowser)
```typescript
const sidebar = page.locator('#global-left-sidebar')
await sidebar.getByRole('button', { name: 'Home' }).click()
```

### Sort/view controls (use icons, not labels — labels may be hidden in narrow viewports)
```typescript
const sortIcon = page.locator('.tabler-icon-calendar, .tabler-icon-sort-a-z')
await sortIcon.first().click()
```

### Prefer semantic selectors
```typescript
// Good
await page.getByRole('button', { name: 'Sign in' }).click()
await page.getByPlaceholder('Username').fill('admin')
await page.getByText('Home', { exact: true })

// Avoid raw CSS selectors when a semantic one exists
```

---

## Waiting patterns

The backend processes operations asynchronously via WebSocket. Always use explicit waits, never `setTimeout`.

### Wait for files to load
```typescript
await Promise.race([
    page.locator('#file-scroller').first().waitFor({ state: 'attached', timeout: 15000 }),
    page.getByText('This folder is empty').waitFor({ state: 'visible', timeout: 15000 }),
])
```

### Wait for API response
```typescript
const searchRequest = page.waitForResponse(
    (res) => res.url().includes('/search') && res.status() === 200,
    { timeout: 10000 }
)
await searchInput.press('Enter')
await searchRequest
```

### Wait for download
```typescript
const downloadPromise = page.waitForEvent('download', { timeout: 15000 })
await downloadBtn.click()
const download = await downloadPromise
expect(download.suggestedFilename()).toBe('expected.txt')
```

### Wait for WebSocket-driven updates (media thumbnails, task progress)
```typescript
await expect(fileCard.locator('.media-image-lowres')).toBeVisible({ timeout: 10_000 })
```

### Generous timeouts
The test backend does real work (MongoDB, file I/O, image processing). Use `{ timeout: 15000 }` for operations that depend on backend processing. Use `test.slow()` for tests involving media processing.

---

## Authentication

### Default admin login (most tests)
```typescript
test('my test', async ({ page, login: _login }) => {
    // Already logged in at /files/home
})
```

### Custom user login
```typescript
test('should work as regular user', async ({ page }) => {
    // First create the user (requires admin)
    await login(page, 'admin', 'adminadmin1')
    await createUser(page, 'regularuser', 'password123')

    // Log out and log in as the new user
    // Navigate to login, fill credentials
    await login(page, 'regularuser', 'password123')
})
```

---

## Test data setup

### Create folders and files in `beforeEach`
```typescript
test.beforeEach(async ({ page, login: _login }) => {
    await uploadTestFile(page, 'test.txt', 'content')
    await createFolder(page, 'My Folder')
    // Reload to get full permissions
    await page.reload()
    await expect(page.locator('[id^="file-card-"]').first()).toBeVisible({ timeout: 15000 })
})
```

### Upload real media files
```typescript
import path from 'path'

const testMediaDir = path.resolve(__dirname, '../../../images/testMedia')
const fileChooserPromise = page.waitForEvent('filechooser')
await page.getByRole('button', { name: 'Upload' }).click()
const fileChooser = await fileChooserPromise
await fileChooser.setFiles([path.join(testMediaDir, 'DSC08113.jpg')])
```

---

## Common assertions

```typescript
// Visibility (with timeout for async operations)
await expect(element).toBeVisible({ timeout: 15000 })
await expect(element).not.toBeVisible({ timeout: 15000 })

// URL
await page.waitForURL('**/files/home')
await expect(page).toHaveURL(/\/files\/home$/)

// CSS class (selection state)
await expect(fileCard).toHaveClass(/bg-card-background-selected/)

// Form state
await expect(button).toBeEnabled()
await expect(button).toBeDisabled()

// Keyboard
await page.keyboard.press('Escape')
await page.keyboard.press('ControlOrMeta+k')
```

---

## Error handling in tests

### Graceful catch for optional elements
```typescript
const download = await downloadPromise.catch(() => null)
```

### Boolean visibility check with fallback
```typescript
const isVisible = await element.isVisible({ timeout: 3000 }).catch(() => false)
```

### Mark slow tests
```typescript
test('should process media images', async ({ page }) => {
    test.slow()  // Triples the timeout
    // ... media processing test
})
```

---

## Running

```bash
# All Playwright tests
make test-ui

# Specific test by name
./scripts/test-playwright.bash --grep 'should create a new tag'
```

---

## Backend isolation

Each Playwright worker gets its own:
- `weblens_debug` binary (spawned from `backend-manager.ts`)
- Unique MongoDB database (`pw-<test-name>`)
- Fresh filesystem at `_build/playwright/fs/worker-<n>`
- Unique port (`10100 + workerIndex * 1000 + random`)
- Logs at `_build/logs/playwright/`

On test failure, the fixture logs the paths to backend and browser console logs.
