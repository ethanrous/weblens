<template>
    <div :class="{ 'flex h-screen w-screen justify-center': true }">
        <div :class="{ 'flex h-full w-[20%] flex-col gap-2 border p-4': true }">
            <div :class="{ 'mb-2 border-b pb-2': true }">
                <h3 :class="{ 'leading-5': true }">{{ userStore.user.fullName }}</h3>
                <h5 :class="{ 'text-text-secondary': true }">{{ userStore.user.username }}</h5>
            </div>

            <WeblensButton
                label="Account"
                fill-width
                @click="toSettingsPage('account')"
            >
                <IconUser />
            </WeblensButton>

            <WeblensButton
                label="Appearance"
                fill-width
                @click="toSettingsPage('appearance')"
            >
                <IconBrush />
            </WeblensButton>

            <Divider
                label="Admin"
                label-justify="left"
            />

            <WeblensButton
                label="Users"
                fill-width
                @click="toSettingsPage('users')"
            >
                <IconUsers />
            </WeblensButton>

            <WeblensButton
                label="Developer"
                fill-width
                @click="toSettingsPage('dev')"
            >
                <IconCode />
            </WeblensButton>

            <WeblensButton
                label="Log Out"
                fill-width
                :class="{ 'mt-auto': true }"
                flavor="danger"
                @click="userStore.logout"
            >
                <IconLogout />
            </WeblensButton>
        </div>

        <div :class="{ 'h-full w-full p-4': true }">
            <NuxtPage />
        </div>
    </div>
</template>

<script setup lang="ts">
import { IconBrush, IconCode, IconLogout, IconUser, IconUsers } from '@tabler/icons-vue'
import WeblensButton from '~/components/atom/WeblensButton.vue'
import Divider from '~/components/atom/Divider.vue'

const userStore = useUserStore()
const route = useRoute()

function toSettingsPage(pageName: string) {
    navigateTo({ path: `/settings/${pageName}` })
}

watch(
    route,
    () => {
        const currentPath = route.path

        if (currentPath === '/settings' || currentPath === '/settings/') {
            navigateTo({ path: '/settings/account', replace: true })
        }
    },
    { immediate: true, deep: true },
)
</script>
