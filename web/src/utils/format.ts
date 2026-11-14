import dayjs from 'dayjs'

export function formatDateTime(value: string): string {
    return dayjs(value).format('YYYY-MM-DD HH:mm:ss')
}

export function formatDate(value: string): string {
    return dayjs(value).format('MMM D, YYYY')
}
