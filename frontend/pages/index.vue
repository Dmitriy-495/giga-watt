<script setup lang="ts">
const config = useRuntimeConfig()

const { data, error, pending } = await useFetch<{ status: string }>('/api/ping', {
  baseURL: config.public.apiBase,
})
</script>

<template>
  <main style="font-family: sans-serif; padding: 2rem;">
    <h1>Giga Watt</h1>
    <p>Enterprise Digital Twin Platform</p>

    <h2>Backend connection</h2>
    <p v-if="pending">Checking backend...</p>
    <p v-else-if="error">
      Backend unreachable: {{ error.message }}
    </p>
    <p v-else>
      Backend status: {{ data?.status }}
    </p>

    <p>
      <NuxtLink to="/organization-units">Организационная структура</NuxtLink>
    </p>
  </main>
</template>
