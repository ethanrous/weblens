import { defineConfig } from 'vite'

export default defineConfig({
  build: {
    lib: {
      entry: 'AllApi.ts',          // your library's entry point
      name: 'weblens-api',                  // global name for UMD build (optional)
      formats: ['es', 'umd'],                // only output ESM
      fileName: 'index',              // output filename (no extension)
    },
    rollupOptions: {
      // Externalize dependencies that shouldn't be bundled
      external: ['axios'],                   // e.g. ['vue'] for a Vue lib
      output: {
        // Provide global variables if using UMD, not needed for ESM only
      },
    },
    outDir: 'dist',                   // output folder
    emptyOutDir: true,
    sourcemap: true,
  },
})
