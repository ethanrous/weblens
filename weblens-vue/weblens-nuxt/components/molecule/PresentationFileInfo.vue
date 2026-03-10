<template>
    <div
        v-if="file"
        class="flex w-full flex-col gap-2 p-1"
    >
        <div :class="{ 'flex items-center gap-1': true }">
            <IconFolder v-if="file.isDir" />
            <IconFile v-else />
            <h3 class="truncate">{{ file.GetFilename() }}</h3>
        </div>

        <div class="flex flex-col">
            <span class="text-text-secondary mb-1 text-xs font-semibold uppercase"> File Details </span>

            <InfoRow
                v-if="mediaTypeName"
                label="Type"
                :value="mediaTypeName"
            >
                <template #icon>
                    <IconFile size="18" />
                </template>
            </InfoRow>

            <InfoRow
                label="Size"
                :value="file.FormatSize()"
            >
                <template #icon>
                    <IconDatabase size="18" />
                </template>
            </InfoRow>

            <InfoRow
                label="Modified"
                :value="file.FormatModified()"
            >
                <template #icon>
                    <IconCalendar size="18" />
                </template>
            </InfoRow>

            <InfoRow
                v-if="file.owner"
                label="Owner"
                :value="file.owner"
            >
                <template #icon>
                    <IconUser size="18" />
                </template>
            </InfoRow>
        </div>

        <div class="flex flex-col">
            <span class="text-text-secondary mb-1 text-xs font-semibold uppercase">Tags</span>
            <div class="flex flex-wrap gap-1">
                <TagPill
                    v-for="tag in fileTags"
                    :key="tag.id"
                    :tag="tag"
                    removable
                    clickable
                    class="group"
                    @click="removeTag(tag.id!)"
                />
                <span
                    class="clickable text-text-tertiary inline-flex items-center gap-0.5 rounded-full text-xs"
                    @click="showTagSelector = !showTagSelector"
                >
                    <IconPlus size="12" />
                    Add
                </span>
            </div>
            <TagSelector
                v-if="showTagSelector && file"
                :file-i-ds="[file.ID()]"
                class="mt-1"
            />
        </div>

        <div :class="{ 'flex gap-4': true }">
            <WeblensButton
                label="Show in Files"
                type="outline"
                :disabled="file.ID() === fileStore.activeFile?.ID()"
                @click.stop="goToFile"
            >
                <IconSearch size="16" />
            </WeblensButton>

            <WeblensButton
                :label="downloadTaskID ? 'Zipping...' : 'Download'"
                type="outline"
                @click.stop="download()"
            >
                <IconDownload size="16" />
            </WeblensButton>
        </div>

        <ProgressSquare
            v-if="zipProgress"
            :class="{ 'h-4 w-full': true }"
            :progress="zipProgress"
        />
    </div>

    <div v-else-if="mediaId">
        <span class="text-text-secondary mb-2 text-xs font-semibold uppercase"> File Details </span>

        <WeblensButton
            label="Show Source File"
            type="outline"
            @click.stop="goToFileByMediaID"
        >
            <IconSearch size="16" />
        </WeblensButton>
    </div>
</template>

<script setup lang="ts">
import {
    IconCalendar,
    IconDatabase,
    IconDownload,
    IconFile,
    IconFolder,
    IconPlus,
    IconSearch,
    IconUser,
} from '@tabler/icons-vue'
import useFilesStore from '~/stores/files'
import useTagsStore from '~/stores/tags'
import WeblensButton from '../atom/WeblensButton.vue'
import InfoRow from '../atom/InfoRow.vue'
import TagSelector from './TagSelector.vue'
import { useWeblensAPI } from '~/api/AllApi'
import WeblensFile from '~/types/weblensFile'
import { handleDownload } from '~/api/FileBrowserApi'
import ProgressSquare from '../atom/ProgressSquare.vue'
import TagPill from '../atom/TagPill.vue'

const fileStore = useFilesStore()
const presentationStore = usePresentationStore()
const mediaStore = useMediaStore()
const tasksStore = useTasksStore()
const tagsStore = useTagsStore()

const showTagSelector = ref(false)

const fileTags = computed(() => {
    if (!file.value) return []
    return tagsStore.getTagsByFileID(file.value.ID())
})

async function removeTag(tagID: string) {
    if (!file.value) return
    await tagsStore.removeFilesFromTag(tagID, [file.value.ID()])
}

const props = defineProps<{
    fileId: string
    mediaId: string
}>()

const file = computed(() => {
    return fileStore.getFileByID(props.fileId)
})

const zipProgress = computed(() => {
    if (!downloadTaskID.value) return null

    const task = tasksStore.tasks?.get(downloadTaskID.value)
    if (!task) return null

    return task.percentComplete
})

const mediaTypeName = computed(() => {
    const media = mediaStore.mediaMap.get(props.mediaId)
    if (media) {
        return media.GetMediaType()?.FriendlyName
    }
    return undefined
})

function goToFile() {
    if (file.value) {
        presentationStore.clearPresentation()
        file.value.GoTo()
    }
}

async function goToFileByMediaID() {
    const fileIDs = mediaStore.mediaMap.get(props.mediaId)?.fileIDs
    if (!fileIDs?.length) {
        console.error('No file found for media ID:', props.mediaId)
        return
    }

    const fileID = fileIDs[0]
    const fileInfo = (await useWeblensAPI().FilesAPI.getFile(fileID)).data
    const f = new WeblensFile(fileInfo)
    presentationStore.clearPresentation()
    f.GoTo()
}

const downloadTaskID = ref<string | null>(null)

async function download() {
    if (!file.value) return

    const downloadInfo = await handleDownload([file.value])
    if (downloadInfo) {
        if (downloadInfo.zipTaskID) {
            downloadTaskID.value = downloadInfo.zipTaskID
        }

        await downloadInfo.downloadPromise
        downloadTaskID.value = null
    }
}
</script>
