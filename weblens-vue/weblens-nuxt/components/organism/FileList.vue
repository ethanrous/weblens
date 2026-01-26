<template>
    <div>
        <span
            v-for="file in files"
            :key="file.id"
            :class="{ block: true }"
        >
            {{ file.GetFilename() }}
        </span>
    </div>
</template>

<script setup lang="ts">
import type WeblensFile from '~/types/weblensFile'

defineProps<{
    files: WeblensFile[]
}>()
</script>
