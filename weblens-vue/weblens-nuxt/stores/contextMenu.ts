import { defineStore } from 'pinia'
import type { coordinates } from '~/types/style'

export type MenuMode = 'rename' | 'newName'

export const useContextMenuStore = defineStore('contextMenu', () => {
    const isOpen = ref<boolean>(false)
    const isSharing = ref<boolean>(false)
    const menuMode = ref<'rename' | 'newName' | undefined>()

    const menuPosition = ref<coordinates>({ x: -1, y: -1 })

    const directTargetID = ref<string>('')

    function setMenuOpen(open: boolean) {
        isOpen.value = open

        if (!open) {
            isSharing.value = false
            menuMode.value = undefined
            directTargetID.value = ''
        }
    }

    function setMenuMode(newMenuMode?: MenuMode) {
        menuMode.value = newMenuMode
    }

    function setSharing(sharing: boolean) {
        isOpen.value = false
        isSharing.value = sharing
    }

    function setMenuPosition(pos: coordinates) {
        menuPosition.value = pos
    }

    function setTarget(id: string) {
        directTargetID.value = id
        menuMode.value = undefined
    }

    return {
        isOpen,
        isSharing,
        menuPosition,
        directTargetID,
        menuMode,

        setMenuOpen,
        setSharing,
        setMenuPosition,
        setTarget,
        setMenuMode,
    }
})
