import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import viteTsconfigPaths from "vite-tsconfig-paths";

export default ({mode}) => {
    process.env = {...process.env, ...loadEnv(mode, process.cwd())};
    console.log(process.env.VITE_PORT)

    return defineConfig({
        // depending on your application, base can also be "/"
        base: "/",
        plugins: [react(), viteTsconfigPaths()],
        server: {
            // this ensures that the browser opens upon server start
            open: true,
            host: "0.0.0.0",
            // this sets a default port to 3000
            port: Number(process.env.VITE_PORT) ? Number(process.env.VITE_PORT) : 3000,
            proxy: {
                "/api": {
                    target: `http://localhost:${process.env.VITE_PROXY_PORT}`,
                },
                "/api/ws": {
                    target: `ws://localhost:${process.env.VITE_PROXY_PORT}`,
                    ws: true,
                },
            },
        },
    });
}