import { IconFolder, IconHome, IconSlash, IconTrash } from '@tabler/icons-react'
import { useSessionStore } from '@weblens/components/UserInfo'
import historyStyle from '@weblens/components/filebrowser/historyStyle.module.scss'
import { filenameFromPath } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { FC } from 'react'

export function PathFmt({ pathName }: { pathName: string }) {
    pathName = pathName.slice(pathName.indexOf(':') + 1)
    const parts = pathName.split('/')

    if (parts[parts.length - 1] === '') {
        parts.pop()
    }

    let StartIcon: FC<{ className: string }>
    while (parts.includes('.user_trash')) {
        parts.shift()
        StartIcon = IconTrash
    }

    if (!StartIcon) {
        if (parts[0] === useSessionStore.getState().user.username) {
            parts.shift()
        }
        StartIcon = IconHome
    }

    return (
        <div
            className="flex items-center min-w-0"
            style={{ flexShrink: parts.length ? 1 : 0 }}
        >
            <StartIcon className="shrink-0 text-[--wl-text-color]" />
            {parts.map((part) => {
                return (
                    <div
                        key={part}
                        className="flex w-max items-center shrink min-w-0"
                    >
                        <IconSlash
                            className="text-[--wl-text-color] shrink-0"
                            size={18}
                        />
                        <p className={historyStyle['path-text']}>{part}</p>
                    </div>
                )
            })}
        </div>
    )
}

export function FileFmt({ pathName }: { pathName: string }) {
    let nameText = '---'
    let StartIcon: FC<{ className: string }> = IconFolder
    if (pathName) {
        const fname = filenameFromPath(pathName)
        nameText = fname.nameText
        StartIcon = fname.StartIcon
    }

    return (
        <div className="flex items-center w-max min-w-0 max-w-full">
            {StartIcon && (
                <StartIcon className="theme-text m-1 shrink-0 text-[--wl-text-color]" />
            )}
            <p className="theme-text select-none font-semibold text-lg text-nowrap truncate">
                {nameText}
            </p>
        </div>
    )
}
