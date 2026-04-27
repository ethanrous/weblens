<template>
    <div
        :class="{
            'xs:hidden hover:text-text-primary text-text-tertiary absolute top-1/2 left-0 z-70 flex max-w-4 cursor-pointer justify-center transition-[left] duration-300': true,
            'bg-card-background-primary hover:bg-card-background-hover border-card-background-hover left-60 max-w-8 rounded border shadow-sm':
                forceOpen,
        }"
        @click="forceOpen = !forceOpen"
    >
        <IconChevronRight v-if="!forceOpen" />
        <IconChevronLeft v-if="forceOpen" />
    </div>

    <div
        id="global-left-sidebar"
        :class="{
            'bg-background-primary xs:static absolute z-60 flex h-full w-0 shrink-0 flex-col gap-1 overflow-hidden border py-4 transition-[width,padding] duration-300': true,
            'xs:w-64': !collapsed,
            'xs:w-16': collapsed,
            'xs:py-2 w-64': forceOpen,
        }"
    >
        <div
            :class="{
                'xs:static absolute flex h-full w-0 flex-col gap-2 px-4 transition-[width] duration-300': true,
                'xs:w-64 xs:px-4': !collapsed,
                'xs:w-16 xs:px-2': collapsed,
                'w-64! px-4 duration-100!': forceOpen,
            }"
        >
            <WeblensButton
                label="Home"
                :type="locationStore.isInSettings ? 'outline' : 'light'"
                :selected="filesStore.activeFile?.IsHome()"
                allow-collapse
                fill-width
                :disabled="!userStore.loggedIn"
                @click.stop="goHome()"
            >
                <IconFolder
                    v-if="locationStore.isInSettings"
                    size="18"
                />
                <IconHome
                    v-else
                    size="18"
                />
            </WeblensButton>

            <WeblensButton
                label="Shared"
                :type="'light'"
                :selected="locationStore.inShareRoot"
                allow-collapse
                fill-width
                :disabled="!userStore.loggedIn || locationStore.isInSettings"
                @click.stop="WeblensFile.ShareRoot().GoTo()"
            >
                <IconUsers size="18" />
            </WeblensButton>

            <WeblensButton
                label="Trash"
                :type="'light'"
                :selected="filesStore.activeFile?.IsTrash()"
                allow-collapse
                fill-width
                :disabled="!userStore.loggedIn || locationStore.isInSettings"
                @click.stop="WeblensFile.Trash().GoTo()"
            >
                <IconTrash size="18" />
            </WeblensButton>

            <Divider
                label="Actions"
                label-justify="left"
            />

            <WeblensButton
                label="New Folder"
                allow-collapse
                fill-width
                :disabled="
                    contextMenuStore.menuMode === 'newName' ||
                    !userStore.loggedIn ||
                    locationStore.isViewingPast ||
                    locationStore.isInSettings
                "
                @click.stop="handleNewFolder"
            >
                <IconFolderPlus size="18" />
            </WeblensButton>

            <UploadButton
                label="Upload"
                allow-collapse
                fill-width
                :disabled="!userStore.loggedIn || locationStore.isViewingPast || locationStore.isInSettings"
                @files-selected="handleUpload"
            >
                <IconUpload size="18" />
            </UploadButton>

            <Divider
                label="Tags"
                label-justify="left"
            />

            <WeblensButton
                label="Manage Tags"
                allow-collapse
                fill-width
                :disabled="!userStore.loggedIn || locationStore.isInSettings"
                @click.stop="showTagManager = true"
            >
                <IconTags size="18" />
            </WeblensButton>

            <div
                v-if="tagsStore.tagsList.length > 0 && !collapsed"
                :class="{ 'flex flex-col gap-0.5': true }"
            >
                <WeblensButton
                    v-for="tag in tagsStore.tagsList"
                    :key="tag.id"
                    :label="tag.name"
                    type="light"
                    :selected="activeTagID === tag.id"
                    allow-collapse
                    fill-width
                    :disabled="!userStore.loggedIn || locationStore.isInSettings"
                    @click.stop="navigateTo(`/files/tag/${tag.id}`)"
                >
                    <span
                        :class="{ 'h-3 w-3 shrink-0 rounded-full': true }"
                        :style="{ backgroundColor: tag.color }"
                    />
                </WeblensButton>
            </div>

            <Divider />

            <div :class="{ 'mt-auto w-full': true }">
                <TaskProgress />
            </div>

            <WeblensButton
                v-if="userStore.loggedIn"
                label="Settings"
                :type="(route.name as string).startsWith('settings') ? 'default' : 'outline'"
                fill-width
                allow-collapse
                @click.stop="goToSettings()"
            >
                <IconSettings size="18" />
            </WeblensButton>

            <WeblensButton
                v-else
                label="Log In"
                fill-width
                allow-collapse
                @click.stop="goToLogin()"
            >
                <IconUser size="18" />
            </WeblensButton>
        </div>
    </div>

    <TagManager
        :visible="showTagManager"
        @close="showTagManager = false"
    />
</template>

<script setup lang="ts">
import {
    IconChevronLeft,
    IconChevronRight,
    IconFolder,
    IconFolderPlus,
    IconHome,
    IconSettings,
    IconTags,
    IconTrash,
    IconUpload,
    IconUser,
    IconUsers,
} from '@tabler/icons-vue'
import WeblensButton from '../atom/WeblensButton.vue'
import Divider from '../atom/Divider.vue'
import TagManager from './TagManager.vue'
import useFilesStore from '~/stores/files'
import useTagsStore from '~/stores/tags'
import WeblensFile from '~/types/weblensFile'
import TaskProgress from './TaskProgress.vue'
import UploadButton from '../molecule/UploadButton.vue'
import { HandleFileSelect } from '~/api/uploadApi'
import { useWindowSize } from '@vueuse/core'
import useLocationStore from '~/stores/location'

const windowSize = useWindowSize()
watch(windowSize.width, (size: number) => {
    if (size >= 480) {
        forceOpen.value = false
    }
})

const filesStore = useFilesStore()
const contextMenuStore = useContextMenuStore()
const locationStore = useLocationStore()
const userStore = useUserStore()
const tagsStore = useTagsStore()
const route = useRoute()

const activeTagID = computed(() => {
    if ((route.name as string)?.startsWith('files-tag')) {
        return route.params.tagID as string
    }
    return ''
})

const forceOpen = ref<boolean>(false)
const showTagManager = ref(false)

defineProps<{
    collapsed?: boolean
}>()

async function handleUpload(files: FileList) {
    await HandleFileSelect(files, locationStore.activeFolderID, false, locationStore.activeShareID ?? '')
}

function handleNewFolder() {
    contextMenuStore.setMenuPosition({ x: 16, y: 144 })
    contextMenuStore.setTarget(locationStore.activeFolderID)
    contextMenuStore.setMenuMode('newName')
    contextMenuStore.setMenuOpen(true)
}

async function goHome() {
    if (locationStore.isInSettings && locationStore.returnTo) {
        const returnTo = locationStore.returnTo
        locationStore.returnTo = null

        return navigateTo(returnTo)
    }

    return WeblensFile.Home().GoTo()
}

async function goToSettings() {
    if (locationStore.isInSettings) {
        goHome()
        return
    }

    locationStore.returnTo = route.fullPath
    navigateTo('/settings/account')
}

async function goToLogin() {
    navigateTo('/login?returnTo=' + encodeURIComponent(route.fullPath))
}
</script>
