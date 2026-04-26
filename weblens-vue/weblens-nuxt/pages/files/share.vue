<template>
    <FileBrowser v-if="locationStore.inShareRoot" />
    <NuxtPage v-else />
</template>

<script setup lang="ts">
import FileBrowser from '~/components/organism/FileBrowser.vue'
import useLocationStore from '~/stores/location'
const locationStore = useLocationStore()
</script>
