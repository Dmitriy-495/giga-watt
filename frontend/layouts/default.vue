<template>
  <div class="app-shell">
    <AppHeader :operator="operator" />

    <AppBreadcrumbs
      :ancestors="operator?.organization_context?.ancestors ?? []"
      :current="operator?.organization_context?.current ?? null"
    />

    <main class="app-main">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
import type { CurrentOperator } from '~/types/user'

const operator = useState<CurrentOperator | null>(
  'current-operator',
  () => null,
)
</script>
