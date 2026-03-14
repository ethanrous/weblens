---
name: add-frontend-feature
description: Add a new frontend feature to the Nuxt 4 SPA including components, stores, pages, and API integration
---

# Add Frontend Feature

The frontend is a Nuxt 4 SPA at `weblens-vue/weblens-nuxt/`. SSR is disabled.

## Directory layout

```
weblens-vue/weblens-nuxt/
  components/
    atom/          # Small reusable components (buttons, inputs, icons)
    molecule/      # Composed components (cards, list items, info panels)
    organism/      # Full sections (modals, sidebars, context menus)
  pages/           # File-based routing (Nuxt auto-routes)
  stores/          # Pinia stores (state management)
  composables/     # Vue composables (reusable logic)
  api/             # Custom API helpers beyond generated client
  types/           # TypeScript type definitions
  util/            # Utility functions
  e2e/             # Playwright test specs
  assets/css/      # Global CSS / Tailwind
```

## Component naming

- Atomic design: atom < molecule < organism
- Vue SFC files use PascalCase: `MyComponent.vue`
- Component tests in `components/__tests__/`

## State management (Pinia)

Stores are in `stores/`. Each store is a single `.ts` file. Existing stores: `files`, `media`, `tags`, `user`, `upload`, `tasks`, `tower`, `websocket`, `presentation`, `confirm`, `contextMenu`, `location`.

## API client

The TypeScript API client is auto-generated at `api/ts/generated/` from Swagger. After backend changes:

```bash
make swag
```

This regenerates the TS client. Import from `@ethanrous/weblens-api` in frontend code.

## Adding a page

1. Create `pages/<name>.vue` — Nuxt auto-generates the route
2. For nested routes, use directory structure: `pages/settings/index.vue`

## Adding a component

1. Decide atom/molecule/organism based on complexity
2. Create the `.vue` file in the correct directory
3. Use Tailwind for styling (project convention)

## Playwright E2E tests

Tests live in `weblens-vue/weblens-nuxt/e2e/`.

Run all:
```bash
make test-ui
```

Run specific:
```bash
./scripts/test-playwright.bash --grep 'test name'
```

## Dev server

```bash
make dev
```

Frontend runs on `:3000` and proxies API calls to backend on `:8080`.
