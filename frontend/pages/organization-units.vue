<script setup lang="ts">
interface OrganizationUnit {
  id: number
  parent_id?: number
  type: string
  name: string
  location: string
  address: string
}

const typeLabels: Record<string, string> = {
  institution: 'Учреждение',
  branch: 'Филиал',
  jks: 'ЖКС',
  production_unit: 'ПУ',
}

const config = useRuntimeConfig()

const { data, error, pending } = await useFetch<OrganizationUnit[]>('/api/organization-units', {
  baseURL: config.public.apiBase,
})
</script>

<template>
  <main style="font-family: sans-serif; padding: 2rem;">
    <p>
      <NuxtLink to="/">&larr; назад</NuxtLink>
    </p>

    <h1>Организационная структура</h1>

    <p v-if="pending">Загрузка...</p>
    <p v-else-if="error">Не удалось загрузить: {{ error.message }}</p>
    <p v-else-if="!data?.length">Организационные единицы ещё не созданы.</p>

    <table v-else style="border-collapse: collapse;">
      <thead>
        <tr>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">ID</th>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Тип</th>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Наименование</th>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Местоположение</th>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Адрес</th>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Родитель</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="unit in data" :key="unit.id">
          <td style="padding: 0.25rem 0.75rem;">{{ unit.id }}</td>
          <td style="padding: 0.25rem 0.75rem;">{{ typeLabels[unit.type] ?? unit.type }}</td>
          <td style="padding: 0.25rem 0.75rem;">{{ unit.name }}</td>
          <td style="padding: 0.25rem 0.75rem;">{{ unit.location }}</td>
          <td style="padding: 0.25rem 0.75rem;">{{ unit.address }}</td>
          <td style="padding: 0.25rem 0.75rem;">{{ unit.parent_id ?? '—' }}</td>
        </tr>
      </tbody>
    </table>
  </main>
</template>
