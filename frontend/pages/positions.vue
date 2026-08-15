<script setup lang="ts">
interface Position {
  id: number
  name: string
}

const config = useRuntimeConfig()

const { data, error, pending } = await useFetch<Position[]>('/api/positions', {
  baseURL: config.public.apiBase,
})
</script>

<template>
  <main style="font-family: sans-serif; padding: 2rem;">
    <p>
      <NuxtLink to="/">&larr; назад</NuxtLink>
    </p>

    <h1>Должности</h1>

    <p v-if="pending">Загрузка...</p>
    <p v-else-if="error">Не удалось загрузить: {{ error.message }}</p>
    <p v-else-if="!data?.length">Должности ещё не созданы.</p>

    <ul v-else>
      <li v-for="p in data" :key="p.id">{{ p.name }}</li>
    </ul>
  </main>
</template>
