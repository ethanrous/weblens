// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt([
    {
        rules: {
            'vue/html-self-closing': 'off',
            'no-console': ['error', { allow: ['warn', 'error', 'debug'] }],
        },
    },
])
