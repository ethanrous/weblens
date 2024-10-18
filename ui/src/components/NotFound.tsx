import { Space } from '@mantine/core'
import WeblensButton from '@weblens/lib/WeblensButton'
import { useNavigate } from 'react-router-dom'
import { useSessionStore } from './UserInfo'
import { IconExclamationCircle } from '@tabler/icons-react'

function FilesErrorDisplay({
    error,
    resourceType,
    link,
    setNotFound,
}: {
    error: number
    resourceType: string
    link: string
    setNotFound: (b: number) => void
}) {
    const { user } = useSessionStore()
    const nav = useNavigate()

    let preText = ''
    if (error === 404) {
        preText = `Could not find ${resourceType}`
    } else {
        preText = `Failed to fetch files`
    }

    return (
        <div className="flex flex-col justify-center items-center h-full w-full">
            <div className="flex flex-col h-max w-[360px] p-12 mb-[40vh] bg-wl-barely-visible outline outline-main-accent rounded justify-center items-center">
                <div className="flex items-center gap-1">
                    <p className="font-bold text-2xl w-max">{preText}</p>
                    <IconExclamationCircle size={32} className="text-red-500" />
                </div>

                <Space h={15} />

                <WeblensButton
                    centerContent
                    fillWidth
                    label={user.isLoggedIn ? 'Go Back' : 'Login'}
                    onClick={() => {
                        setNotFound(0)
                        nav(user.isLoggedIn ? link : '/login')
                    }}
                />
            </div>
        </div>
    )
}

export default FilesErrorDisplay
