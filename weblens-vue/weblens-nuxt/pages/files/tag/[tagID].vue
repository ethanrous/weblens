<template>
    <FileBrowser />
</template>

<script setup lang="ts">
import useTagsStore from '~/stores/tags'
import FileBrowser from '~/components/organism/FileBrowser.vue'

const tagsStore = useTagsStore()

onMounted(() => {
    if (tagsStore.tagsList.length === 0) {
        tagsStore.fetchTags()
    }
})
</script>
