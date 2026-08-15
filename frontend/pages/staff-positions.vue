<script setup lang="ts">
interface OrganizationUnit {
  id: number
  name: string
}

interface Position {
  id: number
  name: string
}

interface StaffPosition {
  id: number
  organization_unit_id: number
  position_id: number
  quantity: number
}

const config = useRuntimeConfig()

const { data: units } = await useFetch<OrganizationUnit[]>('/api/organization-units', {
  baseURL: config.public.apiBase,
})

const { data: positions } = await useFetch<Position[]>('/api/positions', {
  baseURL: config.public.apiBase,
})

const { data, error, pending } = await useFetch<StaffPosition[]>('/api/staff-positions', {
  baseURL: config.public.apiBase,
})

function unitName(id: number): string {
  return units.value?.find((u) => u.id === id)?.name ?? `#${id}`
}

function positionName(id: number): string {
  return positions.value?.find((p) => p.id === id)?.name ?? `#${id}`
}
</script>

<template>
  <main style="font-family: sans-serif; padding: 2rem;">
    <p>
      <NuxtLink to="/">&larr; назад</NuxtLink>
    </p>

    <h1>Штатные единицы</h1>

    <p v-if="pending">Загрузка...</p>
    <p v-else-if="error">Не удалось загрузить: {{ error.message }}</p>
    <p v-else-if="!data?.length">Штатные единицы ещё не созданы.</p>

    <table v-else style="border-collapse: collapse;">
      <thead>
        <tr>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Организационная единица</th>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Должность</th>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Количество</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="s in data" :key="s.id">
          <td style="padding: 0.25rem 0.75rem;">{{ unitName(s.organization_unit_id) }}</td>
          <td style="padding: 0.25rem 0.75rem;">{{ positionName(s.position_id) }}</td>
          <td style="padding: 0.25rem 0.75rem;">{{ s.quantity }}</td>
        </tr>
      </tbody>
    </table>
  </main>
</template>
