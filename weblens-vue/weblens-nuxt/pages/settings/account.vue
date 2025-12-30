<template>
    <div>
        <h3 :class="{ 'mb-4 border-b pb-2': true }">Account</h3>

        <h4>Change Display Name</h4>
        <span :class="{ 'text-text-secondary mb-2': true }">The name others can find you by. e.g. John Smith</span>
        <div :class="{ 'flex items-center': true }">
            <WeblensInput
                v-model:value="fullName"
                merge="row"
                :class="{ 'inline-flex': true }"
                placeholder="Jonh Smith"
            />
            <WeblensButton
                label="Update Name"
                :disabled="!fullName || userStore.user.fullName === fullName"
                :class="{ 'h-12': true }"
                merge="row"
                @click.stop="updateFullName"
            />
        </div>

        <h4 :class="{ 'mt-4 mb-2': true }">Change Password</h4>
        <div :class="{ 'flex w-max flex-col justify-center gap-2 rounded border p-4': true }">
            <WeblensInput
                id="oldPassword"
                v-model:value="oldPassword"
                :class="{ 'inline-flex h-full': true }"
                placeholder="Old Password"
            />

            <WeblensInput
                id="newPassword"
                v-model:value="newPassword"
                :class="{ 'inline-flex': true }"
                placeholder="New Password"
                @submit="updatePasswordButton?.click"
            />

            <WeblensButton
                ref="updatePasswordButton"
                label="Update Password"
                :disabled="!canChnagePassword"
                fill-width
                :error-text="handleChangePasswordFail"
                @click="updatePassword"
            />
        </div>
    </div>
</template>

<script setup lang="ts">
import type { UserInfo } from '@ethanrous/weblens-api'
import type { AxiosError } from 'axios'
import { useWeblensAPI } from '~/api/AllApi'
import WeblensButton from '~/components/atom/WeblensButton.vue'
import WeblensInput from '~/components/atom/WeblensInput.vue'

const userStore = useUserStore()

const fullName = ref<string>()
const oldPassword = ref<string>()
const newPassword = ref<string>()
const updatePasswordButton = ref<typeof WeblensButton>()

async function updateFullName() {
    if (!fullName.value || fullName.value === userStore.user.fullName) {
        return
    }

    const res = await useWeblensAPI().UsersAPI.changeDisplayName(userStore.user.username, fullName.value)
    userStore.setUser(res.data as UserInfo, true)
}

const canChnagePassword = computed(() => {
    return !!oldPassword.value && !!newPassword.value && oldPassword.value !== newPassword.value
})

async function updatePassword() {
    if (!canChnagePassword.value) {
        return
    }

    await useWeblensAPI().UsersAPI.updateUserPassword(userStore.user.username, {
        oldPassword: oldPassword.value,
        newPassword: newPassword.value!,
    })

    oldPassword.value = ''
    newPassword.value = ''
}

function handleChangePasswordFail(error: AxiosError): string {
    if (error.status === 401) {
        return 'Incorrect old password'
    } else if (error.status === 400) {
        return 'Invalid new password'
    }

    return 'Unexpected error'
}

watchEffect(() => {
    if (userStore.user.fullName) {
        fullName.value = userStore.user.fullName
    }
})
</script>
