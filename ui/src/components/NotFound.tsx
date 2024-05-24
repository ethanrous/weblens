import { Box, Space, Text } from '@mantine/core'
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
            <div
                style={{
                    height: 'max-content',
                    width: 'max-content',
                    padding: 50,
                    marginBottom: '40vh',
                    backgroundColor: '#4444ff33',
                    outline: '1px solid #222277',
                    borderRadius: 6,
                }}
            >
                <Text
                    fw={600}
                    size="25px"
                    c="white"
                >{`Could not find ${resourceType}`}</Text>
                <Space h={15} />

                <WeblensButton
                    centerContent
                    height={40}
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
