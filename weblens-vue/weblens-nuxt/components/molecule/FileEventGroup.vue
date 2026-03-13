<template>
    <div
        v-if="group.actions.length === 1"
        :key="group.eventID"
    >
        <FileAction :action="group.actions[0]" />
    </div>

    <div
        v-else
        :key="group.eventID + '-group'"
        :class="{
            'rounded border': true,
            'border-theme-primary': locationStore.viewTimestamp === group.timestamp,
        }"
    >
        <div
            :class="{
                'hover:bg-background-secondary group flex w-full cursor-pointer items-center gap-1 rounded px-1.5 py-1.5': true,
                'rounded-b-none': expanded,
            }"
            @click.stop="locationStore.setViewTimestamp(group.timestamp)"
        >
            <WeblensButton
                type="light"
                :square-size="22"
                @click.stop="expanded = !expanded"
            >
                <IconChevronRight
                    :size="14"
                    :class="{
                        'transition-transform': true,
                        'rotate-90': expanded,
                    }"
                />
            </WeblensButton>

            <span
                :class="{
                    'inline-flex min-w-0 items-center gap-1': true,
                    'text-text-tertiary group-hover:text-text-primary': afterRewindTimestamp,
                }"
            >
                <ActionIcon
                    :action-type="group.actionType"
                    :size="14"
                />
                <span>
                    {{ friendlyActionName(group.actionType) }}
                </span>
                <span
                    :class="{
                        'bg-background-tertiary text-text-secondary inline-flex items-center rounded px-1.5 text-xs': true,
                        'text-text-tertiary group-hover:text-text-secondary': afterRewindTimestamp,
                    }"
                >
                    {{ group.actions.length }} files
                </span>
            </span>

            <span
                :class="{
                    'text-text-secondary ml-auto shrink-0 text-xs': true,
                    'text-text-tertiary group-hover:text-text-secondary': afterRewindTimestamp,
                }"
                :title="new Date(group.timestamp).toLocaleString()"
            >
                {{ relTime }}
            </span>
        </div>

        <div
            v-if="expanded"
            :class="{ 'flex flex-col gap-0.5 border-t px-1 py-1': true }"
        >
            <FileAction
                v-for="action in visibleActions"
                :key="action.fileID"
                :action="action"
                compact
                :class="{ 'border-none': true }"
            />
            <span
                v-if="group.actions.length > maxVisible"
                :class="{ 'text-text-tertiary px-2 text-xs': true }"
            >
                and {{ group.actions.length - maxVisible }} more
            </span>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { FileActionInfo } from '@ethanrous/weblens-api'
import { IconChevronRight } from '@tabler/icons-vue'
import WeblensButton from '../atom/WeblensButton.vue'
import FileAction from './FileAction.vue'
import ActionIcon from './ActionIcon.vue'
import useLocationStore from '~/stores/location'
import { relativeTimeAgo } from '~/util/relativeTime'
import { friendlyActionName } from '~/util/history'

export interface EventGroup {
    eventID: string
    actions: FileActionInfo[]
    timestamp: number
    actionType: string
}

const locationStore = useLocationStore()

const props = defineProps<{
    group: EventGroup
}>()

const expanded = ref(false)
const maxVisible = 20

const visibleActions = computed(() => {
    return props.group.actions.slice(0, maxVisible)
})

const relTime = computed(() => {
    return relativeTimeAgo(props.group.timestamp)
})

const afterRewindTimestamp = computed(() => {
    return locationStore.viewTimestamp > 0 && props.group.timestamp > locationStore.viewTimestamp
})
</script>
