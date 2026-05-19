import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import client from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const username = ref<string | null>(localStorage.getItem('username'))
  const role = ref<string | null>(localStorage.getItem('role'))
  const userId = ref<number | null>(localStorage.getItem('user_id') ? Number(localStorage.getItem('user_id')) : null)

  const isAuthenticated = computed(() => !!token.value)
  const isSuperAdmin = computed(() => role.value === 'superadmin')

  async function login(user: string, password: string) {
    const response = await client.post('/auth/login', { username: user, password })
    const data = response.data
    token.value = data.token
    username.value = user
    localStorage.setItem('token', data.token)
    localStorage.setItem('username', user)

    // Fetch user info to get role
    await fetchMe()
  }

  async function fetchMe() {
    try {
      const response = await client.get('/auth/me')
      const data = response.data
      username.value = data.username
      role.value = data.role
      userId.value = data.id
      localStorage.setItem('username', data.username)
      localStorage.setItem('role', data.role)
      localStorage.setItem('user_id', String(data.id))
    } catch {
      // Token might be invalid
    }
  }

  function logout() {
    token.value = null
    username.value = null
    role.value = null
    userId.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    localStorage.removeItem('role')
    localStorage.removeItem('user_id')
  }

  return {
    token,
    username,
    role,
    userId,
    isAuthenticated,
    isSuperAdmin,
    login,
    logout,
    fetchMe,
  }
})
