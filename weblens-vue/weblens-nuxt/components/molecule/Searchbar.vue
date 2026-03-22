<template>
    <div
        :class="{
            'ease-wl-default absolute top-0 flex h-10.5 w-full justify-center transition-[height,top,width] duration-300': true,
            'h-[min(80vh,424px)] max-lg:top-[calc((100vh-424px)/2)] max-lg:z-50 max-lg:w-[80vw]': filterOpen,
        }"
    >
        <div
            ref="searchbar"
            :class="{
                'bg-background-primary relative mx-4 flex flex-col justify-center overflow-hidden rounded-lg border transition-[width,height] duration-300 lg:w-full': true,
                'w-20 shadow-sm': !filterOpen,
                'w-[80vw] md:w-[50vw]': filterOpen,
            }"
        >
            <WeblensInput
                ref="searchInput"
                v-model:value="locationStore.search"
                :class="{
                    'bg-background-primary h-10! min-h-10! w-full shrink gap-0! border-none': true,
                    'bg-background-primary!': filterOpen,
                    'justify-between lg:justify-center': !filterOpen,
                }"
                :input-class="{
                    'lg:ml-2 h-5': true,
                    'ml-2': filterOpen,
                    'max-w-0 lg:max-w-full': !filterOpen,
                }"
                :placeholder="searchText"
                :key-name="keyHintText"
                :buttonish="!filterOpen"
                search
                :clear-button="filterOpen || windowSize.width.value >= 1024"
                @focused="handleSearchFocused"
                @submit="handleSubmit"
            >
                <div class="relative">
                    <IconSearch
                        size="20"
                        :class="{ 'text-text-tertiary': true }"
                    />
                    <div
                        :class="{
                            'bg-theme-primary absolute top-0 right-0 h-1.5 w-1.5 rounded-[1px] transition': true,
                            'opacity-0': !hasCollapsedSearch,
                        }"
                    />
                </div>
                <template #rightIcon="slotProps">
                    <div
                        v-if="!slotProps.focused && !filterOpen && locationStore.search === ''"
                        :class="{
                            'text-text-tertiary pointer-events-none mr-1 max-h-5 leading-none text-nowrap transition': true,
                            'hidden lg:inline-flex': true,
                        }"
                    >
                        <span>
                            {{ keyHintText }}
                        </span>
                    </div>
                    <div
                        :class="{ 'relative ml-1 flex justify-center border-l pl-2.5': true }"
                        @click.stop.exact="
                            () => {
                                filterOpen = !filterOpen
                            }
                        "
                        @click.shift="() => filesStore.clearSearch()"
                    >
                        <IconFilter2
                            size="20"
                            :class="{
                                'hover:text-text-primary cursor-pointer transition': true,
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

                    <WeblensButton
                        v-if="
                            (locationStore.isInTimeline && !mediaStore.searchUpToDate) ||
                            (!locationStore.isInTimeline && !filesStore.searchUpToDate)
                        "
                        type="light"
                        :square-size="30"
                        :class="{ 'ml-2': true }"
                        @click.stop="handleSubmit"
                    >
                        <IconArrowRight />
                    </WeblensButton>
                </template>
            </WeblensInput>

            <div
                ref="searchFilter"
                :class="{
                    'bg-background-primary relative z-50 h-full w-full overflow-hidden rounded-b shadow-lg outline-none': true,
                    'pointer-events-none': !filterOpen,
                }"
                tabindex="0"
                @keydown.esc="
                    (e) => {
                        if (filterOpen) {
                            e.stopPropagation()
                            filterOpen = false
                        }
                    }
                "
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
import { IconArrowRight, IconFilter2, IconSearch } from '@tabler/icons-vue'
import WeblensInput from '../atom/WeblensInput.vue'
import { onClickOutside, useElementSize, useWindowSize } from '@vueuse/core'
import useFilesStore from '~/stores/files'
import FileSearchFilters from './FileSearchFilters.vue'
import MediaSearchFilters from './MediaSearchFilters.vue'
import useLocationStore from '~/stores/location'
import WeblensButton from '../atom/WeblensButton.vue'

const windowSize = useWindowSize()

const filesStore = useFilesStore()
const locationStore = useLocationStore()
const mediaStore = useMediaStore()

const filterOpen = ref(false)

const searchInput = ref<typeof WeblensInput>()

const searchFilter = ref<HTMLDivElement>()
const searchbar = ref<HTMLDivElement>()

const searchElement = computed(() => {
    return searchInput.value?.$el
})

onClickOutside(
    searchbar,
    () => {
        if (!filterOpen.value) return
        filterOpen.value = false
    },
    { ignore: [searchElement] },
)

const hasCollapsedSearch = computed(() => {
    return !filterOpen.value && windowSize.width.value < 1024 && locationStore.search !== ''
})

const filterModified = computed(() => {
    if (locationStore.isInTimeline) {
        return mediaStore.showRaw === false
    } else {
        return filesStore.searchRecursively || filesStore.filterTagIDs.size > 0
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

watchEffect(() => {
    if (filterOpen.value) {
        nextTick().then(() => {
            searchFilter.value?.focus()
        })
    }
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

async function handleSubmit() {
    filterOpen.value = false

    if (locationStore.isInTimeline) {
        mediaStore.clearData()
        mediaStore.fetchMoreMedia()
    } else {
        filesStore.searchUpToDate = true
    }
}

defineExpose({
    focus: () => {
        searchInput.value?.focus()
    },
})
</script>
