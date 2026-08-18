import type { CurrentOperator } from '~/types/user'

export function useCurrentOperator() {
  const config = useRuntimeConfig()

  const operator = ref<CurrentOperator | null>(null)
  const loading = ref(false)
  const error = ref<unknown>(null)

  async function load() {
    loading.value = true
    error.value = null

    try {
      operator.value = await $fetch<CurrentOperator>(
        '/api/current-user',
        {
          baseURL: config.public.apiBase,
        },
      )
    } catch (err) {
      error.value = err
      operator.value = null
    } finally {
      loading.value = false
    }
  }

  return {
    operator,
    loading,
    error,
    load,
  }
}
