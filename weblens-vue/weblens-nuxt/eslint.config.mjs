// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt([
    {
        rules: {
            'vue/html-self-closing': 'off',
            'no-console': ['error', { allow: ['warn', 'error', 'debug'] }],
            'vue/valid-v-model': 'error',
            eqeqeq: 'error',
            '@typescript-eslint/member-ordering': [
                'error',
                {
                    default: [
                        // Private before public, grouped by type (variables, then constructors, then methods)

                        'private-field',
                        'protected-field',
                        'public-field',

                        'constructor',

                        'private-method',
                        'protected-method',
                        'public-method',
                    ],
                },
            ],
        },
    },
])
