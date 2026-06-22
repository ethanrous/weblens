<template>
    <div :class="{ 'inline-flex w-full max-w-full truncate text-nowrap': true }">
        <div
            v-for="(part, i) of parts"
            :key="i"
            :class="{
                'inline-flex max-w-max min-w-5 grow items-center': true,
                'min-w-8': i === 0,

                // Make sure the last part of the path doesn't truncate
                'min-w-max': i === parts.length - 1,
            }"
        >
            <!-- <template v-if="i !== parts.length - 1 && width < 24"> . </template> -->

            <template v-if="i === 0 && part === userStore.user.username">
                <IconHome size="1.2em" />
            </template>

            <template v-else-if="part === '.user_trash'">
                <IconTrash size="1.2em" />
            </template>

            <template v-else-if="i === 0">
                <IconFolder size="1.2em" />
                <span :class="{ 'absolute mb-1 ml-2': true }">.</span>

                <!-- <span :class="{'ml-1 min-w-0 truncate': true}">{{ part }}</span> -->
            </template>

            <span
                v-else
                :class="{
                    'min-w-0 truncate': i !== parts.length - 1,
                    'min-w-max': i === parts.length - 1,
                }"
            >
                {{ part }}
            </span>

            <IconSlash
                v-if="i < parts.length - 1"
                size="14"
            />
        </div>
    </div>
</template>

<script setup lang="ts">
import { IconFolder, IconHome, IconSlash, IconTrash } from '@tabler/icons-vue'
import type { PortablePath } from '~/types/portablePath'

const userStore = useUserStore()

const props = defineProps<{
    path: PortablePath
    omitFirst?: boolean
    omitLast?: boolean
}>()

const parts = computed(() => {
    const parts = props.path.parts
    if (props.omitFirst) {
        parts.splice(0, 1) // Remove the first part if omitFirst is true
    }

    if (props.omitLast) {
        parts.pop() // Remove the last part if omitLast is true
    }

    return parts
})
</script>
