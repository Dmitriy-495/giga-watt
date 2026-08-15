<script setup lang="ts">
interface ObjectType {
  id: number
  name: string
}

const config = useRuntimeConfig()

const { data, error, pending } = await useFetch<ObjectType[]>('/api/object-types', {
  baseURL: config.public.apiBase,
})
</script>

<template>
  <main style="font-family: sans-serif; padding: 2rem;">
    <p>
      <NuxtLink to="/">&larr; назад</NuxtLink>
    </p>

    <h1>Типы объектов эксплуатации</h1>

    <p v-if="pending">Загрузка...</p>
    <p v-else-if="error">Не удалось загрузить: {{ error.message }}</p>
    <p v-else-if="!data?.length">Типы объектов ещё не созданы.</p>

    <ul v-else>
      <li v-for="t in data" :key="t.id">{{ t.name }}</li>
    </ul>
  </main>
</template>
