import {useSyncExternalStore} from 'react'

const STORAGE_KEY = 'elaina:token'

const listeners = new Set<() => void>()

function readToken(): string {
    try {
        return window.localStorage.getItem(STORAGE_KEY) ?? ''
    } catch {
        return ''
    }
}

let token = readToken()

function emit() {
    for (const listener of listeners) listener()
}

function subscribe(listener: () => void) {
    listeners.add(listener)
    return () => {
        listeners.delete(listener)
    }
}

export const authStore = {
    getToken: () => token,

    setToken(next: string) {
        token = next
        try {
            window.localStorage.setItem(STORAGE_KEY, next)
        } catch {
            // 存储不可用时仅保留内存中的会话。
        }
        emit()
    },

    clear() {
        token = ''
        try {
            window.localStorage.removeItem(STORAGE_KEY)
        } catch {
            // 同上。
        }
        emit()
    },
}

// useIsAuthenticated 订阅登录状态，token 被清空（如 401）时触发重新渲染。
export function useIsAuthenticated(): boolean {
    return useSyncExternalStore(subscribe, () => token !== '', () => false)
}
