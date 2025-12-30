<template>
    <div class="flex h-screen w-screen items-center justify-center">
        <AtomLoader />
        <NuxtPage />
    </div>
</template>

<script setup lang="ts">
import WeblensFile from '~/types/weblensFile'

const userStore = useUserStore()

watchEffect(() => {
    const loggedIn = userStore.user.isLoggedIn

    if (loggedIn.isSet() && loggedIn.get()) {
        WeblensFile.Home().GoTo(true)
    } else if (loggedIn.isSet() && !loggedIn.get()) {
        console.debug('User is not logged at root, redirecting to login page')

        navigateTo({
            path: '/login',
        })
    }
})
</script>
