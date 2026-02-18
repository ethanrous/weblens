<template>
    <div
        :class="{
            'file-action-card hover:bg-background-secondary group flex h-20 w-full cursor-pointer flex-col rounded border p-1.5': true,
            'border-theme-primary': locationStore.viewTimestamp === action.timestamp,
        }"
        @click="locationStore.setViewTimestamp(action.timestamp)"
    >
        <span
            v-if="action.actionType === FileAction.FileMove"
            :class="{ 'inline-flex items-center gap-1 truncate': true }"
        >
            {{ originPath?.filename }}

            <IconArrowRight size="16" />

            <IconTrash v-if="wasTrashed" />

            <IconFolder v-if="!wasTrashed" />
            <FilePath
                v-if="!wasTrashed"
                :path="destPath"
                omit-last
            />
        </span>

        <span
            v-else-if="action.actionType === FileAction.FileCreate"
            :class="{
                'inline-flex items-center gap-0.5': true,
                'text-text-tertiary group-hover:text-text-primary': afterRewindTimestamp,
            }"
        >
            <IconPlus :size="14" />
            <FileIcon :file="WeblensFile.FromAction(action)" />
            {{ originPath?.filename }}
        </span>

        <span v-else>
            {{ action.actionType }}
        </span>

        <div
            :class="{
                'text-text-secondary mt-auto flex': true,
                'text-text-tertiary group-hover:text-text-secondary': afterRewindTimestamp,
            }"
        >
            <span> File {{ actionName }} </span>
            <span
                :class="{ 'ml-auto': true }"
                :title="new Date(action.timestamp).toLocaleString()"
            >
                {{ relTime }}
            </span>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { FileActionInfo } from '@ethanrous/weblens-api'
import { IconArrowRight, IconFolder, IconPlus, IconTrash } from '@tabler/icons-vue'
import { FileAction } from '~/types/fileHistory'
import { PortablePath } from '~/types/portablePath'
import { friendlyActionName } from '~/util/history'
import FilePath from '../atom/FilePath.vue'
import useLocationStore from '~/stores/location'
import { relativeTimeAgo } from '~/util/relativeTime'
import FileIcon from '../atom/FileIcon.vue'
import WeblensFile from '~/types/weblensFile'

const locationStore = useLocationStore()

const props = defineProps<{
    action: FileActionInfo
}>()

const originPath = computed(() => {
    const originPath = props.action.originPath
    if (originPath) {
        return new PortablePath(originPath)
    }

    const filepath = props.action.filepath
    if (filepath) {
        return new PortablePath(filepath)
    }

    return PortablePath.empty()
})

const destPath = computed(() => {
    if (props.action.actionType === 'fileMove' && props.action.destinationPath) {
        return new PortablePath(props.action.destinationPath)
    }

    return PortablePath.empty()
})

const wasTrashed = computed(() => {
    if (!destPath.value) return false
    return destPath.value.isInTrash()
})

const actionName = computed(() => {
    if (wasTrashed.value) {
        return 'Trashed'
    }

    return friendlyActionName(props.action.actionType)
})

const relTime = computed(() => {
    return relativeTimeAgo(props.action.timestamp)
})

const afterRewindTimestamp = computed(() => {
    return locationStore.viewTimestamp > 0 && props.action.timestamp > locationStore.viewTimestamp
})
</script>
