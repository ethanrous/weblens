<template>
    <div
        :class="{
            'file-action-card group flex w-full items-center gap-1 rounded border px-1.5': true,
            'py-1': compact,
            'py-1.5': !compact,
            'border-theme-primary bg-background-secondary': locationStore.viewTimestamp === action.timestamp,
            'hover:bg-background-secondary cursor-pointer': locationStore.viewTimestamp !== action.timestamp,
        }"
        @click.stop="locationStore.setViewTimestamp(action.timestamp)"
    >
        <span
            v-if="action.actionType === FileAction.FileMove"
            :class="{
                'inline-flex min-w-0 flex-1 items-center gap-1 truncate': true,
                'text-text-tertiary group-hover:text-text-primary': afterRewindTimestamp,
            }"
        >
            <span>{{ originPath?.filename }}</span>

            <IconArrowRight size="14" />

            <IconTrash
                v-if="wasTrashed"
                size="14"
            />

            <a
                v-if="!wasTrashed"
                :class="{
                    'inline-flex items-center text-inherit no-underline': true,
                    'hover:text-blue-600 hover:underline': shouldShowLink,
                }"
                :href="shouldShowLink ? actionFile.FileURL({ forcePresent: true }) : undefined"
                @click.stop
            >
                <IconFolder
                    size="14"
                    :class="{ 'mr-1': true }"
                />
                <FilePath
                    :path="destPath"
                    omit-last
                />
            </a>
        </span>

        <span
            v-else-if="action.actionType === FileAction.FileCreate"
            :class="{
                'inline-flex min-w-0 flex-1 items-center gap-0.5 truncate': true,
                'text-text-tertiary group-hover:text-text-primary': afterRewindTimestamp,
            }"
        >
            <IconPlus :size="14" />
            <FileIcon :file="actionFile" />
            {{ destPath?.filename }}
            <span
                v-if="fileStore.activeFile && destPath.equals(fileStore.activeFile?.GetFilepath())"
                :class="{ 'text-text-secondary bg-card-background-primary ml-2 rounded p-1 text-xs': true }"
            >
                This Folder
            </span>
        </span>

        <span
            v-else-if="action.actionType === FileAction.FileDelete"
            :class="{
                'inline-flex min-w-0 flex-1 items-center gap-0.5 truncate': true,
                'text-text-tertiary group-hover:text-text-primary': afterRewindTimestamp,
            }"
        >
            <IconTrash :size="14" />
            <FileIcon :file="actionFile" />
            {{ originPath?.filename }}
        </span>

        <span
            v-else-if="action.actionType === FileAction.FileRestore"
            :class="{
                'inline-flex min-w-0 flex-1 items-center gap-0.5 truncate': true,
                'text-text-tertiary group-hover:text-text-primary': afterRewindTimestamp,
            }"
        >
            <IconRestore :size="14" />
            <FileIcon :file="actionFile" />
            {{ originPath?.filename }}
        </span>

        <span
            v-else
            :class="{ 'min-w-0 flex-1 truncate': true }"
        >
            {{ action.actionType }}
        </span>

        <span
            :class="{
                'text-text-secondary min-w-0 truncate text-xs': true,
                'text-text-tertiary group-hover:text-text-secondary': afterRewindTimestamp,
            }"
        >
            {{ actionName }}
        </span>

        <span
            v-if="!compact"
            :class="{
                'text-text-secondary ml-auto shrink-0 text-xs': true,
                'text-text-tertiary group-hover:text-text-secondary': afterRewindTimestamp,
            }"
        >
            {{ relTime }}
        </span>
    </div>
</template>

<script setup lang="ts">
import type { FileActionInfo } from '@ethanrous/weblens-api'
import { IconArrowRight, IconFolder, IconPlus, IconRestore, IconTrash } from '@tabler/icons-vue'
import { FileAction } from '~/types/fileHistory'
import { PortablePath } from '~/types/portablePath'
import FilePath from '../atom/FilePath.vue'
import useLocationStore from '~/stores/location'
import { relativeTimeAgo } from '~/util/relativeTime'
import FileIcon from '../atom/FileIcon.vue'
import WeblensFile from '~/types/weblensFile'
import { friendlyActionName } from '~/util/history'
import useFilesStore from '~/stores/files'

const locationStore = useLocationStore()
const fileStore = useFilesStore()

const props = withDefaults(
    defineProps<{
        action: FileActionInfo
        compact?: boolean
    }>(),
    {
        compact: false,
    },
)

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
    if (props.action.actionType === FileAction.FileMove && props.action.destinationPath) {
        return new PortablePath(props.action.destinationPath)
    } else if (props.action.actionType === FileAction.FileCreate && props.action.filepath) {
        return new PortablePath(props.action.filepath)
    }

    return PortablePath.empty()
})

const actionFile = computed(() => {
    return WeblensFile.FromAction(props.action)
})

const shouldShowLink = computed(() => {
    if (props.action.actionType !== FileAction.FileMove) {
        return false
    }

    if (!props.action.liveParentID) {
        return false
    }

    if (props.action.liveParentID === fileStore.activeFile?.ID()) {
        return false
    }

    return true
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
