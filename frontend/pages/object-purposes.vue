<script setup lang="ts">
interface ObjectType {
  id: number
  name: string
}

interface ObjectPurpose {
  id: number
  object_type_id: number
  name: string
}

const config = useRuntimeConfig()

const { data: types } = await useFetch<ObjectType[]>('/api/object-types', {
  baseURL: config.public.apiBase,
})

const { data, error, pending } = await useFetch<ObjectPurpose[]>('/api/object-purposes', {
  baseURL: config.public.apiBase,
})

function typeName(id: number): string {
  return types.value?.find((t) => t.id === id)?.name ?? `#${id}`
}
</script>

<template>
  <main style="font-family: sans-serif; padding: 2rem;">
    <p>
      <NuxtLink to="/">&larr; назад</NuxtLink>
    </p>

    <h1>Назначения объектов эксплуатации</h1>

    <p v-if="pending">Загрузка...</p>
    <p v-else-if="error">Не удалось загрузить: {{ error.message }}</p>
    <p v-else-if="!data?.length">Назначения ещё не созданы.</p>

    <table v-else style="border-collapse: collapse;">
      <thead>
        <tr>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Назначение</th>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Тип объекта</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="p in data" :key="p.id">
          <td style="padding: 0.25rem 0.75rem;">{{ p.name }}</td>
          <td style="padding: 0.25rem 0.75rem;">{{ typeName(p.object_type_id) }}</td>
        </tr>
      </tbody>
    </table>
  </main>
</template>
