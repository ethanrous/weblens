// https://nuxt.com/docs/api/configuration/nuxt-config

import tailwindcss from '@tailwindcss/vite'
import { loadEnv } from 'vite'

const mode = process.env.VITE_BUILD === 'true' ? 'production' : 'development'
process.env = { ...process.env, ...loadEnv(mode, process.cwd()) }

if (!process.env.VITE_PROXY_PORT && process.env.VITE_BUILD !== 'true') {
    process.env.VITE_PROXY_PORT = '8080'
    console.warn(`VITE_PROXY_PORT not set\nDefaulting proxy to ${process.env.VITE_PROXY_PORT}`)
}

if (!process.env.VITE_PROXY_HOST && process.env.VITE_BUILD !== 'true') {
    process.env.VITE_PROXY_HOST = '127.0.0.1'
    console.warn(`VITE_PROXY_HOST not set\nDefaulting proxy to ${process.env.VITE_PROXY_HOST}`)
}

// this sets a default port to 3000
const vitePort = Number(process.env.VITE_PORT) ? Number(process.env.VITE_PORT) : 3000

console.debug(`Vite is running in ${mode} mode on port ${vitePort}`)

export default defineNuxtConfig({
    compatibilityDate: '2025-05-15',
    ssr: false,
    devtools: {
        enabled: false,

        timeline: {
            enabled: true,
        },
    },
    modules: ['@nuxt/eslint', '@nuxt/image', '@pinia/nuxt'],
    css: ['~/assets/css/base.css', '~/assets/css/main.css'],
    devServer: {
        port: vitePort,
        host: '0.0.0.0',
        cors: {
            origin: '*',
            methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'],
        },
    },
    nitro: {
        devProxy: {
            '/api/v1/ws': {
                target: `ws://${process.env.VITE_PROXY_HOST}:${process.env.VITE_PROXY_PORT}`,
                prependPath: true,
                changeOrigin: true,
                ws: true,
            },
            '/api/v1': {
                target: `http://${process.env.VITE_PROXY_HOST}:${process.env.VITE_PROXY_PORT}/api/v1`,
                changeOrigin: true,
                prependPath: true,
            },
        },
    },
    sourcemap: {
        client: true,
    },
    vite: {
        plugins: [tailwindcss()],
    },
    app: {
        head: {
            title: 'Weblens',
            link: [
                { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
                { rel: 'preconnect', href: 'https://fonts.gstatic.com' },
                {
                    rel: 'stylesheet',
                    href: 'https://fonts.googleapis.com/css2?family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&family=Noto+Sans:ital,wght@0,100..900;1,100..900&family=Source+Code+Pro:ital,wght@0,200..900;1,200..900&display=swap',
                },
            ],
            meta: [
                { charset: 'utf-8' },
                { name: 'viewport', content: 'width=device-width, initial-scale=1' },

                { property: 'og:site_name', content: 'Weblens' },
                { property: 'al:ios:app_name', content: 'Weblens' },
                { property: 'og:title', content: '{{.Title}}' },
                { property: 'og:description', content: '{{.Description}}' },
                { property: 'og:url', content: '{{.URL}}' },
                { property: 'og:image', content: '{{.Image}}' },
                { property: 'og:type', content: '{{.Type}}' },
                { property: 'og:video:url', content: '{{.VideoURL}}' },
                { property: 'og:video:secure_url', content: '{{.SecureURL}}' },
                { property: 'og:video:type', content: '{{.VideoType}}' },
            ],
        },
    },
})
