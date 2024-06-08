import { Autocomplete, Box, Space, Text } from '@mantine/core'
import { shareFiles, updateFileShare } from '../../api/FileBrowserApi'
import {
    IconLink,
    IconUser,
    IconUserCancel,
    IconUsersGroup,
    IconX,
} from '@tabler/icons-react'
import { WeblensButton } from '../../components/WeblensButton'

import { useCallback, useContext, useEffect, useState } from 'react'
import { UserContext } from '../../Context'
import { AuthHeaderT, UserContextT } from '../../types/Types'
import { WeblensFile } from '../../classes/File'
import { AutocompleteUsers } from '../../api/ApiFetch'
import { ShareDataT } from '../../classes/Share'

export function ShareInput({
    isPublic,
    sharedUsers,
    setSharedUsers,
}: {
    isPublic: boolean
    sharedUsers: string[]
    setSharedUsers: (v) => void
}) {
    const { usr, authHeader }: UserContextT = useContext(UserContext)
    const [userSearchResult, setUserSearch] = useState([])
    const [search, setSearch] = useState('')

    const searchUsers = useCallback(
        async (query: string) => {
            if (query.length < 2) {
                setUserSearch([])
            }

            let users: string[] = await AutocompleteUsers(query, authHeader)
            const selfIndex = users.indexOf(usr.username)
            if (selfIndex !== -1) {
                users.splice(selfIndex, 1)
            }
            setUserSearch(users)
        },
        [usr.username, authHeader]
    )

    const renderAutocompleteOption = useCallback(
        ({ option }) => {
            return (
                <Box
                    style={{ height: '100%', width: '100%' }}
                    onClick={(e) => {
                        e.stopPropagation()
                        setSharedUsers((v) => {
                            v.push(option.value)
                            return [...v]
                        })
                    }}
                >
                    <Text
                        style={{
                            color: sharedUsers.includes(option.value)
                                ? '#555555'
                                : 'white',
                        }}
                    >
                        {option.value}
                    </Text>
                </Box>
            )
        },
        [sharedUsers, setSharedUsers]
    )

    return (
        <Box style={{ width: 300 }}>
            <Autocomplete
                disabled={isPublic}
                data={userSearchResult.filter((v) => !sharedUsers.includes(v))}
                value={search}
                renderOption={renderAutocompleteOption}
                onChange={(s) => {
                    setSearch(s)
                    searchUsers(s)
                }}
                comboboxProps={{ dropdownPadding: 0 }}
                placeholder="Add people"
                style={{ width: '100%', padding: 8 }}
            />
            <Text
                c={isPublic ? '#777777' : 'white'}
                style={{
                    width: '100%',
                    textAlign: 'center',
                    marginTop: 10,
                    userSelect: 'none',
                }}
            >
                Shared with
            </Text>
            <div className="flex flex-col items-center h-max min-h-[33px] bg-[#00000044] mt-1 pt-2 pb-2">
                {sharedUsers.map((v) => {
                    return (
                        <div className="flex flex-row w-[90%] h-9 p-2" key={v}>
                            <IconUser color={isPublic ? '#777777' : 'white'} />
                            <Space w={10} />
                            <Text
                                c={isPublic ? '#777777' : 'white'}
                                style={{ userSelect: 'none' }}
                                size="20px"
                            >
                                {v}
                            </Text>
                            <Space style={{ display: 'flex', flexGrow: 1 }} />
                            <div
                                className="xBox"
                                style={{
                                    pointerEvents: isPublic ? 'none' : 'all',
                                }}
                                onClick={() => {
                                    setSharedUsers((u) => {
                                        u.splice(u.indexOf(v), 1)
                                        return [...u]
                                    })
                                }}
                            >
                                <IconX
                                    scale={'3px'}
                                    color={isPublic ? '#777777' : 'white'}
                                />
                            </div>
                        </div>
                    )
                })}
                {sharedUsers.length === 0 && (
                    <p className="h-max w-max select-none">Not shared</p>
                )}
            </div>
        </Box>
    )
}

export function ShareBox({
    candidates,
    authHeader,
}: {
    candidates: WeblensFile[]
    authHeader: AuthHeaderT
}) {
    const [sharedUsers, setSharedUsers] = useState([])
    const [pub, setPublic] = useState(false)
    const [shareData, setShareData]: [
        shareData: ShareDataT[],
        setShareData: (v) => void,
    ] = useState(null)
    useEffect(() => {
        if (
            !candidates ||
            candidates.length === 0 ||
            candidates[0].GetShares().length === 0
        ) {
            return
        }
        setShareData(candidates[0].GetShares().filter((s) => !s.IsWormhole()))
    }, [candidates, authHeader])

    useEffect(() => {
        if (!shareData || shareData.length === 0) {
            return
        }
        setPublic(shareData[0].Public)
        if (shareData[0].Accessors) {
            setSharedUsers(shareData[0].Accessors)
        }
    }, [shareData])

    const shareOrUpdate = useCallback(async () => {
        if (!shareData || shareData.length === 0) {
            const res = await shareFiles(
                candidates,
                pub,
                sharedUsers,
                authHeader
            ).catch((r) => {
                console.error(r)
                return false
            })
            if (!res) {
                return ''
            }
            return res.shareData.shareId
        } else {
            const res = await updateFileShare(
                shareData[0].shareId,
                pub,
                sharedUsers,
                authHeader
            ).catch((r) => {
                console.error(r)
                return false
            })
            if (!res) {
                return ''
            }
            return shareData[0].shareId
        }
    }, [candidates, shareData, pub, sharedUsers, authHeader])

    return (
        <div className="flex flex-col items-center gap-1 p-1">
            <ShareInput
                isPublic={pub}
                sharedUsers={sharedUsers}
                setSharedUsers={setSharedUsers}
            />
            <WeblensButton
                key={'public-button'}
                toggleOn={pub}
                allowRepeat
                squareSize={50}
                onClick={() => setPublic(!pub)}
                label={pub ? 'Public' : 'Private'}
                postScript={
                    pub
                        ? 'Anyone with link can access'
                        : 'Only shared users can access'
                }
                Left={pub ? <IconUsersGroup /> : <IconUserCancel />}
            />

            <Box
                style={{
                    display: 'flex',
                    flexDirection: 'row',
                    justifyContent: 'space-between',
                    width: '100%',
                }}
            >
                <WeblensButton
                    label="Copy link"
                    squareSize={40}
                    centerContent
                    Left={<IconLink />}
                    onClick={(e) => {
                        e.stopPropagation()
                        shareOrUpdate().then((v) => {
                            navigator.clipboard.writeText(
                                `${window.location.origin}/share/${v}`
                            )
                        })
                        return true
                    }}
                    // style={{ width: "50%" }}
                />
                <WeblensButton
                    label="Save"
                    squareSize={40}
                    centerContent
                    onClick={(e) => {
                        e.stopPropagation()
                        shareOrUpdate()
                        return true
                    }}
                    // style={{ width: '50%' }}
                />
            </Box>
        </div>
    )
}
