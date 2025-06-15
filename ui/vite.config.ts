import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import fs from 'node:fs'
import { ServerOptions } from 'node:https'
import path from 'node:path'
import { defineConfig, loadEnv } from 'vite'
import viteTsconfigPaths from 'vite-tsconfig-paths'

export default ({ mode }: { mode: string }) => {
    process.env = { ...process.env, ...loadEnv(mode, process.cwd()) }

    if (!process.env.VITE_PROXY_PORT && process.env.VITE_BUILD !== 'true') {
        process.env.VITE_PROXY_PORT = '8080'
        console.warn(
            `VITE_PROXY_PORT not set\nDefaulting proxy to ${process.env.VITE_PROXY_PORT}`
        )
    }
    if (!process.env.VITE_PROXY_HOST && process.env.VITE_BUILD !== 'true') {
        process.env.VITE_PROXY_HOST = '127.0.0.1'
        console.warn(
            `VITE_PROXY_HOST not set\nDefaulting proxy to ${process.env.VITE_PROXY_HOST}`
        )
    }

    // this sets a default port to 3000
    const vitePort = Number(process.env.VITE_PORT)
        ? Number(process.env.VITE_PORT)
        : 3000

    console.log(`Vite is running in ${mode} mode on port ${vitePort}`)

    const useHTTPS = process.env.VITE_USE_HTTPS === 'true'
    let httpsConfig: ServerOptions | undefined
    if (useHTTPS) {
        httpsConfig = {
            key: fs.readFileSync(
                path.resolve(__dirname, '../cert/local.weblens.io.key')
            ),
            cert: fs.readFileSync(
                path.resolve(__dirname, '../cert/local.weblens.io.crt')
            ),
        }
        console.log('Using HTTPS for development server')
    } else {
        console.log('Using HTTP for development server')
    }

    return defineConfig({
        // depending on your application, base can also be "/"
        base: '/',
        plugins: [react(), viteTsconfigPaths(), tailwindcss()],
        mode: 'development',
        server: {
            // this ensures that the browser opens upon server start
            open: 'local.weblens.io',
            host: '0.0.0.0',
            allowedHosts: ['local.weblens.io'],
            port: vitePort,
            proxy: {
                '/api/v1': {
                    target: `http://${process.env.VITE_PROXY_HOST}:${process.env.VITE_PROXY_PORT}`,
                },
                '/api/v1/ws': {
                    target: `ws://${process.env.VITE_PROXY_HOST}:${process.env.VITE_PROXY_PORT}`,
                    ws: true,
                },
            },
            https: httpsConfig,
        },
        resolve: {
            alias: {
                '@weblens': path.resolve(__dirname, './src'),
                '~': path.resolve(__dirname, './src/assets'),
                '@tabler/icons-react':
                    '@tabler/icons-react/dist/esm/icons/index.mjs',
            },
        },
        css: {
            preprocessorOptions: {
                scss: {
                    api: 'modern',
                },
            },
            modules: {
                localsConvention: 'camelCaseOnly',
            },
        },
    })
}
