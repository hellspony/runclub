<template>
  <div class="club-detail">
    <div class="club-detail-header">
      <router-link to="/clubs" class="back-link">&larr; Back to Clubs</router-link>
      <h1 v-if="club">{{ club.name }}</h1>
    </div>
    <div v-if="!club" class="loading">Loading club...</div>
    <router-view v-else />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import clubApi from '@/api/club'
import type { Club } from '@/api/club'

const route = useRoute()
const club = ref<Club | null>(null)

onMounted(async () => {
  try {
    const id = Number(route.params.id)
    const res = await clubApi.get(id)
    club.value = res.data
  } catch {
    // error handled by router
  }
})
</script>

<style scoped>
.club-detail-header {
  margin-bottom: 24px;
}

.back-link {
  color: #4a4a8a;
  text-decoration: none;
  font-size: 14px;
}

.back-link:hover {
  text-decoration: underline;
}

.club-detail h1 {
  margin: 8px 0 0;
  color: #1a1a2e;
}
</style>
