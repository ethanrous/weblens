<template>
    <div
        v-if="towerStore.towerInfo?.role === TowerRole.INIT"
        :class="{ 'flex h-screen w-screen flex-col items-center': true }"
    >
        <div class="mt-[4vh] mb-[30vh] flex w-full justify-center text-center">
            <Logo :size="100" />
            <h1 class="xs:inline-block mt-auto hidden">EBLENS</h1>
        </div>

        <!-- Two expand options for each server type, both include a form with options to setup server -->
        <!-- First, core server option -->

        <div
            :class="{
                'mb-4 flex w-1/2 flex-col rounded border transition-[height]': true,
            }"
        >
            <span
                :class="{
                    'hover:bg-background-hover mx-auto inline-flex h-12 w-full cursor-pointer items-center rounded px-4 leading-12': true,
                }"
                @click="towerTypeSelection = towerTypeSelection === TowerRole.CORE ? '' : TowerRole.CORE"
            >
                <IconChevronRight
                    v-if="towerTypeSelection !== TowerRole.CORE"
                    :class="{ inline: true }"
                />
                <IconChevronDown
                    v-if="towerTypeSelection === TowerRole.CORE"
                    :class="{ inline: true }"
                />
                <span :class="{ 'inline w-full text-center text-lg': true }"> Setup as a Core Server </span>
            </span>
            <form
                id="setup"
                :class="{
                    'mx-auto flex w-full min-w-0 flex-col gap-3 overflow-hidden px-4': true,
                    'mt-0 h-0': towerTypeSelection !== TowerRole.CORE,
                    'h-96 border-t pt-4': towerTypeSelection === TowerRole.CORE,
                }"
            >
                <WeblensInput
                    v-model:value="serverName"
                    placeholder="Server Name"
                    auto-focus
                    square-size="{44}"
                    auto-complete="server-address"
                />
                <WeblensInput
                    v-model:value="adminUsername"
                    placeholder="Admin Username"
                    square-size="{44}"
                    auto-complete="admin-username"
                />
                <WeblensInput
                    v-model:value="adminPassword"
                    placeholder="Admin Password"
                    square-size="{44}"
                    password
                    auto-complete="admin-password"
                />
                <span
                    v-if="formError"
                    class="text-center text-red-500"
                >
                    {{ formError }}
                </span>
                <div class="my-3">
                    <WeblensButton
                        :label="loading ? 'Setting up...' : 'Setup Server'"
                        fill-width
                        :square-size="50"
                        :disabled="serverName === '' || adminUsername === '' || adminPassword.length < 6 || loading"
                        center-content
                        @click="setupServer"
                    />
                </div>
            </form>
        </div>

        <div
            :class="{
                'text-button-text-disabled flex w-1/2 flex-col rounded border transition-[height]': true,
            }"
        >
            <span
                :class="{
                    'bg-abyss-500 mx-auto inline-flex h-12 w-full items-center rounded px-4 leading-12': true,
                }"
            >
                <IconChevronRight
                    v-if="towerTypeSelection !== TowerRole.BACKUP"
                    :class="{ inline: true }"
                />
                <IconChevronDown
                    v-if="towerTypeSelection === TowerRole.BACKUP"
                    :class="{ inline: true }"
                />
                <span :class="{ 'inline w-full text-center text-lg': true }"> Setup as a Backup Server </span>
            </span>

            <!-- not yet implemented -->
        </div>
    </div>
</template>

<script setup lang="ts">
import Logo from '~/components/atom/Logo.vue'
import WeblensInput from '~/components/atom/WeblensInput.vue'
import WeblensButton from '~/components/atom/WeblensButton.vue'
import { IconChevronDown, IconChevronRight } from '@tabler/icons-vue'
import { useWeblensAPI } from '~/api/AllApi'
import useLocationStore from '~/stores/location'

const towerTypeSelection = ref<TowerRole | ''>('')
const towerStore = useTowerStore()

// Hook into location store to ensure redirects work
// properly before and after setup
useLocationStore()

const serverName = ref('')
const adminUsername = ref('')
const adminPassword = ref('')
const formError = ref('')
const loading = ref(false)

function setupServer() {
    loading.value = true
    if (towerTypeSelection.value === '') {
        formError.value = 'Please select a server type to setup.'
        return
    }

    useWeblensAPI()
        .TowersAPI.initializeTower({
            name: serverName.value,
            username: adminUsername.value,
            password: adminPassword.value,
            role: towerTypeSelection.value,
        })
        .then(() => {
            console.log('Tower initialized successfully.')
            return towerStore.refreshTowerInfo()
        })
        .finally(() => {
            loading.value = false
        })
}
</script>
