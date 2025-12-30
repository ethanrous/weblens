<template>
    <div
        :class="{
            'ease-wl-default absolute top-0 z-50 flex h-10 w-full justify-center transition-[height,top,width] duration-300': true,
            'h-[min(80vh,424px)] max-lg:top-[calc((100vh-424px)/2)] max-lg:w-[80vw]': filterOpen,
        }"
    >
        <div
            :class="{
                'bg-background-primary relative mx-4 flex flex-col overflow-hidden rounded-lg transition-[width,height] duration-300 lg:w-full': true,
                'w-20 shadow-sm': !filterOpen,
                'w-[80vw] md:w-[50vw]': filterOpen,
            }"
        >
            <WeblensInput
                ref="searchInput"
                v-model:value="filesStore.fileSearch"
                :class="{
                    'bg-background-primary !h-10 w-full shrink-1 rounded-none border-b-0': true,
                    '!bg-background-primary': filterOpen,
                }"
                :placeholder="searchText"
                :key-name="keyHintText"
                :buttonish="!filterOpen"
                search
                clear-button
                @focused="handleSearchFocused"
                @submit="handleSubmit"
                @clear="handleSubmit('')"
                @update="
                    (v) => {
                        if (!locationStore.isInTimeline) {
                            filesStore.setFileSearch(v)
                        }
                    }
                "
            >
                <IconSearch
                    size="20"
                    :class="{ 'text-text-tertiary': true }"
                />
                <template #rightIcon>
                    <div
                        :class="{ 'relative flex justify-center border-l pl-2': true }"
                        @click.stop="
                            () => {
                                filterOpen = !filterOpen
                            }
                        "
                    >
                        <IconFilter2
                            size="20"
                            :class="{
                                'hover:text-text-primary transition': true,
                                'text-text-tertiary': !filterModified,
                                'text-text-secondary': filterModified,
                            }"
                        />
                        <div
                            :class="{
                                'bg-theme-primary absolute top-0 right-0 h-1.5 w-1.5 rounded-[1px] transition': true,
                                'opacity-0': !filterModified,
                            }"
                        />
                    </div>
                </template>
            </WeblensInput>
            <div
                ref="searchFilter"
                :class="{
                    'bg-background-primary relative z-50 h-full w-full overflow-hidden rounded-b border border-t-0 shadow-lg': true,
                    'pointer-events-none': !filterOpen,
                }"
            >
                <FileSearchFilters
                    v-if="!locationStore.isInTimeline"
                    @done="closeFilters(windowSize.width.value >= 1024)"
                />
                <MediaSearchFilters
                    v-if="locationStore.isInTimeline"
                    @done="closeFilters(false)"
                />
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { IconFilter2, IconSearch } from '@tabler/icons-vue'
import WeblensInput from '../atom/WeblensInput.vue'
import { onClickOutside, useElementSize, useWindowSize } from '@vueuse/core'
import useFilesStore from '~/stores/files'
import FileSearchFilters from './FileSearchFilters.vue'
import MediaSearchFilters from './MediaSearchFilters.vue'
import useLocationStore from '~/stores/location'

const windowSize = useWindowSize()

const filesStore = useFilesStore()
const locationStore = useLocationStore()
const mediaStore = useMediaStore()

const filterOpen = ref(false)

const searchInput = ref<typeof WeblensInput>()

const searchFilter = ref<HTMLDivElement>()
const searchElement = computed(() => {
    return searchInput.value?.$el
})

onClickOutside(
    searchFilter,
    () => {
        if (!filterOpen.value) return
        filterOpen.value = false
    },
    { ignore: [searchElement] },
)

const filterModified = computed(() => {
    if (locationStore.isInTimeline) {
        return mediaStore.showRaw === false
    } else {
        return filesStore.searchRecurively
    }
})

const containerSize = useElementSize(searchElement)

const searchText = computed(() => {
    if (locationStore.isInTimeline) {
        return 'Search Media...'
    } else {
        return 'Search Files...'
    }
})

const keyHintText = computed(() => {
    if (containerSize.width.value < 200) {
        return ''
    }

    if (locationStore.operatingSystem === 'macos') {
        return '⌘K'
    }

    return 'Ctrl+K'
})

function closeFilters(doFocus: boolean) {
    filterOpen.value = false

    if (doFocus) {
        searchInput.value?.focus()
    }
}

function handleSearchFocused() {
    if (windowSize.width.value < 1024) {
        filterOpen.value = true
    }
}

async function handleSubmit(v: string) {
    filterOpen.value = false

    if (locationStore.isInTimeline) {
        mediaStore.setImageSearch(v)
    } else {
        filesStore.setFileSearch(v)
        filesStore.setLoading(true)
        await filesStore.doSearch()
        filesStore.setLoading(false)
    }
}

defineExpose({
    focus: () => {
        searchInput.value?.focus()
    },
})
</script>
