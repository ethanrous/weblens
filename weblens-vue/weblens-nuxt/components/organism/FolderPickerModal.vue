<template>
    <Teleport to="body">
        <div
            v-if="visible"
            class="fixed inset-0 z-9998 bg-black/50"
            @click.self="$emit('close')"
        />

        <div
            v-if="visible"
            class="fixed inset-0 z-9999 flex items-center justify-center p-8"
            @click.self="$emit('close')"
        >
            <div
                class="bg-background-primary flex max-h-[60vh] w-full max-w-md flex-col gap-4 rounded-lg border p-6 shadow-lg"
                @click.stop
            >
                <h4>Select Folder</h4>

                <WeblensInput
                    ref="searchInput"
                    v-model:value="displayPath"
                    placeholder="File path..."
                    @update:value="onSearchInput"
                />

                <div class="flex max-h-64 flex-col gap-1 overflow-y-auto">
                    <div
                        v-if="currentFolder"
                        class="flex items-center gap-2 rounded border px-3 py-2"
                    >
                        <IconFolder
                            :size="16"
                            class="shrink-0 opacity-50"
                        />
                        <span class="truncate text-sm opacity-70">{{ currentFolderName }}</span>
                        <WeblensButton
                            label="Select"
                            type="outline"
                            class="ml-auto"
                            :on-click="selectCurrentFolder"
                        />
                    </div>

                    <div
                        v-for="(folder, i) in folderResults"
                        :key="folder.id"
                        :class="{
                            'hover:bg-background-secondary flex cursor-pointer items-center gap-2 rounded px-3 py-2 transition': true,
                            'bg-background-secondary': selectedResult === i,
                        }"
                        @click="navigateInto(folder)"
                    >
                        <IconFolder
                            :size="16"
                            class="shrink-0"
                        />
                        <span class="truncate text-sm">{{ folderDisplayName(folder) }}</span>
                        <IconChevronRight
                            :size="14"
                            class="ml-auto shrink-0 opacity-40"
                        />
                    </div>

                    <div
                        v-if="errorMessage && !loading"
                        class="text-danger py-4 text-center text-sm"
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

                <div class="flex justify-end">
                    <WeblensButton
                        label="Cancel"
                        type="outline"
                        :on-click="() => emit('close')"
                    />
                </div>
            </div>
        </div>
    </Teleport>
</template>

<script setup lang="ts">
import { IconChevronRight, IconFolder } from '@tabler/icons-vue'
import { onKeyDown, useDebounceFn } from '@vueuse/core'
import type { FileInfo } from '@ethanrous/weblens-api'
import WeblensButton from '../atom/WeblensButton.vue'
import { useWeblensAPI } from '~/api/AllApi'
import { PortablePath } from '~/types/portablePath'
import WeblensInput from '../atom/WeblensInput.vue'

const props = defineProps<{
    suggestedPath?: PortablePath
    visible: boolean
}>()

const searchInput = ref<HTMLInputElement>()
const initialPath = props.suggestedPath ?? PortablePath.Home()

const displayPath = ref<string>(initialPath.friendlyPath)
const folderResults = ref<FileInfo[]>([])
const selectedResult = ref<number>(-1)
const loading = ref(false)
const errorMessage = ref('')

onKeyDown(['ArrowUp', 'ArrowDown', 'Tab', 'Enter'], (e) => {
    if ((selectedResult.value === -1 || folderResults.value.length === 0) && e.key === 'Enter') {
        selectCurrentFolder()
        return
    } else if (folderResults.value.length === 0) {
        return
    }

    e.preventDefault()
    if (e.key === 'ArrowUp') {
        selectedResult.value = Math.max(selectedResult.value - 1, -1)
    } else if (e.key === 'ArrowDown') {
        selectedResult.value = Math.min(selectedResult.value + 1, folderResults.value.length - 1)
    } else if (
        (e.key === 'Enter' || e.key === 'Tab') &&
        selectedResult.value >= 0 &&
        selectedResult.value < folderResults.value.length
    ) {
        navigateInto(folderResults.value[selectedResult.value])
        selectedResult.value = -1
    }
})

const currentFolder = ref<FileInfo | undefined>(
    props.suggestedPath ? { portablePath: props.suggestedPath.toString() } : undefined,
)

const emit = defineEmits<{
    (e: 'close'): void
    (e: 'select', folderID: string): void
}>()

const currentFolderName = computed(() => {
    if (!currentFolder.value?.portablePath) return ''
    return new PortablePath(currentFolder.value.portablePath).friendlyPath
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
    try {
        const portable = PortablePath.fromFriendly(displayPath.value)
        debouncedFetch(portable.toString())
    } catch {
        // The user is mid-edit and the path is not yet a valid friendly path
        // (e.g. they cleared the input). Drop results until they type more.
        folderResults.value = []
        currentFolder.value = undefined
        errorMessage.value = ''
    }
}

function navigateInto(folder: FileInfo) {
    if (!folder.portablePath) return

    const folderPath = new PortablePath(
        folder.portablePath.endsWith('/') ? folder.portablePath : folder.portablePath + '/',
    )

    displayPath.value = folderPath.friendlyPath
    fetchAutocomplete(folderPath.toString())
}

function selectCurrentFolder() {
    if (!currentFolder.value?.id) return
    emit('select', currentFolder.value.id)
}

watch(
    () => props.visible,
    (vis) => {
        if (vis) {
            const target = props.suggestedPath ?? PortablePath.Home()
            displayPath.value = target.friendlyPath
            fetchAutocomplete(target.toString())
            nextTick(() => searchInput.value?.focus())
        } else {
            folderResults.value = []
            currentFolder.value = undefined
        }
    },
    { immediate: true },
)
</script>
