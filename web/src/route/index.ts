import {createRouter, createWebHistory, RouteRecordRaw} from "vue-router";

const allRouters: Array<RouteRecordRaw> = [
    {
        path: '/',
        redirect: '/dashboard'
    },
    {
        path: '/dashboard',
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
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes: allRouters,
    scrollBehavior() {
        return {el: '#app', top: 0, behavior: 'smooth'}
    }
})

export default router;
