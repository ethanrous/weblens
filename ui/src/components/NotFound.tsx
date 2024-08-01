import { Space } from '@mantine/core'
import { useNavigate } from 'react-router-dom'
import WeblensButton from './WeblensButton'
import { useSessionStore } from './UserInfo'

function NotFound({
    resourceType,
    link,
    setNotFound,
}: {
    resourceType: string
    link: string
    setNotFound: (b: boolean) => void
}) {
    const user = useSessionStore((state) => state.user)
    const nav = useNavigate()
    return (
        <div className="flex flex-col justify-center items-center h-full w-full">
            <div className="flex flex-col h-max w-[360px] p-12 mb-[40vh] bg-bottom-grey outline outline-main-accent rounded justify-center items-center">
                <p className="font-bold text-2xl w-max">{`Could not find ${resourceType}`}</p>

                <Space h={15} />

                <WeblensButton
                    centerContent
                    fillWidth
                    label={user.username ? 'Go Back' : 'Login'}
                    onClick={() => {
                        setNotFound(false)
                        nav(user.username ? link : '/login')
                    }}
                />
            </div>
        </div>
    )
}

export default NotFound
