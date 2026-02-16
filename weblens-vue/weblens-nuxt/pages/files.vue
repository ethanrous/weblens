<template>
    <div :class="{ 'page-root relative': true }">
        <Presentation
            v-if="presentationStore.presentationFileID || presentationStore.presentationMediaID"
            :next="presentationNextFn"
            :previous="presentationPrevFn"
        >
            <template #media="props">
                <PresentationMediaContent
                    v-if="locationStore.isInTimeline || presentingFile?.contentID"
                    :media-id="mediaID"
                    :presentation-size="props.presentationSize"
                />

                <div
                    v-else-if="presentingFile?.IsFolder()"
                    :class="{
                        'flex h-full w-full min-w-max flex-col items-center justify-center': true,
                    }"
                >
                    <IconFolder
                        size="18rem"
                        stroke="1"
                    />

                    <div :class="{ 'mb-6 flex w-[20rem] items-center': true }">
                        <IconDatabase />
                        <h4 :class="{ 'mr-auto ml-2': true }">{{ presentingFile.FormatSize() }}</h4>

                        <IconCalendar />
                        <h4 :class="{ 'ml-2': true }">{{ presentingFile.FormatModified() }}</h4>
                    </div>

                    <h1>{{ presentingFile.GetFilename() }}</h1>
                </div>
            </template>

            <template #fileInfo>
                <PresentationFileInfo
                    :file-id="presentationStore.presentationFileID"
                    :media-id="mediaID"
                />
            </template>

            <template
                v-if="mediaID"
                #mediaInfo
            >
                <PresentationMediaInfo :media-id="mediaID" />
            </template>
        </Presentation>
        <UploadProgress />

        <div :class="{ 'flex h-full w-full max-w-full min-w-0 flex-col': true }">
            <FileHeader />
            <NuxtPage />
            <PathCrumbs />
        </div>

        <WebsocketStatus
            :class="{ 'absolute right-4 bottom-4 hidden sm:block': true }"
            :ws-status="wsStore.status"
        />
    </div>
</template>

<script setup lang="ts">
import FileHeader from '~/components/organism/FileHeader.vue'
import PathCrumbs from '~/components/organism/PathCrumbs.vue'
import Presentation from '~/components/organism/Presentation.vue'
import PresentationFileInfo from '~/components/molecule/PresentationFileInfo.vue'
import PresentationMediaContent from '~/components/molecule/PresentationMediaContent.vue'
import UploadProgress from '~/components/organism/UploadProgress.vue'
import WebsocketStatus from '~/components/atom/WebsocketStatus.vue'
import useFilesStore from '~/stores/files'
import useLocationStore from '~/stores/location'
import useWebsocketStore from '~/stores/websocket'
import { IconCalendar, IconDatabase, IconFolder } from '@tabler/icons-vue'
import PresentationMediaInfo from '~/components/molecule/PresentationMediaInfo.vue'

const wsStore = useWebsocketStore()
const locationStore = useLocationStore()
const presentationStore = usePresentationStore()
const filesStore = useFilesStore()
const mediaStore = useMediaStore()

const presentingFile = computed(() => {
    return filesStore.children?.find((f) => f.id === presentationStore.presentationFileID)
})

const mediaID = computed(() => {
    return locationStore.isInTimeline ? presentationStore.presentationMediaID : (presentingFile.value?.contentID ?? '')
})

const presentationNextFn = computed(() => {
    return () => {
        if (locationStore.isInTimeline) {
            const nextID = mediaStore.getNextMediaID(presentationStore.presentationMediaID)
            if (nextID) {
                presentationStore.setPresentationMediaID(nextID)
            }
        } else {
            const nextID = filesStore.getNextFileID(presentationStore.presentationFileID)
            if (nextID) {
                presentationStore.setPresentationFileID(nextID)
            }
        }
    }
})

const presentationPrevFn = computed(() => {
    return () => {
        if (locationStore.isInTimeline) {
            const prevID = mediaStore.getPreviousMediaID(presentationStore.presentationMediaID)
            if (prevID) {
                presentationStore.setPresentationMediaID(prevID)
            }
        } else {
            const prevID = filesStore.getPreviousFileID(presentationStore.presentationFileID)
            if (prevID) {
                presentationStore.setPresentationFileID(prevID)
            }
        }
    }
})
</script>
