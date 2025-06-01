import { IconExclamationCircle } from '@tabler/icons-react'
import WeblensButton from '@weblens/lib/WeblensButton.tsx'
import { useNavigate } from 'react-router-dom'

import { useSessionStore } from './UserInfo'

function FilesErrorDisplay({
    error,
    notFound,
    resourceType,
    link,
    setNotFound,
}: {
    error: number
    notFound: boolean
    resourceType: string
    link: string
    setNotFound: (b: number) => void
}) {
    const { user } = useSessionStore()
    const nav = useNavigate()

    let preText = ''
    if (notFound) {
        preText = `Could not find ${resourceType}`
    } else {
        console.error(error)
        preText = `Failed to fetch files`
    }

    return (
        <div className="flex h-full w-full flex-col items-center justify-center">
            <div className="bg-wl-barely-visible mb-[40vh] flex h-max w-[360px] flex-col items-center justify-center rounded-sm border p-12">
                <div className="mb-4 flex items-center gap-1">
                    <p className="w-max text-2xl font-bold">{preText}</p>
                    <IconExclamationCircle size={32} className="text-red-500" />
                </div>

                <WeblensButton
                    centerContent
                    fillWidth
                    label={user.isLoggedIn ? 'Go Back' : 'Login'}
                    onClick={() => {
                        setNotFound(0)
                        if (user.isLoggedIn) {
                            nav(link)
                        } else {
                            nav('/login', {
                                state: { returnTo: window.location.pathname },
                            })
                        }
                    }}
                />
            </div>
        </div>
    )
}

export default FilesErrorDisplay
