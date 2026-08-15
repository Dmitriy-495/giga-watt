<script setup lang="ts">
interface OrganizationUnit {
  id: number
  name: string
}

interface ObjectType {
  id: number
  name: string
}

interface ObjectPurpose {
  id: number
  name: string
}

interface OperationalObject {
  id: number
  organization_unit_id: number
  object_type_id: number
  object_purpose_id?: number
  name: string
  address?: string
}

const config = useRuntimeConfig()

const { data: units } = await useFetch<OrganizationUnit[]>('/api/organization-units', {
  baseURL: config.public.apiBase,
})

const { data: types } = await useFetch<ObjectType[]>('/api/object-types', {
  baseURL: config.public.apiBase,
})

const { data: purposes } = await useFetch<ObjectPurpose[]>('/api/object-purposes', {
  baseURL: config.public.apiBase,
})

const { data, error, pending } = await useFetch<OperationalObject[]>('/api/operational-objects', {
  baseURL: config.public.apiBase,
})

function unitName(id: number): string {
  return units.value?.find((u) => u.id === id)?.name ?? `#${id}`
}

function typeName(id: number): string {
  return types.value?.find((t) => t.id === id)?.name ?? `#${id}`
}

function purposeName(id?: number): string {
  if (!id) return '—'
  return purposes.value?.find((p) => p.id === id)?.name ?? `#${id}`
}
</script>

<template>
  <main style="font-family: sans-serif; padding: 2rem;">
    <p>
      <NuxtLink to="/">&larr; назад</NuxtLink>
    </p>

    <h1>Объекты эксплуатации</h1>

    <p v-if="pending">Загрузка...</p>
    <p v-else-if="error">Не удалось загрузить: {{ error.message }}</p>
    <p v-else-if="!data?.length">Объекты эксплуатации ещё не созданы.</p>

    <table v-else style="border-collapse: collapse;">
      <thead>
        <tr>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Наименование</th>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">ПУ</th>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Тип</th>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Назначение</th>
          <th style="text-align: left; border-bottom: 1px solid #ccc; padding: 0.25rem 0.75rem;">Адрес</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="o in data" :key="o.id">
          <td style="padding: 0.25rem 0.75rem;">{{ o.name }}</td>
          <td style="padding: 0.25rem 0.75rem;">{{ unitName(o.organization_unit_id) }}</td>
          <td style="padding: 0.25rem 0.75rem;">{{ typeName(o.object_type_id) }}</td>
          <td style="padding: 0.25rem 0.75rem;">{{ purposeName(o.object_purpose_id) }}</td>
          <td style="padding: 0.25rem 0.75rem;">{{ o.address ?? '—' }}</td>
        </tr>
      </tbody>
    </table>
  </main>
</template>
