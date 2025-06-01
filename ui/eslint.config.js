import pluginJs from '@eslint/js'
import pluginReact from 'eslint-plugin-react'
import * as reactHooks from 'eslint-plugin-react-hooks'
import globals from 'globals'
import tseslint from 'typescript-eslint'

/** @type {import('eslint').Linter.Config[]} */
export default [
    { files: ['**/*.{js,mjs,cjs,ts,jsx,tsx}'] },
    { languageOptions: { globals: globals.browser } },
    pluginJs.configs.recommended,
    ...tseslint.configs.recommended,
    pluginReact.configs.flat.recommended,
    reactHooks.configs.recommended,
    {
        rules: {
            'react/jsx-uses-react': 'off',
            'react/react-in-jsx-scope': 'off',
            'react-hooks/react-compiler': 'error',
        },
    },
    {
        ignores: ['**/node_modules/**', '**/dist/**', '**/swag/**'],
    },
]
