/** @type {import('tailwindcss').Config} */

export default {
    content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
    theme: {
        extend: {
            colors: {
                'dark-paper': 'var(--dark-paper)',
                'theme-text': 'var(--wl-text-color)',
                'theme-text-inverted': 'var(--wl-text-color-inverted)',
                'wl-barely-visible': 'var(--wl-barely-visible)',
                'wl-background': 'var(--wl-background)',
                'wl-outline-subtle': 'var(--wl-outline-subtle)',
                'main-accent': 'var(--wl-theme-color)',
                'bottom-grey': '#121212',
                'raised-grey': '#212124',
                'light-paper': '#381eaa',
                background: '#111418',
            },
            'wl-outline-subtle': {
                outline: '1px solid var(--wl-outline-subtle)',
            },
            boxShadow: {
                soft: '2px 2px 12px #000000aa',
            },
            animation: {
                fade: 'fadeIn 200ms ease-in-out',
                'fade-short': 'fadeIn 100ms ease-in-out',
            },
            keyframes: () => ({
                // fadeIn: {
                //     '0%': { opacity: 0 },
                //     '100%': { opacity: 100 },
                // },
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
