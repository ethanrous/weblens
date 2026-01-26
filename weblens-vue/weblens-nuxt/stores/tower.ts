import type { TowerInfo } from '@ethanrous/weblens-api'
import { defineStore } from 'pinia'
import type { ShallowRef } from 'vue'
import { useWeblensAPI } from '~/api/AllApi'
import type { BackupInfo } from '~/types/backupTypes'

export enum TowerRole {
    CORE = 'core',
    BACKUP = 'backup',
    UNINITIALIZED = 'init',
}

export const useTowerStore = defineStore('tower', () => {
    const { data: towerInfo, refresh } = useAsyncData(
        'tower',
        async () => {
            const towerRes = await useWeblensAPI().TowersAPI.getServerInfo()
            return towerRes.data
        },
        { immediate: true, lazy: false },
    )

    return {
        towerInfo,
        refreshTowerInfo: refresh,
    }
})

export const useRemotesStore = defineStore('remotes', () => {
    const remotes: ShallowRef<Map<string, TowerInfo> | undefined> = shallowRef()

    async function refreshRemotes() {
        const remotes_response = await useWeblensAPI()
            .TowersAPI.getRemotes()
            .then((res) => res.data.filter((tower) => tower.role !== TowerRole.BACKUP))

        const remotes_map: Map<string, TowerInfo> = new Map()
        for (const tower of remotes_response) {
            remotes_map.set(tower.id, tower)
        }

        remotes.value = remotes_map
    }

    onMounted(() => {
        refreshRemotes()
    })

    return {
        remotes,
        refreshRemotes,
    }
})

export const useBackupStore = defineStore('backup', () => {
    const activeBackups: ShallowRef<Map<string, BackupInfo>> = shallowRef(new Map())

    function updateBackup(info: BackupInfo) {
        const existingInfo = activeBackups.value.get(info.TowerID)
        if (existingInfo) {
            existingInfo.merge(info)
            info = existingInfo
        }

        activeBackups.value.set(info.TowerID, info.clone())
        activeBackups.value = new Map(activeBackups.value)
    }

    function setBackupComplete(towerID: string, endTime: number) {
        const existingInfo = activeBackups.value.get(towerID)
        if (!existingInfo) {
            console.warn(`No existing backup info for towerID ${towerID} to set complete`)
            return
        }

        existingInfo.complete(endTime)

        activeBackups.value.set(towerID, existingInfo.clone())
        activeBackups.value = new Map(activeBackups.value)
    }

    return {
        activeBackups,
        updateBackup,
        setBackupComplete,
    }
})
