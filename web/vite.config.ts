import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import svgLoader from 'vite-svg-loader';
import path from 'node:path';

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [vue(), svgLoader()],

    css: {
        preprocessorOptions: {
            less: {
                modifyVars: {
                    hack: `true; @import (reference) "${path.resolve('src/style/variables.less')}";`,
                },
                math: 'strict',
                javascriptEnabled: true,
            },
        },
    },

    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
        },
    },

    server: {
        port: 3030,
        host: '0.0.0.0',
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            }
        }
    }
})
