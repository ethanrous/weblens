import { useLocalStorage } from '@vueuse/core'
import { defineStore } from 'pinia'

export const usePresentationStore = defineStore('presentation', () => {
    const presentationFileID = ref('')
    const presentationMediaID = ref('')

    const onMovePresentation = ref<((direction: number) => void) | null>(null)

    const infoOpen = useLocalStorage('presentationInfoOpen', true)

    function setPresentationFileID(newID: string) {
        presentationFileID.value = newID
    }

    function setPresentationMediaID(newID: string) {
        presentationMediaID.value = newID
    }

    function setOnMovePresentation(cb: (direction: number) => void) {
        onMovePresentation.value = cb
    }

    function clearPresentation() {
        presentationFileID.value = ''
        presentationMediaID.value = ''
    }

    return {
        presentationFileID,
        presentationMediaID,
        onMovePresentation,
        infoOpen,
        setPresentationFileID,
        setPresentationMediaID,
        setOnMovePresentation,
        clearPresentation,
    }
})
