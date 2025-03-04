function withOpacity(variableName) {
    return ({ opacityValue }) => {
        if (opacityValue !== undefined) {
            return `rgba(var(${variableName}), ${opacityValue})`
        }
        return `rgb(var(${variableName}))`
    }
}

export default {
    content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
    theme: {
        extend: {
            colors: {
                'wl-theme-color-primary':
                    'rgb(from var(--wl-theme-color-primary) r g b / <alpha-value>)',
                'wl-message-color':
                    'rgb(from var(--wl-message-color) r g b / <alpha-value>)',

                'wl-text-color-primary': 'var(--wl-text-color-primary)',
                'wl-text-color-secondary': 'var(--wl-text-color-secondary)',

                'wl-border-color-primary': 'var(--wl-border-color-primary)',

                'wl-color-nearly-invisible': 'var(--wl-color-nearly-invisible)',

                'wl-background-color-primary':
                    'rgb(from var(--wl-background-color-primary) r g b / <alpha-value>)',
                'wl-background-color-secondary':
                    'rgb(from var(--wl-background-color-secondary) r g b / <alpha-value>)',

                'wl-progress-bar-color-primary':
				'rgb(from var(--wl-progress-bar-color-primary) r g b / <alpha-value>)',
                'wl-progress-bar-color-secondary':
                    'var(--wl-progress-bar-color-secondary)',

                'wl-button-text-color': 'var(--wl-button-text-color)',
            },

            animation: {
                fade: 'fadeIn 200ms ease-in-out',
                popup: 'popup 150ms ease-in-out',
                'fade-short': 'fadeIn 100ms ease-in-out',
            },
            keyframes: () => ({
                popup: {
                    '0%': { opacity: 0, transform: 'scale(0.90)' },
                    '100%': { opacity: 100, transform: 'scale(1)' },
                },
                fadeInAndOut: {
                    '0%': { opacity: 35 },
                    '50%': { opacity: 100 },
                    '100%': { opacity: 35 },
                },
            }),
        },
    },
    plugins: [
        ({ addUtilities }) => {
            const newUtilities = {
                '.wl-outline': {
                    borderRadius: 'var(--wl-border-radius)',
                    outline: '1px solid var(--wl-outline)',
                },
                '.wl-outline-subtle': {
                    borderRadius: 'var(--wl-border-radius)',
                    outline: '1px solid var(--wl-outline-subtle)',
                },
            }
            addUtilities(newUtilities, ['responsive', 'hover'])
        },
    ],
}
