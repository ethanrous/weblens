import { StateCreator, create } from 'zustand'

export enum ThemeStateEnum {
    LIGHT = 'light',
    DARK = 'dark',
    OS = 'os',
}
type AbsoluteTheme = ThemeStateEnum.DARK | ThemeStateEnum.LIGHT

function globalApplyTheme(
    newTheme: AbsoluteTheme,
    setTheme?: (theme: AbsoluteTheme) => void
) {
    const currentMode =
        document.documentElement.style.getPropertyValue('color-scheme') ===
        'dark'
            ? ThemeStateEnum.DARK
            : ThemeStateEnum.LIGHT

    if (newTheme !== currentMode) {
        document.documentElement.style.setProperty('color-scheme', newTheme)
    }
    if (setTheme) {
        setTheme(newTheme)
    }
}

function getOsTheme(): AbsoluteTheme {
    return window.matchMedia('(prefers-color-scheme: dark)').matches
        ? ThemeStateEnum.DARK
        : ThemeStateEnum.LIGHT
}

function readInitTheme(): { theme: AbsoluteTheme; isOSControlled: boolean } {
    const localTheme = localStorage.getItem('theme') as ThemeStateEnum | null
    if (localTheme && localTheme !== ThemeStateEnum.OS) {
        return { theme: localTheme, isOSControlled: false }
    }
    if (!localTheme) {
        const osTheme = getOsTheme()
        globalApplyTheme(osTheme)
        return { theme: osTheme, isOSControlled: true }
    }

    const osTheme = getOsTheme()
    globalApplyTheme(osTheme)
    return { theme: osTheme, isOSControlled: true }
}

export interface ThemeStateT {
    theme: AbsoluteTheme
    isOSControlled: boolean
    changeTheme: (newTheme: ThemeStateEnum) => void
}

const ThemeControl: StateCreator<ThemeStateT, [], []> = (set) => {
    const callback = (event: MediaQueryListEvent) => {
        set((state) => {
            if (!state.isOSControlled) {
                return state
            }
            const newColorScheme = event.matches
                ? ThemeStateEnum.DARK
                : ThemeStateEnum.LIGHT
            globalApplyTheme(newColorScheme)

            state.theme = newColorScheme
            return { ...state }
        })
    }
    window
        .matchMedia('(prefers-color-scheme: dark)')
        .addEventListener('change', callback)

    const { theme, isOSControlled } = readInitTheme()
    return {
        theme: theme,
        isOSControlled: isOSControlled,

        changeTheme: (newTheme: ThemeStateEnum) => {
            set((state) => {
                let newAbsTheme: AbsoluteTheme
                if (newTheme === ThemeStateEnum.OS) {
                    newAbsTheme = getOsTheme()
                    state.isOSControlled = true
                } else {
                    newAbsTheme = newTheme
                    state.isOSControlled = false
                }
                localStorage.setItem('theme', newTheme)
                globalApplyTheme(newAbsTheme)
                state.theme = newAbsTheme

                return { ...state }
            })
        },
    } as ThemeStateT
}

export const useWlTheme = create<ThemeStateT>()(ThemeControl)
