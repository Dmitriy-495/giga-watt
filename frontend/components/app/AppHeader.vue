<template>
  <header class="app-header">
    <div class="app-header__brand">
      GIGA WATT
    </div>

    <div
      v-if="operator?.user"
      class="app-header__operator"
    >
      <span>{{ operator.user.email }}</span>
      <span
        v-if="operator.user.is_superuser"
        class="app-header__superuser"
      >
        SUPERUSER
      </span>
    </div>
  </header>
</template>

<script setup lang="ts">
import type { CurrentOperator } from '~/types/user'

defineProps<{
  operator: CurrentOperator | null
}>()
</script>
