<template>
    <div :class="{ 'page-root items-center p-8': true }">
        <div
            :class="{
                'bg-background-primary border-border-primary relative z-10 m-auto flex h-max max-h-screen min-w-0 flex-col items-center justify-center gap-2 rounded-2xl border p-8 shadow-2xl sm:justify-normal lg:my-0': true,
            }"
        >
            <div class="flex w-full justify-center text-center">
                <Logo :size="100" />
                <h1 class="xs:inline-block mt-auto hidden">EBLENS</h1>
            </div>
            <form
                id="login"
                action="#"
                class="mx-auto mt-8 flex w-96 max-w-full min-w-0 flex-col gap-3 px-4"
                @submit="
                    (e) => {
                        e.preventDefault()
                        e.stopPropagation()
                        doLogin()
                    }
                "
            >
                <WeblensInput
                    v-model:value="username"
                    placeholder="Username"
                    auto-focus
                    square-size="{44}"
                    auto-complete="username"
                />
                <WeblensInput
                    v-model:value="password"
                    placeholder="Password"
                    square-size="{44}"
                    password
                    auto-complete="current-password"
                />
                <span
                    v-if="formError"
                    class="text-center text-red-500"
                >
                    {{ formError }}
                </span>
                <div class="my-3">
                    <WeblensButton
                        :label="loading ? 'Signing in...' : 'Sign in'"
                        fill-width
                        :square-size="50"
                        :disabled="username === '' || password.length < 6 || loading"
                        center-content
                    />
                </div>
                <div class="border-color-border-primary flex items-center justify-center gap-2 border-t-[1px] p-2">
                    <span class="text-color-text-primary ml-auto">New Here?</span>
                    <a href="/signup">Request an Account</a>
                </div>
            </form>
        </div>
        <a
            href="https://github.com/ethanrous/weblens"
            :class="{ 'absolute right-0 bottom-0 m-4 flex flex-row bg-transparent': true }"
            target="_blank"
            rel="noreferrer"
        >
            <IconBrandGithub />
            GitHub
        </a>
    </div>
</template>

<script setup lang="ts">
import { IconBrandGithub } from '@tabler/icons-vue'
import { onKeyDown } from '@vueuse/core'
import type { AxiosError } from 'axios'
import { useWeblensAPI } from '~/api/AllApi'
import Logo from '~/components/atom/Logo.vue'
import WeblensButton from '~/components/atom/WeblensButton.vue'
import WeblensInput from '~/components/atom/WeblensInput.vue'
import User from '~/types/user'

const username = ref('')
const password = ref('')
const loading = ref(false)
const formError = ref<string | null>(null)

onKeyDown('Enter', doLogin)

async function doLogin() {
    formError.value = ''

    if (username.value === '' || password.value === '') {
        formError.value = 'username and password must not be empty'
        return
    }

    loading.value = true

    return useWeblensAPI()
        .UsersAPI.loginUser({ username: username.value, password: password.value })
        .then((res) => {
            // useFileBrowserStore.getState().reset()
            const user = new User(res.data)

            useUserStore().setUser(user, true)

            navigateTo({ path: '/files/home' })
        })
        .catch((err: AxiosError) => {
            loading.value = false
            if (err.status === 401) {
                formError.value = 'Invalid username or password'
            } else {
                formError.value = err.message || 'An error occurred'
            }
        })
}
</script>
