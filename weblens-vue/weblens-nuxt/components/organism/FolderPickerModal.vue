<template>
    <Teleport to="body">
        <div
            v-if="visible"
            class="fixed inset-0 z-[9998] bg-black/50"
            @click.self="$emit('close')"
        />

        <div
            v-if="visible"
            class="fixed inset-0 z-[9999] flex items-center justify-center p-8"
            @click.self="$emit('close')"
        >
            <div
                class="bg-background-primary flex max-h-[60vh] w-full max-w-md flex-col gap-4 rounded-lg border p-6 shadow-lg"
                @click.stop
            >
                <h4>Select Folder</h4>

                <input
                    ref="searchInput"
                    v-model="displayPath"
                    class="bg-background-secondary w-full rounded border px-3 py-2 text-sm outline-none focus:border-blue-500"
                    placeholder="Type a folder path..."
                    @input="onSearchInput"
                />

                <div class="flex max-h-64 flex-col gap-1 overflow-y-auto">
                    <div
                        v-if="currentFolder"
                        class="flex items-center gap-2 rounded border px-3 py-2"
                    >
                        <IconFolder :size="16" class="shrink-0 opacity-50" />
                        <span class="truncate text-sm opacity-70">{{ currentFolderName }}</span>
                        <div class="ml-auto" @click="selectCurrentFolder">
                            <WeblensButton
                                label="Select"
                                type="outline"
                            />
                        </div>
                    </div>

                    <div
                        v-for="folder in folderResults"
                        :key="folder.id"
                        class="hover:bg-background-secondary flex cursor-pointer items-center gap-2 rounded px-3 py-2 transition"
                        @click="navigateInto(folder)"
                    >
                        <IconFolder :size="16" class="shrink-0" />
                        <span class="truncate text-sm">{{ folderDisplayName(folder) }}</span>
                        <IconChevronRight :size="14" class="ml-auto shrink-0 opacity-40" />
                    </div>

                    <div
                        v-if="errorMessage && !loading"
                        class="text-red-400 py-4 text-center text-sm"
                    >
                        {{ errorMessage }}
                    </div>

                    <div
                        v-else-if="folderResults.length === 0 && !loading"
                        class="text-text-tertiary py-4 text-center text-sm"
                    >
                        {{ displayPath.length > 0 ? 'No matching folders' : 'Start typing to search' }}
                    </div>

                    <div
                        v-if="loading"
                        class="text-text-tertiary py-4 text-center text-sm"
                    >
                        Searching...
                    </div>
                </div>

                <div class="flex justify-end" @click="$emit('close')">
                    <WeblensButton
                        label="Cancel"
                        type="outline"
                    />
                </div>
            </div>
        </div>
    </Teleport>
</template>

<script setup lang="ts">
import { IconChevronRight, IconFolder } from '@tabler/icons-vue'
import { useDebounceFn } from '@vueuse/core'
import type { FileInfo } from '@ethanrous/weblens-api'
import WeblensButton from '../atom/WeblensButton.vue'
import { useWeblensAPI } from '~/api/AllApi'

const props = defineProps<{
    visible: boolean
}>()

const emit = defineEmits<{
    (e: 'close'): void
    (e: 'select', folderID: string): void
}>()

const userStore = useUserStore()
const searchInput = ref<HTMLInputElement>()

const homePrefix = computed(() => `USERS:${userStore.user.username}/`)

function toDisplayPath(portablePath: string): string {
    if (portablePath.startsWith(homePrefix.value)) {
        return '~/' + portablePath.slice(homePrefix.value.length)
    }
    return portablePath
}

function toPortablePath(display: string): string {
    if (display.startsWith('~/')) {
        return homePrefix.value + display.slice(2)
    }
    return display
}

const displayPath = ref('~/')
const folderResults = ref<FileInfo[]>([])
const currentFolder = ref<FileInfo>()
const loading = ref(false)
const errorMessage = ref('')

const currentFolderName = computed(() => {
    if (!currentFolder.value?.portablePath) return ''
    return toDisplayPath(currentFolder.value.portablePath).replace(/\/$/, '').split('/').pop() || '~'
})

function folderDisplayName(f: FileInfo): string {
    if (!f.portablePath) return f.id ?? ''
    const parts = f.portablePath.split('/')
    return parts.filter(Boolean).pop() ?? ''
}

async function fetchAutocomplete(portable: string) {
    if (!portable || !portable.includes(':')) return

    loading.value = true

    try {
        const res = await useWeblensAPI().FilesAPI.autocompletePath(portable)
        const data = res.data

        currentFolder.value = data.self
        folderResults.value = (data.children ?? []).filter((c) => c.isDir)
        errorMessage.value = ''
    } catch {
        folderResults.value = []
        currentFolder.value = undefined
        errorMessage.value = 'Could not load this path'
    } finally {
        loading.value = false
    }
}

const debouncedFetch = useDebounceFn((portable: string) => fetchAutocomplete(portable), 250)

function onSearchInput() {
    debouncedFetch(toPortablePath(displayPath.value))
}

function navigateInto(folder: FileInfo) {
    if (!folder.portablePath) return

    const portable = folder.portablePath.endsWith('/')
        ? folder.portablePath
        : folder.portablePath + '/'

    displayPath.value = toDisplayPath(portable)
    fetchAutocomplete(portable)
}

function selectCurrentFolder() {
    if (!currentFolder.value?.id) return
    emit('select', currentFolder.value.id)
}

watch(() => props.visible, (vis) => {
    if (vis) {
        displayPath.value = '~/'
        fetchAutocomplete(homePrefix.value)
        nextTick(() => searchInput.value?.focus())
    } else {
        folderResults.value = []
        currentFolder.value = undefined
    }
}, { immediate: true })
</script>
