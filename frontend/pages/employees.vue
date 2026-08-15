<script setup lang="ts">
interface Employee {
  id: number
  last_name: string
  first_name: string
  middle_name: string
  short_name: string
  birth_date?: string
  gender?: string
}

const genderLabels: Record<string, string> = {
  male: 'М',
  female: 'Ж',
}

const config = useRuntimeConfig()

const { data, error, pending } = await useFetch<Employee[]>('/api/employees', {
  baseURL: config.public.apiBase,
})
</script>

<template>
  <main style="font-family: sans-serif; padding: 2rem;">
    <p>
      <NuxtLink to="/">&larr; назад</NuxtLink>
    </p>

    <h1>Сотрудники</h1>

    <p v-if="pending">Загрузка...</p>
    <p v-else-if="error">Не удалось загрузить: {{ error.message }}</p>
    <p v-else-if="!data?.length">Сотрудники ещё не созданы.</p>

    <table v-else style="border-collapse: collapse;">
      <thead>
        <tr>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">ID</th>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">ФИО (сокращённо)</th>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Дата рождения</th>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Пол</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="e in data" :key="e.id">
          <td style="padding: 0.25rem 0.75rem;">{{ e.id }}</td>
          <td style="padding: 0.25rem 0.75rem;">{{ e.short_name }}</td>
          <td style="padding: 0.25rem 0.75rem;">{{ e.birth_date?.slice(0, 10) ?? '—' }}</td>
          <td style="padding: 0.25rem 0.75rem;">{{ e.gender ? genderLabels[e.gender] ?? e.gender : '—' }}</td>
        </tr>
      </tbody>
    </table>
  </main>
</template>
