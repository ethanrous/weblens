export default {
    content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
    theme: {
        extend: {
            colors: {
                'wl-color-amethyst-50': '#ebe6fa',
                'wl-color-amethyst-100': '#d7ccf5',
                'wl-color-amethyst-200': '#af99eb',
                'wl-color-amethyst-300': '#8666e0',
                'wl-color-amethyst-400': '#5e33d6',
                'wl-color-amethyst-500': '#4a1ad1',
                'wl-color-amethyst-600': '#3600cc',
                'wl-color-amethyst-700': '#3100b8',
                'wl-color-amethyst-800': '#2b00a3',
                'wl-color-amethyst-900': '#26008f',

                'wl-color-bluenova-50': '#e5e5ff',
                'wl-color-bluenova-100': '#ccccff',
                'wl-color-bluenova-200': '#9999ff',
                'wl-color-bluenova-300': '#6666ff',
                'wl-color-bluenova-400': '#3333ff',
                'wl-color-bluenova-500': '#1a1aff',
                'wl-color-bluenova-600': '#0000ff',
                'wl-color-bluenova-700': '#0000e5',
                'wl-color-bluenova-800': '#0000cc',
                'wl-color-bluenova-900': '#0000b3',

                'wl-color-graphite-50': '#f5f5f5',
                'wl-color-graphite-100': '#e0e0e0',
                'wl-color-graphite-200': '#cccccc',
                'wl-color-graphite-300': '#b3b3b3',
                'wl-color-graphite-400': '#999999',
                'wl-color-graphite-500': '#808080',
                'wl-color-graphite-600': '#666666',
                'wl-color-graphite-700': '#4d4d4d',
                'wl-color-graphite-800': '#333333',
                'wl-color-graphite-900': '#1a1a1a',

                'wl-text-color-primary': 'var(--wl-text-color-primary)',
                'wl-text-color-secondary': 'var(--wl-text-color-secondary)',
				'wl-border-color-primary': 'var(--wl-border-color-primary)',
            },

            // colors: {
            //     'dark-paper': 'var(--dark-paper)',
            //     'theme-text': 'var(--wl-text-color)',
            //     'theme-text-inverted': 'var(--wl-text-color-inverted)',
            //     'wl-text-dull': 'var(--wl-text-color-dull)',
            //     'wl-barely-visible': 'var(--wl-barely-visible)',
            //     'wl-background': 'var(--wl-background)',
            //     'wl-outline-subtle': 'var(--wl-outline-subtle)',
            //     'main-accent': 'var(--wl-theme-color)',
            //     'bottom-grey': '#121212',
            //     'raised-grey': '#212124',
            //     'light-paper': '#381eaa',
            //     background: '#111418',
            // },
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
