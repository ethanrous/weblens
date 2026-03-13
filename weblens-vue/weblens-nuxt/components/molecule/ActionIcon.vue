<template>
    <IconPlus
        v-if="actionType === 'fileCreate'"
        :size="size"
    />
    <IconArrowRight
        v-else-if="actionType === 'fileMove'"
        :size="size"
    />
    <IconTrash
        v-else-if="actionType === 'fileDelete'"
        :size="size"
    />
    <IconRestore
        v-else-if="actionType === 'fileRestore'"
        :size="size"
    />
    <IconFile
        v-else
        :size="size"
    />
</template>

<script setup lang="ts">
import { IconArrowRight, IconFile, IconPlus, IconRestore, IconTrash } from '@tabler/icons-vue'

withDefaults(
    defineProps<{
        actionType: string
        size?: number
    }>(),
    {
        size: 16,
    },
)
</script>
