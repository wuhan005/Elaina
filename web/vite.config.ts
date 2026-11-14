import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'node:path'

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [react(), tailwindcss()],

    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
        },
    },

    build: {
        rollupOptions: {
            output: {
                manualChunks: {
                    react: ['react', 'react-dom', 'react-router'],
                    'react-aria': ['react-aria-components'],
                    motion: ['motion'],
                },
            },
        },
    },

    server: {
        port: 3030,
        host: '0.0.0.0',
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            },
            '/r/': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            },
            '/static/': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            },
        },
    },
})
