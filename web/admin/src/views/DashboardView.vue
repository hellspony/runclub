<template>
  <div class="dashboard">
    <h1>Dashboard</h1>
    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error-message">{{ error }}</div>
    <div v-else class="dashboard-content">
      <div class="stat-card">
        <div class="stat-value">{{ clubs.length }}</div>
        <div class="stat-label">Clubs</div>
      </div>
      <div class="dashboard-actions">
        <router-link to="/clubs" class="btn btn--primary">Manage Clubs</router-link>
      </div>
      <div v-if="clubs.length > 0" class="club-list">
        <h2>Your Clubs</h2>
        <ul>
          <li v-for="club in clubs" :key="club.id">
            <router-link :to="`/clubs/${club.id}/members`">{{ club.name }}</router-link>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import clubApi from '@/api/club'
import type { Club } from '@/api/club'

const clubs = ref<Club[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    const res = await clubApi.list()
    clubs.value = res.data
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load clubs'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.dashboard h1 {
  margin: 0 0 24px;
  color: #1a1a2e;
}

.stat-card {
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  display: inline-block;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  margin-bottom: 24px;
}

.stat-value {
  font-size: 36px;
  font-weight: 700;
  color: #1a1a2e;
}

.stat-label {
  font-size: 14px;
  color: #666;
  margin-top: 4px;
}

.dashboard-actions {
  margin-bottom: 24px;
}

.club-list h2 {
  font-size: 18px;
  margin: 0 0 12px;
  color: #333;
}

.club-list ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.club-list li {
  padding: 8px 0;
}

.club-list li a {
  color: #4a4a8a;
  text-decoration: none;
}

.club-list li a:hover {
  text-decoration: underline;
}
</style>
