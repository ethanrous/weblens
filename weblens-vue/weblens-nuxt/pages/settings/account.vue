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
                :disabled="!canChangePassword"
                fill-width
                :error-text="(e) => handleChangePasswordFail(e as AxiosError)"
                @click="updatePassword"
            />
        </div>

        <h4 :class="{ 'mt-4 mb-2': true }">API Keys</h4>
        <div :class="{ 'flex max-w-96 min-w-max flex-col justify-center rounded border': true }">
            <div
                v-if="apiKeys && apiKeys.length > 0"
                :class="{ 'flex flex-col': true }"
            >
                <div
                    v-for="key in apiKeys"
                    :key="key.id"
                    :class="{ 'flex w-full items-center p-2 not-last:border-b': true }"
                >
                    <div :class="{ 'mr-4 flex flex-col': true }">
                        <span :class="{ 'mb-1 font-medium': true }">
                            {{ key.nickname }}
                        </span>
                        <CopyBox :text="key.token" />
                    </div>
                    <WeblensButton
                        flavor="danger"
                        :class="{ 'ml-auto max-h-8': true }"
                        :square-size="16"
                        @click="deleteAPIKey(key.id, key.nickname)"
                    >
                        <IconTrash :size="16" />
                    </WeblensButton>
                </div>
            </div>
            <div
                v-else
                :class="{ 'text-text-secondary': true }"
            >
                No API Keys found.
            </div>
        </div>

        <div :class="{ 'flex items-center': true }">
            <WeblensInput
                v-model:value="apiKeyName"
                :class="{ 'mt-2 mr-2': true }"
                placeholder="API Key Name"
            />
            <WeblensButton
                label="Create New API Key"
                :class="{ 'mt-2': true }"
                @click="generateNewAPIKey"
            />
        </div>
    </div>
</template>

<script setup lang="ts">
import type { UserInfo } from '@ethanrous/weblens-api'
import { IconTrash } from '@tabler/icons-vue'
import type { AxiosError } from 'axios'
import { useWeblensAPI } from '~/api/AllApi'
import WeblensButton from '~/components/atom/WeblensButton.vue'
import WeblensInput from '~/components/atom/WeblensInput.vue'
import CopyBox from '~/components/molecule/CopyBox.vue'
import useConfirmDialog from '~/stores/confirm'

const userStore = useUserStore()

const fullName = ref<string>()
const oldPassword = ref<string>()
const newPassword = ref<string>()
const apiKeyName = ref<string>()
const updatePasswordButton = ref<typeof WeblensButton>()
const { open: openConfirmDialog } = useConfirmDialog()

const { data: apiKeys, refresh } = useAsyncData('apiKeys', async () => {
    return useWeblensAPI()
        .APIKeysAPI.getAPIKeys()
        .then((res) => res.data)
})

async function updateFullName() {
    if (!fullName.value || fullName.value === userStore.user.fullName) {
        return
    }

    const res = await useWeblensAPI().UsersAPI.changeDisplayName(userStore.user.username, fullName.value)
    userStore.setUser(res.data as UserInfo, true)
}

const canChangePassword = computed(() => {
    return !!oldPassword.value && !!newPassword.value && oldPassword.value !== newPassword.value
})

async function updatePassword() {
    if (!canChangePassword.value) {
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

function generateNewAPIKey() {
    if (!apiKeyName.value) {
        return
    }

    useWeblensAPI()
        .APIKeysAPI.createAPIKey({ name: apiKeyName.value })
        .then(() => {
            refresh()
        })
}

function deleteAPIKey(keyId: string, keyNickname: string) {
    openConfirmDialog({
        actionVerb: 'delete',
        actionItemName: `API key "${keyNickname}"`,
        onAccept: () =>
            useWeblensAPI()
                .APIKeysAPI.deleteAPIKey(keyId)
                .then(() => {
                    refresh()
                }),
    })
}

watchEffect(() => {
    if (userStore.user.fullName) {
        fullName.value = userStore.user.fullName
    }
})
</script>
