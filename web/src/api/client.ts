import axios from 'axios'
import {toast} from '@/components/base/toast/toast'
import {authStore} from '@/store/auth'
import {Api, type ErrorResponse} from './api'

export const client = new Api({
    baseURL: '/',
})

client.instance.interceptors.request.use((config) => {
    const token = authStore.getToken()
    if (token) {
        config.headers.Authorization = `Token ${token}`
    }

    return config
})

client.instance.interceptors.response.use(
    (response) => response,
    (error: unknown) => {
        if (axios.isAxiosError<ErrorResponse>(error)) {
            if (error.response?.status === 401) {
                // 清空 token 后路由守卫会自动跳转到登录页。
                if (authStore.getToken()) {
                    toast.error('Session expired', {description: 'Please sign in again.', id: 'session-expired'})
                }
                authStore.clear()
                return Promise.reject(error)
            }

            toast.error(error.response?.data?.msg || error.message || 'Unknown error')
        } else {
            toast.error('Unknown error')
        }

        return Promise.reject(error)
    },
)
