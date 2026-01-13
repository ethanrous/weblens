import { defineStore } from 'pinia'
import { useWeblensAPI } from '~/api/AllApi'

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
