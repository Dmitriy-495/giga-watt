<template>
  <section class="workspace">
    <div class="workspace__heading">
      <p class="workspace__eyebrow">
        Рабочее пространство
      </p>

      <h1 class="workspace__title">
        {{ organizationName }}
      </h1>
    </div>

    <div class="workspace__sections">
      <NuxtLink
        to="/employees"
        class="workspace-card"
      >
        <h2>Сотрудники</h2>
        <p>
          Сотрудники предприятия и их рабочие назначения.
        </p>
      </NuxtLink>

      <NuxtLink
        to="/staff-positions"
        class="workspace-card"
      >
        <h2>Штатная структура</h2>
        <p>
          Штатные позиции и организационная структура.
        </p>
      </NuxtLink>

      <NuxtLink
        to="/operational-objects"
        class="workspace-card"
      >
        <h2>Объекты эксплуатации</h2>
        <p>
          Объекты, находящиеся в рабочем контексте подразделения.
        </p>
      </NuxtLink>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { OrganizationContext } from '~/types/user'

const props = defineProps<{
  context: OrganizationContext | null
}>()

const organizationName = computed(() => {
  return props.context?.current?.name ?? 'Рабочий контекст не определён'
})
</script>
