import type {ChipColor} from '@/components/base/badges/chip'

export interface Language {
    label: string
    value: string
    runtime: string
    color: ChipColor
}

export const LANGUAGES: Language[] = [
    {label: 'PHP', value: 'php', runtime: 'PHP 8.2', color: 'purple'},
    {label: 'Python', value: 'python', runtime: 'Python 3.12', color: 'yellow'},
    {label: 'Go', value: 'go', runtime: 'Go 1.27.3', color: 'cyan'},
    {label: 'JavaScript', value: 'javascript', runtime: 'QuickJS (not Node.js)', color: 'orange'},
    {label: 'C', value: 'c', runtime: 'Clang 22', color: 'blue'},
]

export const LANGUAGES_MAP: Record<string, Language> = Object.fromEntries(
    LANGUAGES.map((language) => [language.value, language]),
)
