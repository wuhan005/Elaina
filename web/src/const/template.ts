export const LANGUAGES = [
    {label: 'PHP', value: 'php'},
    {label: 'Python', value: 'python'},
    {label: 'Go', value: 'go'},
    {label: 'JavaScript', value: 'javascript'},
    {label: 'C', value: 'c'},
]

export const LANGUAGES_MAP: Record<string, string> = Object.fromEntries(
    LANGUAGES.map(({label, value}) => [value, label]),
)
