import { IconFolder, IconHome, IconSlash, IconTrash } from '@tabler/icons-react'
import { useSessionStore } from '@weblens/components/UserInfo'
import { filenameFromPath } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { FC } from 'react'

export function PathFmt({
    pathName,
    excludeBasenameMatching,
    className,
}: {
    pathName: string
    excludeBasenameMatching?: string
    className?: string
}) {
    pathName = pathName.slice(pathName.indexOf(':') + 1)
    const parts = pathName.split('/')

    if (parts[parts.length - 1] === '') {
        parts.pop()
    }

    let StartIcon: FC<{ className: string; size?: number }>
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

    if (parts[parts.length - 1] === excludeBasenameMatching) {
        parts.pop()
    }

    return (
        <div
            className={'flex min-w-0 items-center ' + (className ?? '')}
            style={{ flexShrink: parts.length ? 1 : 0 }}
        >
            <StartIcon className="shrink-0" size={16} />
            {parts.map((part) => {
                return (
                    <div
                        key={part}
                        className="flex w-max min-w-0 shrink items-center"
                    >
                        <IconSlash className="shrink-0" size={16} />
                        <span className="font-[inherit] [font-weight:inherit] text-nowrap truncate">
                            {part}
                        </span>
                    </div>
                )
            })}
        </div>
    )
}

export function FileFmt({
    pathName,
    className,
}: {
    pathName: string
    className?: string
}) {
    let nameText = '---'
    let StartIcon: FC<{ className: string; size: number }> = IconFolder
    if (pathName) {
        const fname = filenameFromPath(pathName)
        nameText = fname.nameText
        StartIcon = fname.StartIcon
    }

    return (
        <div
            className={
                'flex w-max max-w-full min-w-0 items-center gap-1 ' +
                (className ?? '')
            }
        >
            {StartIcon && (
                <StartIcon className="theme-text shrink-0" size={16} />
            )}
            <span className="truncate font-[inherit] [font-weight:inherit] text-nowrap select-none">
                {nameText}
            </span>
        </div>
    )
}
