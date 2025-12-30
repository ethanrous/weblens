<template>
    <div :class="{ 'absolute bottom-4 z-[99] flex w-full justify-center': true }">
        <h4
            v-if="locationStore.isViewingPast"
            :class="{
                'bg-theme-secondary/50 border-theme-secondary inline-flex items-center gap-2 rounded border p-2 text-lg': true,
            }"
            :title="`Viewing at ${new Date(locationStore.viewTimestamp).toLocaleString()}`"
        >
            <IconClock />
            Rewound to {{ relativeTimeAgo(locationStore.viewTimestamp) }}
            <IconX
                size="20"
                :class="{ 'text-text-secondary hover:text-text-primary cursor-pointer': true }"
                @click="locationStore.setViewTimestamp(0)"
            />
        </h4>
    </div>
</template>

<script setup lang="ts">
import { IconClock, IconX } from '@tabler/icons-vue'
import useLocationStore from '~/stores/location'
import { relativeTimeAgo } from '~/util/relativeTime'

const locationStore = useLocationStore()
</script>
