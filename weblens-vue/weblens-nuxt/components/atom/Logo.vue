<template>
    <svg
        xmlns="http://www.w3.org/2000/svg"
        version="1.1"
        viewBox="0 0 768.2 748"
        :style="{
            width: size + 'px',
            height: size + 'px',
            flexShrink: 0,
            transition: 'all 200ms ease-in-out',
        }"
    >
        <defs>
            <linearGradient
                id="linear-gradient"
                x1="136.9"
                y1="60.8"
                x2="538.5"
                y2="756.3"
                gradientTransform="translate(-1.2 748.8) rotate(.2) scale(1 -1)"
                gradientUnits="userSpaceOnUse"
            >
                <stop
                    offset="0"
                    stop-color="#612d96"
                />
                <stop
                    offset="0.2"
                    stop-color="#5c2d9b"
                />
                <stop
                    offset="0.4"
                    stop-color="#4f30aa"
                />
                <stop
                    offset="0.6"
                    stop-color="#3f34be"
                />
                <stop
                    offset="1"
                    stop-color="#0082c6"
                />
            </linearGradient>
        </defs>
        <g>
            <path
                fill="url(#linear-gradient)"
                class="cls-1"
                d="M462.1,88.8l49.7,221.8c2.2,9.8,16.3,9.4,17.9-.5l36.8-220.4c.7-4.4,4.6-7.7,9-7.6l167.4.5c5.9,0,10.2,5.5,8.9,11.2l-131,566.4c-1,4.2-4.7,7.1-8.9,7.1l-159.7-.5c-4.2,0-7.9-2.9-8.9-7l-51.2-215.6c-2.2-9.3-15.5-9.4-17.8,0l-52.5,215.3c-1,4.1-4.7,7-8.9,7l-159.8-.5c-4.3,0-8-3-8.9-7.1L17,91.4c-1.3-5.7,3.1-11.2,8.9-11.1l167.4.5c4.5,0,8.3,3.3,9,7.7l34.8,220.1c1.6,9.9,15.6,10.4,17.9.7l51.7-221c1-4.1,4.7-7.1,8.9-7.1l137.6.4c4.3,0,8,3,8.9,7.1Z"
            />
        </g>
    </svg>
</template>

<script setup lang="ts">
defineProps<{
    size?: number
}>()
</script>
