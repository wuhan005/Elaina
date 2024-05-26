import {createRouter, createWebHistory, RouteRecordRaw} from "vue-router";
import {useAuthStore} from "@/store";

const allRouters: Array<RouteRecordRaw> = [
    {
        path: '/sign-in',
        name: 'signIn',
        component: () => import('@/views/SignIn.vue')
    },
    {
        path: '',
        redirect: '/dashboard'
    },
    {
        path: '',
        name: 'layout',
        component: () => import('@/views/Layout.vue'),
        children: [
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
                path: '/template/new',
                name: 'createTemplate',
                component: () => import('@/views/TemplateModify.vue')
            },
            {
                path: '/template/:id',
                name: 'editTemplate',
                component: () => import('@/views/TemplateModify.vue')
            },
            {
                path: '/sandbox',
                name: 'sandbox',
                component: () => import('@/views/Sandbox.vue')
            },
            {
                path: '/sandbox/new',
                name: 'createSandbox',
                component: () => import('@/views/SandboxModify.vue')
            },
            {
                path: '/sandbox/:id',
                name: 'editSandbox',
                component: () => import('@/views/SandboxModify.vue')
            }
        ]
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes: allRouters,
    scrollBehavior() {
        return {el: '#app', top: 0, behavior: 'smooth'}
    }
})

router.beforeEach((to, from, next) => {
    const authStore = useAuthStore()

    if (to.name !== 'signIn' && !authStore.isAuthenticated) {
        next({name: 'signIn'})
    } else {
        next()
    }
})

export default router;
