<template>
    <h3 v-if="file">{{ file?.GetFilename() }}</h3>
    <WeblensButton
        v-else-if="mediaId"
        label="Show Source File"
        @click.stop="goToFileByMediaID"
    />
</template>

<script setup lang="ts">
import useFilesStore from '~/stores/files'
import WeblensButton from '../atom/WeblensButton.vue'
import { useWeblensAPI } from '~/api/AllApi'
import WeblensFile from '~/types/weblensFile'

const fileStore = useFilesStore()
const presentationStore = usePresentationStore()
const mediaStore = useMediaStore()

const props = defineProps<{
    fileId: string
    mediaId: string
}>()

const file = computed(() => {
    return fileStore.getFileByID(props.fileId)
})

async function goToFileByMediaID() {
    const fileIDs = mediaStore.mediaMap.get(props.mediaId)?.fileIDs
    if (fileIDs?.length) {
        const fileID = fileIDs[0]
        const fileInfo = (await useWeblensAPI().FilesAPI.getFile(fileID)).data
        const file = new WeblensFile(fileInfo)
        if (file) {
            presentationStore.clearPresentation()
            console.log('Going to file', file)
            file.GoTo()
        }
    }

    console.error('No file found for media ID:', props.mediaId)
}
</script>
