import dayjs from 'dayjs'
import {getLocale} from '@/i18n'

const zh = () => getLocale() === 'zh-CN'

export function formatDateTime(value: string): string {
    return dayjs(value).format('YYYY-MM-DD HH:mm:ss')
}

export function formatDate(value: string): string {
    return dayjs(value).format(zh() ? 'YYYY年M月D日' : 'MMM D, YYYY')
}

export function formatMonthDay(value: string): string {
    return dayjs(value).format(zh() ? 'M月D日' : 'MMM D')
}

export function formatWeekdayMonthDay(value: string): string {
    return dayjs(value).format(zh() ? 'M月D日 ddd' : 'ddd, MMM D')
}

export function formatMonthDayTime(value: string): string {
    return dayjs(value).format(zh() ? 'M月D日 HH:mm' : 'MMM D, HH:mm')
}
