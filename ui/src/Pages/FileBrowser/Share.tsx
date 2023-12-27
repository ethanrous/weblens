import { Button, Combobox, Loader, Modal, Pill, PillsInput, Space, useCombobox } from "@mantine/core"
import { useContext, useEffect, useState } from "react"
import { AutocompleteUsers, ShareFiles } from "../../api/FileBrowserApi"
import { itemData } from "../../types/Types"
import { userContext } from "../../Context"

export function ShareInput({ valueSetCallback }) {
    const combobox = useCombobox({
        onDropdownClose: () => combobox.resetSelectedOption(),
    })
    const { userInfo, authHeader } = useContext(userContext)
    const [userSearch, setUserSearch] = useState(null)
    const [empty, setEmpty] = useState(false)
    const [loading, setLoading] = useState(false)
    const [search, setSearch] = useState('')
    const [value, setValue] = useState([])

    useEffect(() => {
        valueSetCallback(value)
    }, [value])

    const searchUsers = async (query: string) => {
        if (query.length < 2) {
            setUserSearch([])
            setEmpty(true)
        }

        setLoading(true)
        const users: string[] = await AutocompleteUsers(query, authHeader)
        const selfIndex = users.indexOf(userInfo.username)
        if (selfIndex !== -1) {
            users.splice(selfIndex, 1)
        }
        setUserSearch(users)
        setLoading(false)
        setEmpty(users.length === 0)
    }

    const options = (userSearch || []).map((item) => (
        <Combobox.Option value={item} key={item}>
            {item}
        </Combobox.Option>
    ))

    useEffect(() => {
        combobox.selectFirstOption()
    }, [userSearch])

    const handleValueSelect = (val: string) =>
        setValue((current) => { searchUsers(""); return current.includes(val) ? current.filter((v) => v !== val) : [...current, val] })

    const handleValueRemove = (val: string) =>
        setValue((current) => current.filter((v) => v !== val))

    const values = value.map((item) => (
        <Pill key={item} withRemoveButton onRemove={() => handleValueRemove(item)}>
            {item}
        </Pill>
    ))

    return (
        <Combobox
            onOptionSubmit={str => { setSearch(''); handleValueSelect(str) }}
            withinPortal={false}
            store={combobox}
        >
            <Combobox.DropdownTarget>
                <PillsInput
                    label="People to share with"
                    onClick={() => combobox.openDropdown()}
                    rightSection={loading && <Loader size={18} />}
                    placeholder='Search users to share with'
                >
                    {values}
                    <Combobox.EventsTarget>
                        <PillsInput.Field
                            value={search}
                            onChange={(e) => {
                                setSearch(e.currentTarget.value)
                                searchUsers(e.currentTarget.value)
                                combobox.updateSelectedOptionIndex()
                                combobox.openDropdown()
                            }}
                            onClick={(e) => { e.stopPropagation(); combobox.openDropdown() }}
                            onFocus={() => {
                                combobox.openDropdown()
                                if (userSearch === null) {
                                    searchUsers(search)
                                }
                            }}
                            onBlur={() => combobox.closeDropdown()}
                            onKeyDown={(event) => {
                                if (event.key === 'Backspace' && search.length === 0) {
                                    event.preventDefault()
                                    handleValueRemove(value[value.length - 1])
                                }
                            }}
                        />

                    </Combobox.EventsTarget>
                </PillsInput>
            </Combobox.DropdownTarget>
            <Combobox.Dropdown hidden={search === "" || search === null}>
                <Combobox.Options>
                    {options}
                    {(empty && !loading) && <Combobox.Empty>No results found</Combobox.Empty>}
                </Combobox.Options>
            </Combobox.Dropdown>
        </Combobox>
    )
}

function ShareDialogue({ sharing, selectedMap, dirMap, dispatch }) {
    const combobox = useCombobox({
        onDropdownClose: () => combobox.resetSelectedOption(),
    })
    const { authHeader } = useContext(userContext)
    const [value, setValue] = useState([])

    return (
        <Modal opened={sharing} onClose={() => { dispatch({ type: "close_share" }) }} title={`Share ${selectedMap.size} Files`} centered>
            <ShareInput valueSetCallback={setValue} />
            <Space h={'md'} />
            <Button onClick={() => ShareFiles(Array.from(selectedMap.keys()).map((key: string) => { const item: itemData = dirMap.get(key); return { parentFolderId: item.parentFolderId, filename: item.filename } }), value, authHeader)}>Share</Button>
        </Modal>
    )
}

export default ShareDialogue