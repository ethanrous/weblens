<template>
    <div @click.stop>
        <MglMap
            :bounds="bounds"
            :center="compCords"
            :zoom="6"
        >
            <MglMarker
                :coordinates="compCords"
                color="#cc0000"
                :scale="0.5"
            />
        </MglMap>
    </div>
</template>

<script setup lang="ts">
import type { LngLatBoundsLike } from 'maplibre-gl'
import { MglDefaults, MglMap, MglMarker } from 'vue-maplibre-gl'

const bounds = ref<LngLatBoundsLike>()

const props = defineProps<{ coords: [number, number] }>()
const compCords = computed(() => {
    return [...props.coords].reverse() as [number, number]
})

const key = 'vKmgtyP5fhJvYl5mB3Zd'
MglDefaults.style = `https://api.maptiler.com/maps/voyager/style.json?key=${key}`
</script>

<style>
@import 'maplibre-gl/dist/maplibre-gl.css';
@import 'vue-maplibre-gl/dist/vue-maplibre-gl.css';

.map-container {
    height: 400px;
    width: 800px;
    max-width: 100%;
    resize: both;
    overflow: auto;
    border: 1px solid #d6d6d6;
    margin-bottom: 20px;
}
</style>
