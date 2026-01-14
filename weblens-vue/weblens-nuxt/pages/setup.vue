<template>
    <div
        v-if="towerStore.towerInfo?.role === TowerRole.UNINITIALIZED"
        :class="{ 'flex h-screen w-screen flex-col items-center': true }"
    >
        <div class="mt-[4vh] mb-[20vh] flex w-full justify-center text-center">
            <Logo :size="100" />
            <h1 class="xs:inline-block mt-auto hidden">EBLENS</h1>
        </div>

        <form
            id="general-info"
            :class="{
                'mx-auto flex h-max w-110 min-w-0 flex-col gap-3 overflow-hidden rounded border p-4': true,
            }"
        >
            <span>Name this Weblens server</span>
            <WeblensInput
                v-model:value="serverName"
                placeholder="Server Name"
                auto-focus
                square-size="{44}"
                auto-complete="server-address"
            />

            <span>Create a server owner account</span>
            <WeblensInput
                v-model:value="ownerDisplayName"
                key-name="John Doe"
                placeholder="Owner Display Name"
                square-size="{44}"
                auto-complete="owner-username"
            />
            <WeblensInput
                key-name="john_doe"
                placeholder="Owner Username"
                square-size="{44}"
                auto-complete="owner-username"
                :value="ownerUsername"
                @update:value="
                    (u) => {
                        if (u.includes(' ')) {
                            u = u.replace(/ /g, '_')
                        }
                        if (u.match(/[^A-Za-z0-9_]/g)) {
                            u = u.replace(/[^a-z0-9_]/g, '')
                        }
                        if (u !== u.toLowerCase()) {
                            u = u.toLowerCase()
                        }

                        console.debug('Updating ownerUsername to', u)
                        ownerUsername = u
                        usernameSetManually = u.length > 0
                    }
                "
            />

            <WeblensInput
                v-model:value="ownerPassword"
                placeholder="Owner Password"
                square-size="{44}"
                password
                auto-complete="owner-password"
                @keydown:enter="setUpServer"
            />
        </form>

        <div
            :class="{
                'flex h-25 flex-col items-center overflow-hidden transition-[height]': true,
                'h-0!': backupDropdownOpen,
            }"
        >
            <WeblensButton
                :label="loading && !backupDropdownOpen ? 'Setting up...' : 'Set up as a Core Server'"
                :square-size="50"
                :disabled="
                    serverName === '' ||
                    ownerDisplayName === '' ||
                    ownerUsername === '' ||
                    ownerPassword.length < 6 ||
                    backupDropdownOpen ||
                    loading
                "
                center-content
                :class="{ 'mt-4 mb-3 w-110': true }"
                @click="setUpServer(TowerRole.CORE)"
            >
                <template #rightIcon>
                    <IconArrowRight />
                </template>
            </WeblensButton>

            <span :class="{ 'mt-3': true }">Or</span>
        </div>

        <!-- The backup server options -->
        <div
            :class="{
                'mt-6 flex w-110 flex-col rounded border transition-[height]': true,
            }"
        >
            <span
                :class="{
                    'hover:bg-background-hover mx-auto inline-flex h-10 w-full cursor-pointer items-center rounded px-4': true,
                }"
                @click="backupDropdownOpen = !backupDropdownOpen"
            >
                <IconChevronRight
                    v-if="!backupDropdownOpen"
                    :class="{ inline: true }"
                />
                <IconChevronDown
                    v-if="backupDropdownOpen"
                    :class="{ inline: true }"
                />
                <span :class="{ 'inline w-full text-center': true }"> Set up as a Backup Server </span>
            </span>

            <form
                :class="{
                    'mx-auto flex w-full min-w-0 flex-col gap-3 overflow-hidden px-4': true,
                    'mt-0 h-0': !backupDropdownOpen,
                    'h-max border-t pt-4': backupDropdownOpen,
                }"
            >
                <WeblensInput
                    v-model:value="remoteAddress"
                    key-name="https://weblens.example.com"
                    placeholder="Core Server Address"
                    square-size="{44}"
                    auto-complete="core-server-address"
                />
                <WeblensInput
                    v-model:value="apiKey"
                    placeholder="API Key"
                    square-size="{44}"
                    password
                    auto-complete="api-key"
                    @keydown:enter="setUpServer"
                />
                <span
                    v-if="formError"
                    class="text-center text-red-500"
                >
                    {{ formError }}
                </span>
                <WeblensButton
                    :label="loading ? 'Setting up...' : 'Set up Backup Server'"
                    fill-width
                    :square-size="50"
                    :disabled="
                        serverName === '' ||
                        ownerDisplayName === '' ||
                        ownerUsername === '' ||
                        remoteAddress === '' ||
                        apiKey === '' ||
                        ownerPassword.length < 6 ||
                        !validCoreUrl ||
                        loading
                    "
                    center-content
                    :class="{ 'mt-auto mb-3': true }"
                    @click="setUpServer(TowerRole.BACKUP)"
                >
                    <template #rightIcon>
                        <IconArrowRight />
                    </template>
                </WeblensButton>
            </form>
        </div>
    </div>
