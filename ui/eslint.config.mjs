// import reactPlugin from 'eslint-plugin-react';
// import globals from 'globals';
//
// import babelParser from "@babel/eslint-parser"
// import ts from '@typescript-eslint/eslint-plugin';
// import tsParser from '@typescript-eslint/parser';

import eslint from '@eslint/js'
import tseslint from 'typescript-eslint'

export default tseslint.config(
    {
        ignores: [
            'dist/**/*.ts',
            'dist/**',
            '**/*.mjs',
            'eslint.config.mjs',
            '**/*.js',
            'src/api/swag/**/*.ts',
            'src/api/swag/**',
        ],
    },
    eslint.configs.recommended,
    // ...tseslint.configs.recommended,
    ...tseslint.configs.recommendedTypeChecked,
    {
        languageOptions: {
            parserOptions: {
                projectService: true,
                tsconfigRootDir: import.meta.dirname,
            },
        },
    }
)

// const WeblensConfig = [
//     {
//         files: ['**/*.{js,jsx,mjs,cjs,ts,tsx}'],
//         plugins: {
//             '@typescript-eslint': ts,
//             ts,
//             react: reactPlugin,
//         },
//         languageOptions: {
//             parser: tsParser,
//             parserOptions: {
//                 ecmaFeatures: { modules: true },
//                 ecmaVersion: 'latest',
//                 project: './tsconfig.json',
//                 requireConfigFile: false
//             },
//             globals: {
//                 ...globals.browser,
//             },
//         },
//         rules: {
//             ...reactPlugin.configs['jsx-runtime'].rules,
//             ...ts.configs['eslint-recommended'].rules,
//             ...ts.configs['recommended'].rules,
//             'ts/return-await': 2,
//         },
//
//         settings: {
//             react: {
//                 version: 'detect',
//             },
//         },
//     },
// ];
//
// export default WeblensConfig
