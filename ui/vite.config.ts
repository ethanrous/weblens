import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'
import sass from 'sass'
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

	return defineConfig({
		// depending on your application, base can also be "/"
		base: '/',
		plugins: [react(), viteTsconfigPaths(), tailwindcss()],
		mode: 'development',
		server: {
			// this ensures that the browser opens upon server start
			open: true,
			host: '0.0.0.0',
			// this sets a default port to 3000
			port: Number(process.env.VITE_PORT)
				? Number(process.env.VITE_PORT)
				: 3000,
			proxy: {
				'/api/v1': {
					target: `http://${process.env.VITE_PROXY_HOST}:${process.env.VITE_PROXY_PORT}`,
				},
				'/api/v1/ws': {
					target: `ws://${process.env.VITE_PROXY_HOST}:${process.env.VITE_PROXY_PORT}`,
					ws: true,
				},
			},
		},
		resolve: {
			alias: { '@weblens': path.resolve(__dirname, './src') },
			// {
			//     find: '@weblens',
			//     replacement: fileURLToPath(
			//         new URL('./src', import.meta.url)
			//     ),
			// },
		},
		css: {
			preprocessorOptions: {
				scss: {
					implementation: sass,
				},
			},
			modules: {
				localsConvention: 'camelCaseOnly',
			},
		},
	})
}
