/** @type {import('tailwindcss').Config} */
export default {
    content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
    theme: {
        extend: {
            colors: {
                'dark-paper': '#1c1049',
                'main-accent': '#3636ca',
                'bottom-grey': '#121212',
                'raised-grey': '#212124',
            },
            animation: {
                fade: 'fadeIn 200ms ease-in-out',
            },
            keyframes: () => ({
                fadeIn: {
                    '0%': { opacity: 0 },
                    '100%': { opacity: 100 },
                },
            }),
        },
    },
    plugins: [],
};
