import { createRouter, createWebHistory } from 'vue-router'

const routes = [
    {
        path: '/',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue')
    },
    {
        path: '/create',
        name: 'CreateLink',
        component: () => import('@/views/CreateLink.vue')
    },
    {
        path: '/manage',
        name: 'ManageLinks',
        component: () => import('@/views/ManageLinks.vue')
    },
    {
        path: '/analytics',
        name: 'Analytics',
        component: () => import('@/views/Analytics.vue')
    },
    {
        path: '/batch',
        name: 'BatchCreate',
        component: () => import('@/views/BatchCreate.vue')
    },
    {
        path: '/:pathMatch(.*)*',
        redirect: '/'
    }
]

const router = createRouter({
    history: createWebHistory(),
    routes
})

export default router