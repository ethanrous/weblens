<template>
    <div :class="{ 'bg-card-background-primary flex h-16 w-full shrink-0 items-center justify-between': true }">
        <div :class="{ 'flex min-w-max flex-1 flex-col': true }">
            <span :class="{ 'inline-flex items-center': true }">
                <IconChevronLeft
                    :class="{
                        'mx-1 inline-flex items-center justify-center rounded pr-0.5 leading-none transition md:mx-2': true,
                        'text-text-tertiary': !canNavigate,
                        'hover:bg-card-background-hover cursor-pointer': canNavigate,
                    }"
                    @click="navigateBack"
                />

                <h3
                    :class="{
                        'inline max-h-max cursor-pointer truncate text-lg text-nowrap select-none md:text-2xl': true,
                    }"
                    @contextmenu.stop.prevent="openContextMenu"
                    @click.stop="openContextMenu"
                >
                    {{ fileName }}
                </h3>
            </span>
        </div>

        <div :class="{ 'relative flex h-10 max-w-30 flex-[1.5] justify-center lg:max-w-125': true }">
            <Searchbar ref="searchbar" />
        </div>

        <div :class="{ 'relative mr-4 flex h-10 min-w-0 flex-1 flex-row items-center justify-end gap-2': true }">
            <Loader
                v-if="tasksStore.anyRunning"
                :class="{ 'mr-2': true }"
            />
            <FileSortControls v-if="!locationStore.isInTimeline" />
            <TimelineControls v-if="locationStore.isInTimeline" />

            <WeblensButton @click="locationStore.isInTimeline = !locationStore.isInTimeline">
                <IconFolder v-if="locationStore.isInTimeline" />
                <IconPhoto v-if="!locationStore.isInTimeline" />
            </WeblensButton>
        </div>
    </div>
</template>

<script setup lang="ts">
import { IconChevronLeft, IconFolder, IconPhoto } from '@tabler/icons-vue'
import useFilesStore from '~/stores/files'
import WeblensFile from '~/types/weblensFile'
import WeblensButton from '../atom/WeblensButton.vue'
import TimelineControls from '../molecule/TimelineControls.vue'
import { useMagicKeys, whenever } from '@vueuse/core'
import Searchbar from '../molecule/Searchbar.vue'
import FileSortControls from '../molecule/FileSortControls.vue'
import useLocationStore from '~/stores/location'
import Loader from '../atom/Loader.vue'

const filesStore = useFilesStore()
const locationStore = useLocationStore()
const menuStore = useContextMenuStore()
const userStore = useUserStore()
const tasksStore = useTasksStore()

const searchbar = ref<typeof Searchbar>()

const keys = useMagicKeys()

whenever(
    () => keys.Cmd_K?.value || keys.Ctrl_K?.value,
    () => {
        if (locationStore.isInTimeline) {
            return
        }

        filesStore.setSearchRecurively(keys.shift?.value || false)

        searchbar.value?.focus()
    },
)

const activeFile = computed(() => {
    return filesStore.activeFile
})

const fileName = computed(() => {
    return activeFile.value ? activeFile.value.GetFilename() : ''
})

const canNavigate = computed(() => {
    if (activeFile.value?.IsShareRoot()) {
        return false
    }

    if (activeFile.value?.IsSharedWithMe()) {
        return true
    }

    if (!activeFile.value || !activeFile.value.parentID) {
        return false
    }

    if (userStore.user.homeID === activeFile.value?.id || activeFile.value?.parentID === 'USERS') {
        return false
    }

    return true
})

function navigateBack() {
    if (!canNavigate.value) {
        return
    }

    // If in a share, and the file has no parents, go to share root (/files/share)
    if (activeFile.value?.IsSharedWithMe() && activeFile.value?.parents.length === 0) {
        return WeblensFile.ShareRoot().GoTo()
    }

    return new WeblensFile({ id: activeFile.value?.parentID, isDir: true }).GoTo()
}

function openContextMenu(e: MouseEvent) {
    if (!filesStore.activeFile?.ID()) return

    menuStore.setTarget(filesStore.activeFile?.ID())
    menuStore.setMenuOpen(true)

    const fbBox = document.getElementById('filebrowser-container')
    if (!fbBox) {
        console.error('File browser container not found')
        return
    }

    const rect = fbBox.getBoundingClientRect()

    menuStore.setMenuPosition({ x: e.pageX - rect.left, y: e.pageY - rect.top })
}
</script>
