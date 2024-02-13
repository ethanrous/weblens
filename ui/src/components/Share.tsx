import { Combobox, Loader, Pill, PillsInput, Space, useCombobox } from "@mantine/core"
import { useCallback, useContext, useEffect, useMemo, useState } from "react"
import { AutocompleteUsers } from "../api/FileBrowserApi"
import { userContext } from "../Context"

export function ShareInput({ valueSetCallback, initValues }: { valueSetCallback, initValues?}) {
    const combobox = useCombobox({
        onDropdownClose: () => combobox.resetSelectedOption(),
    })
    const { userInfo, authHeader } = useContext(userContext)
    const [userSearch, setUserSearch] = useState(null)
    const [empty, setEmpty] = useState(false)
    const [loading, setLoading] = useState(false)
    const [search, setSearch] = useState('')
    const [value, setValue] = useState(initValues)

    useEffect(() => {
        valueSetCallback(value)
    }, [value])

    const searchUsers = useCallback(async (query: string) => {
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
    }, [])

    const handleValueSelect = useCallback((val: string) => setValue((current) => { searchUsers(""); return current.includes(val) ? current.filter((v) => v !== val) : [...current, val] }), [])
    const handleValueRemove = useCallback((val: string) => setValue((current) => current.filter((v) => v !== val)), [])

    const options = useMemo(() => {
        combobox.selectFirstOption()
        return (userSearch || []).map((user) => (
            <Combobox.Option value={user} key={user}>
                {user}
            </Combobox.Option>
        ))
    }, [userSearch])

    const values = useMemo(() =>
        value.map((user) => (
            <Pill key={user} withRemoveButton onRemove={() => handleValueRemove(user)}>
                {user}
            </Pill>
        )), [value])

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