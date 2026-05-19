<template>
  <div class="layout">
    <aside class="sidebar">
      <div class="sidebar-header">
        <h2>RunClub</h2>
      </div>
      <nav class="sidebar-nav">
        <router-link to="/dashboard" class="nav-item" active-class="nav-item--active">
          Dashboard
        </router-link>
        <router-link to="/clubs" class="nav-item" active-class="nav-item--active">
          Clubs
        </router-link>
        <router-link v-if="authStore.isSuperAdmin" to="/admin-users" class="nav-item" active-class="nav-item--active">
          Admin Users
        </router-link>

        <template v-if="currentClub">
          <div class="nav-section-title">{{ currentClub.name }}</div>
          <router-link :to="`/clubs/${clubId}/members`" class="nav-item nav-item--sub" active-class="nav-item--active">
            Members
          </router-link>
          <router-link :to="`/clubs/${clubId}/locations`" class="nav-item nav-item--sub" active-class="nav-item--active">
            Locations
          </router-link>
          <router-link :to="`/clubs/${clubId}/races`" class="nav-item nav-item--sub" active-class="nav-item--active">
            Races
          </router-link>
          <router-link :to="`/clubs/${clubId}/templates`" class="nav-item nav-item--sub" active-class="nav-item--active">
            Templates
          </router-link>
          <router-link :to="`/clubs/${clubId}/trainings`" class="nav-item nav-item--sub" active-class="nav-item--active">
            Trainings
          </router-link>
          <router-link :to="`/clubs/${clubId}/joint-runs`" class="nav-item nav-item--sub" active-class="nav-item--active">
            Joint Runs
          </router-link>
        </template>
      </nav>
    </aside>

    <div class="main">
      <header class="topbar">
        <div class="topbar-left">
          <span class="topbar-title">RunClub Admin</span>
        </div>
        <div class="topbar-right">
          <span class="topbar-user">{{ authStore.username }}</span>
          <button class="btn btn--small" @click="handleLogout">Logout</button>
        </div>
      </header>

      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import clubApi from '@/api/club'
import type { Club } from '@/api/club'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const clubId = computed(() => {
  const id = route.params.id
  const raw = Array.isArray(id) ? id[0] : id || null
  return raw ? Number(raw) : null
})

const currentClub = ref<Club | null>(null)

const clubCache = new Map<number, Club>()

watch(clubId, async (newId) => {
  if (newId) {
    if (!clubCache.has(newId)) {
      try {
        const res = await clubApi.get(newId)
        clubCache.set(newId, res.data)
      } catch {
        // ignore
      }
    }
    currentClub.value = clubCache.get(newId) || null
  } else {
    currentClub.value = null
  }
}, { immediate: true })

function handleLogout() {
  authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.layout {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  width: 240px;
  background: #1a1a2e;
  color: #fff;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-header {
  padding: 20px 16px;
  border-bottom: 1px solid #2a2a4a;
}

.sidebar-header h2 {
  margin: 0;
  font-size: 20px;
}

.sidebar-nav {
  padding: 8px 0;
  flex: 1;
}

.nav-item {
  display: block;
  padding: 10px 16px;
  color: #c0c0d0;
  text-decoration: none;
  font-size: 14px;
  transition: background 0.2s, color 0.2s;
}

.nav-item:hover {
  background: #2a2a4a;
  color: #fff;
}

.nav-item--active {
  background: #3a3a5a;
  color: #fff;
}

.nav-item--sub {
  padding-left: 32px;
}

.nav-section-title {
  padding: 12px 16px 4px;
  font-size: 12px;
  text-transform: uppercase;
  color: #8888aa;
  letter-spacing: 0.5px;
}

.main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.topbar {
  height: 56px;
  background: #fff;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  flex-shrink: 0;
}

.topbar-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.topbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.topbar-user {
  font-size: 14px;
  color: #666;
}

.content {
  flex: 1;
  padding: 24px;
  background: #f5f5f8;
  overflow-y: auto;
}
</style>
