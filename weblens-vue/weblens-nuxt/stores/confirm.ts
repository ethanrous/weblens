import { defineStore } from 'pinia'

const useConfirmDialog = defineStore('confirm', () => {
    const isOpen = ref(false)
    const verb = ref('')
    const itemName = ref('')
    const callback = ref<(() => void | Promise<void>) | null>(null)

    function open({
        actionVerb = '',
        actionItemName = '',
        onAccept = null,
    }: {
        actionVerb?: string
        actionItemName?: string
        onAccept?: (() => void | Promise<void>) | null
    } = {}) {
        isOpen.value = true

        verb.value = actionVerb
        itemName.value = actionItemName
        callback.value = onAccept
    }

    function close() {
        isOpen.value = false
        verb.value = ''
        itemName.value = ''
        callback.value = null
    }

    async function accept() {
        await callback.value?.()

        close()
    }

    return {
        isOpen,
        verb,
        itemName,
        callback,

        open,
        accept,
        cancel: close,
    }
})

export default useConfirmDialog
