import { Autocomplete, Box, Divider, Space, Text } from "@mantine/core";
import { AutocompleteUsers, GetFileShare, ShareFiles, UpdateFileShare } from "../../api/FileBrowserApi";
import { IconLink, IconUser, IconUserCancel, IconUsersGroup, IconX } from "@tabler/icons-react";
import { WeblensButton } from "../../components/WeblensButton";
import { Combobox, Pill, useCombobox } from "@mantine/core"
import { useCallback, useContext, useEffect, useMemo, useState } from "react"
import { userContext } from "../../Context";
import { ColumnBox, RowBox } from "./FilebrowserStyles";
import { fileData, shareData } from "../../types/Types";
import { notifications } from "@mantine/notifications";

export function ShareInput({ isPublic, sharedUsers, setSharedUsers }: { isPublic: boolean, sharedUsers: string[], setSharedUsers: (v) => void }) {
    const { userInfo, authHeader } = useContext(userContext)
    const [userSearchResult, setUserSearch] = useState([])
    const [empty, setEmpty] = useState(false)
    const [loading, setLoading] = useState(false)
    const [search, setSearch] = useState('')

    const searchUsers = useCallback(async (query: string) => {
        console.log("HERE")
        if (query.length < 2) {
            setUserSearch([])
            setEmpty(true)
        }

        setLoading(true)
        let users: string[] = await AutocompleteUsers(query, authHeader)
        const selfIndex = users.indexOf(userInfo.username)
        if (selfIndex !== -1) {
            users.splice(selfIndex, 1)
        }
        users = users.filter(v => !sharedUsers.includes(v))
        setUserSearch(users)
        setLoading(false)
        setEmpty(users.length === 0)
    }, [sharedUsers, userInfo.username, authHeader])

    const renderAutocompleteOption = useCallback(({ option }) => {
        if (sharedUsers.includes(option.value)) {
            return
        }
        return (
            <Box style={{ height: '100%', width: '100%' }} onClick={e => { e.stopPropagation(); setSharedUsers(v => { v.push(option.value); return [...v] }) }}>
                <Text style={{ color: sharedUsers.includes(option.value) ? '#555555' : 'white' }}>{option.value}</Text>
            </Box>
        )
    }, [sharedUsers, setSharedUsers])

    return (
        <ColumnBox style={{ width: 300 }}>
            <Autocomplete disabled={isPublic} data={userSearchResult} value={search} renderOption={renderAutocompleteOption} onChange={s => { setSearch(s); searchUsers(s) }} comboboxProps={{ dropdownPadding: 0 }} placeholder="Add people" style={{ width: '100%', padding: 8 }} />
            <Text c={isPublic ? '#777777' : 'white'} style={{ width: '100%', textAlign: 'center', marginTop: 10, userSelect: 'none' }}>Shared with</Text>
            <ColumnBox style={{ height: 'max-content', minHeight: '100px', backgroundColor: '#00000044', width: '100%', display: 'flex', marginTop: 5, marginLeft: -10, marginRight: -10 }}>
                {sharedUsers.map(v => {
                    return (
                        <RowBox key={v} style={{ width: '90%', padding: 10 }}>
                            <IconUser color={isPublic ? '#777777' : 'white'} />
                            <Text c={isPublic ? '#777777' : 'white'} style={{ userSelect: 'none' }}>{v}</Text>
                            <Space style={{ display: 'flex', flexGrow: 1 }} />
                            <Box className="xBox" style={{ pointerEvents: isPublic ? 'none' : 'all' }} onClick={() => { setSharedUsers(u => { u.splice(u.indexOf(v), 1); console.log(u); return [...u] }) }}>
                                <IconX scale={'3px'} color={isPublic ? '#777777' : 'white'} />
                            </Box>
                        </RowBox>
                    )
                })}
            </ColumnBox>
        </ColumnBox>

        // <Combobox
        //     onOptionSubmit={str => { setSearch(''); handleValueSelect(str) }}
        //     withinPortal={false}
        //     store={combobox}
        // >
        //     <Combobox.DropdownTarget>
        //         <PillsInput
        //             label="People to share with"
        //             onClick={() => combobox.openDropdown()}
        //             rightSection={loading && <Loader size={18} />}
        //             placeholder='Search users to share with'
        //         >
        //             {values}
        //             <Combobox.EventsTarget>
        //                 <PillsInput.Field
        //                     value={search}
        //                     onChange={(e) => {
        //                         setSearch(e.currentTarget.value)
        //                         searchUsers(e.currentTarget.value)
        //                         combobox.updateSelectedOptionIndex()
        //                         combobox.openDropdown()
        //                     }}
        //                     onClick={(e) => { e.stopPropagation(); combobox.openDropdown() }}
        //                     onFocus={() => {
        //                         combobox.openDropdown()
        //                         if (userSearch === null) {
        //                             searchUsers(search)
        //                         }
        //                     }}
        //                     onBlur={() => combobox.closeDropdown()}
        //                     onKeyDown={(event) => {
        //                         if (event.key === 'Backspace' && search.length === 0) {
        //                             event.preventDefault()
        //                             handleValueRemove(value[value.length - 1])
        //                         }
        //                     }}
        //                 />

        //             </Combobox.EventsTarget>
        //         </PillsInput>
        //     </Combobox.DropdownTarget>
        //     <Combobox.Dropdown hidden={search === "" || search === null}>
        //         <Combobox.Options>
        //             {options}
        //             {(empty && !loading) && <Combobox.Empty>No results found</Combobox.Empty>}
        //         </Combobox.Options>
        //     </Combobox.Dropdown>
        // </Combobox>
    )
}

