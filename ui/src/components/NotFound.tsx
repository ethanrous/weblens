import { Space } from '@mantine/core'
import { useNavigate } from 'react-router-dom'
import WeblensButton from './WeblensButton'
import { useContext } from 'react'
import { UserContext } from '../Context'
import { UserContextT } from '../types/Types'

function NotFound({
    resourceType,
    link,
    setNotFound,
}: {
    resourceType: string
    link: string
    setNotFound: (b: boolean) => void
}) {
    const { usr }: UserContextT = useContext(UserContext)
    const nav = useNavigate()
    return (
        <div className="flex flex-col justify-center items-center h-full w-full">
            <div className="flex flex-col h-max w-[360px] p-12 mb-[40vh] bg-bottom-grey outline outline-main-accent rounded justify-center items-center">
                <p className="font-bold text-2xl w-max">{`Could not find ${resourceType}`}</p>

                <Space h={15} />

                <WeblensButton
                    centerContent
                    fillWidth
                    label={usr.username ? 'Go Back' : 'Login'}
                    onClick={(e) => {
                        setNotFound(false)
                        nav(usr.username ? link : '/login')
                    }}
                />
            </div>
        </div>
    )
}

export default NotFound
