<template>
    <div
        :class="{
            'flex h-max w-full items-center border-t py-1 pl-1': true,
            gone: crumbFiles.length === 0,
        }"
    >
        <div
            v-for="(file, index) in crumbFiles"
            :key="index"
            :class="{
                'text-text-secondary inline-flex items-center': true,
                'group cursor-pointer': index !== crumbFiles.length - 1 && userStore.loggedIn,
            }"
            @click="
                () => {
                    if (index === crumbFiles.length - 1 || !userStore.loggedIn) {
                        return
                    }
                    file.GoTo()
                }
            "
            @mouseup="
                () => {
                    if (filesStore.dragging) moveFiles(file)
                }
            "
        >
            <IconChevronRight
                v-if="index > 0"
                size="12"
                :class="{ 'mx-1 text-inherit': true }"
            />

            <FileIcon
                :file="file"
                :class="{
                    'h-8 p-1 transition-[height,outline,background-color]': true,
                    'group-hover:bg-card-background-hover group-hover:text-text-primary rounded':
                        index !== crumbFiles.length - 1,
                    'pointer-events-none outline-none': index === crumbFiles.length - 1,
                    'hover:outline-theme-primary hover:bg-theme-primary/25 h-16 !text-lg outline outline-dashed hover:outline-2':
                        filesStore.dragging,
                }"
                with-name
            />
        </div>
    </div>
</template>
<script setup lang="ts">
import { IconChevronRight } from '@tabler/icons-vue'
import useFilesStore from '~/stores/files'
import WeblensFile from '~/types/weblensFile'
import FileIcon from '../atom/FileIcon.vue'
import { moveFiles } from '~/api/FileBrowserApi'
import useLocationStore from '~/stores/location'

const filesStore = useFilesStore()
const locationStore = useLocationStore()
const userStore = useUserStore()

const crumbFiles = computed(() => {
    let files: WeblensFile[] = []

    if (filesStore.parents && filesStore.parents.length !== 0) {
        files = [...filesStore.parents]
    }

    if (filesStore.activeFile) {
        files.push(filesStore.activeFile)
    }

    if (locationStore.isInShare) {
        files.unshift(WeblensFile.ShareRoot())
    }

    files.sort((a, b) => {
        return a.portablePath.length - b.portablePath.length
    })

    return files
})
</script>
