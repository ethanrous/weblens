<template>
    <div :class="{ 'flex h-screen w-screen flex-col items-center justify-center': true }">
        <div
            v-if="error.statusCode === 404"
            class="mb-4 text-center"
        >
            <h1 class="text-4xl font-bold">404</h1>
            <p class="text-lg">Page not found</p>
        </div>

        <div
            v-else
            :class="{ 'flex flex-row': true }"
        >
            <IconFileSad
                size="200"
                :stroke="'2'"
            />
            <div :class="{ 'relative ml-12 flex w-max flex-col overflow-y-auto': true }">
                <h2 :class="{ 'block text-nowrap': true }">Well... thats not good</h2>
                <h6 :class="{ 'absolute mt-12 w-full': true }">
                    {{ error.statusCode }} - {{ error.cause }} {{ error }}
                </h6>
            </div>
        </div>

        <WeblensButton
            label="Go Home"
            center-content
            @click="WeblensFile.Home().GoTo()"
        >
            <IconHome />
        </WeblensButton>
    </div>
</template>

<script setup lang="ts">
import type { NuxtError } from '#app'
import { IconFileSad, IconHome } from '@tabler/icons-vue'
import WeblensButton from '~/components/atom/WeblensButton.vue'
import WeblensFile from './types/weblensFile'

defineProps<{ error: NuxtError }>()
</script>
