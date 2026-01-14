<template>
    <div :class="{ 'page-root flex-col!': true }">
        <div :class="{ 'mt-4 ml-4': true }">
            <span :class="{ 'inline-flex items-center gap-2 text-lg font-semibold': true }">
                {{ towerStore.towerInfo?.name }}
                <WebsocketStatus
                    :ws-status="websocketStore.status"
                    :show-as-text="websocketStore.status !== 'OPEN'"
                />
            </span>
            <h5 :class="{ 'text-text-secondary': true }">
                {{ capitalizeFirstLetter(towerStore.towerInfo?.role ?? '') }}
            </h5>
        </div>

        <div
            v-if="towers"
            :class="{ 'my-4 flex flex-col border-t px-4 pt-2': true }"
        >
            <h1 :class="{ 'mb-4 text-2xl font-semibold': true }">Linked Towers</h1>
            <ul :class="{ 'space-y-2': true }">
                <li
                    v-for="tower in towers"
                    :key="tower.id"
                    :class="{
                        'hover:bg-card-background-hover bg-card-background-primary max-w-64 cursor-pointer rounded border p-4 shadow transition duration-150': true,
                        'bg-card-background-selected hover:bg-card-background-selected-hover':
                            tower.id === selectedTowerID,
                    }"
                    @click="selectedTowerID = selectedTowerID === tower.id ? null : tower.id"
                >
                    <h2 :class="{ 'text-lg font-medium': true }">{{ tower.name }}</h2>
                    <p :class="{ 'text-text-secondary': true }">ID: {{ tower.id }}</p>
                </li>
            </ul>
        </div>

        <!-- Selected Tower Info -->
        <div :class="{ 'm-2 flex h-full flex-col rounded border border-t px-4 pt-2': true }">
            <div
                v-if="selectedTower"
                :class="{ 'flex flex-col gap-2': true }"
            >
                <span :class="{ 'mb-4': true }">Tower Details</span>

                <span :class="{ 'text-lg': true }">
                    <IconServer class="inline" />
                    {{ selectedTower.name }}
                </span>
                <span :class="{ 'inline-flex items-center gap-1 text-lg': true }">
                    <IconId class="inline" />
                    <CopyBox :text="selectedTower.id" />
                </span>
                <span :class="{ 'inline-flex items-center gap-1 text-lg': true }">
                    <IconAddressBook class="inline" />
                    <CopyBox :text="selectedTower.coreAddress" />
                </span>
                <span :class="{ 'inline-flex items-center gap-1 text-lg': true }">
                    {{ selectedTower.backupSize }}
                </span>
                <WebsocketStatus
                    :ws-status="selectedTower.online ? 'OPEN' : 'CLOSED'"
                    show-as-text
                    :size="12"
                />

                <WeblensButton
                    label="Backup Now"
                    :class="{ 'mt-4 w-32': true }"
                    center-content
                    @click="useWeblensAPI().TowersAPI.launchBackup(selectedTower.id)"
                />
            </div>
            <p
                v-else
                :class="{ 'text-text-tertiary m-auto text-center': true }"
            >
                Select a tower to view its details.
            </p>
        </div>

        <WeblensButton
            label="Settings"
            allow-collapse
            :class="{ 'm-3': true }"
            @click.stop="navigateTo('/settings')"
        >
            <IconSettings size="18" />
        </WeblensButton>
    </div>
</template>

<script setup lang="ts">
import { IconAddressBook, IconId, IconServer, IconSettings } from '@tabler/icons-vue'
import { useWeblensAPI } from '~/api/AllApi'
import WeblensButton from '~/components/atom/WeblensButton.vue'
import WebsocketStatus from '~/components/atom/WebsocketStatus.vue'
import CopyBox from '~/components/molecule/CopyBox.vue'
import useWebsocketStore from '~/stores/websocket'

import { capitalizeFirstLetter } from '~/util/string'
const towerStore = useTowerStore()
const websocketStore = useWebsocketStore()

const selectedTowerID = ref<string | null>(null)

const { data: towers } = useAsyncData('towers', async () => {
    return useWeblensAPI()
        .TowersAPI.getRemotes()
        .then((res) => res.data.filter((tower) => tower.role !== TowerRole.BACKUP))
})

const selectedTower = computed(() => {
    return towers.value?.find((tower) => tower.id === selectedTowerID.value) || null
})

watchEffect(() => {
    if (towers.value?.length === 1) {
        selectedTowerID.value = towers.value[0].id
    }
})
</script>
