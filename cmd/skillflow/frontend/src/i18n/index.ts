import { zh } from './zh'
import { en } from './en'

export type Lang = 'zh' | 'en'
export type Translations = typeof zh

// Compile-time guard: errors if en and zh key sets diverge
type _EnCheck = { [K in keyof Translations]: (typeof en)[K] }

export const locales: Record<Lang, Translations> = { zh, en }
