<template>
    <div
        v-if="file"
        :class="{ 'fileIcon inline-flex items-center': true }"
    >
        <IconHome
            v-if="file.IsHome()"
            :stroke-width="1.5"
        />
        <IconUser
            v-else-if="file.IsShareRoot()"
            :stroke-width="1.5"
        />
        <IconTrash
            v-else-if="file.IsTrash()"
            :stroke-width="1.5"
        />
        <IconFolder
            v-else-if="file.IsFolder()"
            :stroke-width="1.5"
        />
        <IconFile
            v-else
            :stroke-width="1.5"
        />

        <h5
            v-if="withName"
            :class="{ 'ml-0.5 truncate text-nowrap select-none': true }"
        >
            {{ file.GetFilename() }}
        </h5>
    </div>
</template>

<script setup lang="ts">
import { IconFile, IconFolder, IconHome, IconTrash, IconUser } from '@tabler/icons-vue'
import type WeblensFile from '~/types/weblensFile'

defineProps<{
    file?: WeblensFile
    withName?: boolean
}>()
</script>
