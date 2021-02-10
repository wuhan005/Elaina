import Vue from 'vue'
import VueRouter from 'vue-router'

Vue.use(VueRouter)

const routes = [
    {
        path: '/',
        name: 'dashboard',
        component: () => import('@/views/Dashboard.vue')
    },
    {
        path: '/template',
        name: 'template',
        component: () => import('@/views/Template.vue')
    },
    {
        path: '/sandbox',
        name: 'sandbox',
        component: () => import('@/views/Sandbox.vue')
    },
]

const router = new VueRouter({
    mode: 'hash',
    base: process.env.BASE_URL,
    routes
})

export default router
