import axios from 'axios'
import {toast} from '@/components/base/toast/toast'
import {getLocale, translate} from '@/i18n'
import {authStore} from '@/store/auth'
import {Api, type ErrorResponse} from './api'

declare module 'axios' {
    interface AxiosRequestConfig {
        // Statuses whose errors the page shows itself, so no global toast is shown.
        quietStatuses?: number[]
    }
}

export const client = new Api({
    baseURL: '/',
})

client.instance.interceptors.request.use((config) => {
    // The server's i18n middleware prefers ?lang= over its cookie and Accept-Language.
    config.params = {...config.params, lang: getLocale() === 'zh-CN' ? 'zh-CN' : 'en-US'}
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
                // Clearing the token makes the route guard redirect to the sign-in page.
                if (authStore.getToken()) {
                    toast.error(translate('Session expired'), {description: translate('Please sign in again.'), id: 'session-expired'})
                }
                authStore.clear()
                return Promise.reject(error)
            }
            if (error.response && error.config?.quietStatuses?.includes(error.response.status)) {
                return Promise.reject(error)
            }

            const fallback = error.response
                ? translate('Request failed with status {status}', {status: error.response.status})
                : translate('Network error. Check your connection and try again.')
            toast.error(error.response?.data?.msg || fallback)
        } else {
            toast.error(translate('Unknown error'))
        }

        return Promise.reject(error)
    },
)
