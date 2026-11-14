import {useSyncExternalStore} from 'react'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import {zhCN} from './zh-CN'

export type Locale = 'en' | 'zh-CN'

export const LOCALES: {value: Locale; label: string}[] = [
    {value: 'en', label: 'English'},
    {value: 'zh-CN', label: '简体中文'},
]

const STORAGE_KEY = 'elaina:locale'
const listeners = new Set<() => void>()

function detectLocale(): Locale {
    try {
        const stored = window.localStorage.getItem(STORAGE_KEY)
        if (stored === 'en' || stored === 'zh-CN') return stored
    } catch {
        // Fall back to the browser language when storage is unavailable.
    }
    const preferred = navigator.languages?.find((language) => /^(en|zh)\b/i.test(language)) ?? navigator.language
    return /^zh\b/i.test(preferred ?? '') ? 'zh-CN' : 'en'
}

let locale: Locale = detectLocale()

function apply() {
    document.documentElement.lang = locale
    dayjs.locale(locale === 'zh-CN' ? 'zh-cn' : 'en')
}

apply()

export function getLocale(): Locale {
    return locale
}

export function setLocale(next: Locale) {
    if (next === locale) return
    locale = next
    try {
        window.localStorage.setItem(STORAGE_KEY, next)
    } catch {
        // The choice still applies to this tab.
    }
    apply()
    for (const listener of listeners) listener()
}

function subscribe(listener: () => void) {
    listeners.add(listener)
    return () => {
        listeners.delete(listener)
    }
}

export type TranslateVars = Record<string, string | number>

// translate looks up the English source text in the current locale and fills {name} placeholders.
export function translate(text: string, vars?: TranslateVars): string {
    const template = locale === 'zh-CN' ? (zhCN[text] ?? text) : text
    if (!vars) return template
    return template.replace(/\{(\w+)\}/g, (match: string, key: string) => (key in vars ? String(vars[key]) : match))
}

export function useLocale(): Locale {
    return useSyncExternalStore(subscribe, getLocale, getLocale)
}

// useT re-renders the component when the locale changes.
export function useT(): typeof translate {
    useLocale()
    return translate
}
