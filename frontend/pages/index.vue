<template>
  <div>
    <section
      v-if="loading"
      class="state state--loading"
    >
      Загрузка рабочего пространства…
    </section>

    <section
      v-else-if="error"
      class="state state--error"
    >
      <h1>Не удалось определить рабочего оператора</h1>

      <p>
        Backend не смог предоставить текущий пользовательский контекст.
      </p>

      <button
        type="button"
        class="state__action"
        @click="load"
      >
        Повторить
      </button>
    </section>

    <section
      v-else-if="!operator"
      class="state"
    >
      Рабочий контекст оператора не определён.
    </section>

    <section
      v-else-if="!operator.employee"
      class="state state--warning"
    >
      <h1>Сотрудник не найден</h1>

      <p>
        Пользователь Платформы найден, но соответствующий сотрудник
        предприятия не определён через employee_emails.
      </p>
    </section>

    <section
      v-else-if="!operator.organization_context?.current"
      class="state state--warning"
    >
      <h1>Рабочий контекст не определён</h1>

      <p>
        Для сотрудника не удалось определить текущее организационное
        подразделение.
      </p>
    </section>

    <OperatorWorkspace
      v-else
      :context="operator.organization_context"
    />
  </div>
</template>

<script setup lang="ts">
const {
  operator,
  loading,
  error,
  load,
} = useCurrentOperator()

const currentOperator = useState(
  'current-operator',
  () => null,
)

watch(
  operator,
  (value) => {
    currentOperator.value = value
  },
  { immediate: true },
)

await load()
</script>