export function ShareBox({ candidates, authHeader }: { candidates: fileData[], authHeader }) {
    const [sharedUsers, setSharedUsers] = useState([])
    const [pub, setPublic] = useState(false)
    const [shareData, setShareData]: [shareData: shareData[], setShareData: (v) => void] = useState(null)

    useEffect(() => {
        if (candidates[0].shares.length === 0) {
            return
        }
        GetFileShare(candidates[0].id, authHeader).then((v: shareData[]) => setShareData(v.filter(s => !s.Wormhole))).catch(r => notifications.show({ title: "Failed to get share data", message: String(r), color: 'red' }))
    }, [candidates, authHeader])

    useEffect(() => {
        if (!shareData || shareData.length === 0) {
            return
        }
        setPublic(shareData[0].Public)
        setSharedUsers(shareData[0].Accessors)
    }, [shareData])

    const shareOrUpdate = useCallback(async () => {
        console.log(shareData)
        if (shareData.length === 0) {
            const res = await ShareFiles(candidates.map(v => v.id), pub, sharedUsers, authHeader)
            return res.shareId
        } else {
            await UpdateFileShare(shareData[0].shareId, pub, sharedUsers, authHeader)
            return shareData[0].shareId
        }
    }, [candidates, shareData, pub, sharedUsers, authHeader])

    return (
        <Box>
            <ShareInput isPublic={pub} sharedUsers={sharedUsers} setSharedUsers={setSharedUsers} />
            <WeblensButton toggleOn={pub} onToggle={setPublic} label={pub ? "Public" : "Private"} postScript={pub ? "Anyone with link can access" : "Only shared users can access"} Left={pub ? <IconUsersGroup /> : <IconUserCancel />} />

            <RowBox style={{ justifyContent: 'space-between' }}>
                <WeblensButton label="Copy link" Left={<IconLink />} onClick={e => { e.stopPropagation(); shareOrUpdate().then(v => navigator.clipboard.writeText(`${window.location.origin}/share/${v.shareId}`)) }} style={{ width: '40%' }} />
                <WeblensButton label="Save" onClick={e => { e.stopPropagation(); shareOrUpdate() }} style={{ width: '40%' }} />
            </RowBox>
        </Box>
    )
}