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
    {
        path: '/login',
        name: 'login',
        component: () => import('@/views/Login.vue')
    },
]

let router = new VueRouter({
    mode: 'hash',
    base: process.env.BASE_URL,
    routes
})

router.beforeEach((to, from, next) => {
    // Router guard
    if (!localStorage.getItem('login') && to.name !== 'login') {
        next({
            name: 'login'
        })
        return
    }
    // Login again
    if (localStorage.getItem('login') && to.name === 'login') {
        next({
            name: 'dashboard'
        })
        return
    }
    next()
})

export default router