</template>

<script setup lang="ts">
import Logo from '~/components/atom/Logo.vue'
import WeblensInput from '~/components/atom/WeblensInput.vue'
import WeblensButton from '~/components/atom/WeblensButton.vue'
import { IconArrowRight, IconChevronDown, IconChevronRight } from '@tabler/icons-vue'
import { useWeblensAPI } from '~/api/AllApi'
import useLocationStore from '~/stores/location'

const towerStore = useTowerStore()
const backupDropdownOpen = ref(false)

// Hook into location store to ensure redirects work
// properly before and after setup
useLocationStore()

const serverName = ref('')
const ownerDisplayName = ref('')
const ownerUsername = ref('')
const ownerPassword = ref('')
const remoteAddress = ref('')
const apiKey = ref('')
const formError = ref('')
const loading = ref(false)

const usernameSetManually = ref(false)

const validCoreUrl = computed(() => {
    if (remoteAddress.value === '') {
        return false
    }
    try {
        const url = new URL(remoteAddress.value)
        return url.protocol === 'http:' || url.protocol === 'https:'
    } catch {
        return false
    }
})

function setUpServer(towerTypeSelection: TowerRole) {
    loading.value = true
    formError.value = ''

    if (serverName.value === '') {
        formError.value = 'Server Name is required.'
        loading.value = false
        return
    }

    if (ownerUsername.value === '') {
        formError.value = 'owner Username is required for core servers.'
        loading.value = false
        return
    }

    if (ownerPassword.value === '') {
        formError.value = 'owner Password is required for core servers.'
        loading.value = false
        return
    }

    if (towerTypeSelection === TowerRole.BACKUP) {
        if (remoteAddress.value === '') {
            formError.value = 'Core server address is required for backup servers.'
            loading.value = false
            return
        }

        if (apiKey.value === '') {
            formError.value = 'API Key is required for backup servers.'
            loading.value = false
            return
        }
    }

    useWeblensAPI()
        .TowersAPI.initializeTower({
            name: serverName.value,
            fullName: ownerDisplayName.value,
            username: ownerUsername.value,
            password: ownerPassword.value,
            role: towerTypeSelection,
            coreAddress: remoteAddress.value,
            coreKey: apiKey.value,
        })
        .then(() => {
            return towerStore.refreshTowerInfo()
        })
        .finally(() => {
            loading.value = false
        })
}
watchEffect(() => {
    if (ownerDisplayName.value.length > 0 && !usernameSetManually.value) {
        ownerUsername.value = ownerDisplayName.value
            .toLowerCase()
            .replace(/[^a-z0-9]/g, '_')
            .replace(/_+/g, '_')
            .replace(/^_+|_+$/g, '')
    }
})
</script>
