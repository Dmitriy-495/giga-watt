<template>
  <nav
    v-if="items.length > 0"
    class="breadcrumbs"
    aria-label="Навигация по организационной структуре"
  >
    <NuxtLink
      v-for="(item, index) in items"
      :key="item.id"
      :to="index === items.length - 1 ? undefined : `/organization/${item.id}`"
      :class="[
        'breadcrumbs__item',
        {
          'breadcrumbs__item--current': index === items.length - 1,
        },
      ]"
      :aria-current="index === items.length - 1 ? 'page' : undefined"
    >
      {{ item.name }}

      <span
        v-if="index < items.length - 1"
        class="breadcrumbs__separator"
        aria-hidden="true"
      >
        →
      </span>
    </NuxtLink>
  </nav>
</template>

<script setup lang="ts">
import type { OrganizationUnit } from '~/types/user'

const props = defineProps<{
  ancestors: OrganizationUnit[]
  current: OrganizationUnit | null
}>()

const items = computed(() => {
  if (!props.current) {
    return props.ancestors
  }

  return [...props.ancestors, props.current]
})
</script>
