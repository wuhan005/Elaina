import {useSyncExternalStore} from 'react'
import type {User} from '@/api/api'

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
let user: User | null = null

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
            // Keep the session in memory only when storage is unavailable.
        }
        emit()
    },

    setUser(next: User | null) {
        user = next
        emit()
    },

    clear() {
        token = ''
        user = null
        try {
            window.localStorage.removeItem(STORAGE_KEY)
        } catch {
            // Same as above.
        }
        emit()
    },
}

// useIsAuthenticated re-renders when the token is cleared, e.g. after a 401.
export function useIsAuthenticated(): boolean {
    return useSyncExternalStore(subscribe, () => token !== '', () => false)
}

export function useCurrentUser(): User | null {
    return useSyncExternalStore(subscribe, () => user, () => null)
}
