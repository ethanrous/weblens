import { Space, Text } from '@mantine/core'
import { useNavigate } from 'react-router-dom'
import { WeblensButton } from './WeblensButton'
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
            <div className="h-max w-max p-12 mb-[40vh] bg-bottom-grey outline outline-main-accent rounded">
                <Text
                    fw={600}
                    size="25px"
                    c="white"
                >{`Could not find ${resourceType}`}</Text>
                <Space h={15} />

                <WeblensButton
                    centerContent
                    squareSize={40}
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
