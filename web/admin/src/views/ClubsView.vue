<template>
  <div class="clubs-view">
    <div class="page-header">
      <h1>Clubs</h1>
      <button v-if="authStore.isSuperAdmin" class="btn btn--primary" @click="openCreateForm">Add Club</button>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error-message">{{ error }}</div>
    <div v-else>
      <table v-if="clubs.length > 0" class="data-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Telegram Chat ID</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="club in clubs" :key="club.id">
            <td>
              <router-link :to="`/clubs/${club.id}/members`">{{ club.name }}</router-link>
            </td>
            <td>{{ club.telegram_chat_id }}</td>
            <td class="actions-cell">
              <button class="btn btn--small" @click="openEditForm(club)">Edit</button>
              <button class="btn btn--small btn--danger" @click="handleDelete(club)">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">No clubs found. Create one to get started.</div>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal">
        <h2>{{ editing ? 'Edit Club' : 'Create Club' }}</h2>
        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="name">Name</label>
            <input id="name" v-model="form.name" type="text" required />
          </div>
          <div class="form-group">
            <label for="telegram_chat_id">Telegram Chat ID</label>
            <input id="telegram_chat_id" v-model.number="form.telegram_chat_id" type="number" required />
          </div>
          <div class="form-group form-group--checkbox">
            <label><input type="checkbox" v-model="form.welcome_enabled" /> Welcome notifications</label>
          </div>
          <div class="form-group form-group--checkbox">
            <label><input type="checkbox" v-model="form.birthday_enabled" /> Birthday notifications</label>
          </div>
          <div class="form-group form-group--checkbox">
            <label><input type="checkbox" v-model="form.race_notify_enabled" /> Race notifications</label>
          </div>
          <div v-if="formError" class="error-message">{{ formError }}</div>
          <div class="modal-actions">
            <button type="button" class="btn" @click="closeModal">Cancel</button>
            <button type="submit" class="btn btn--primary" :disabled="submitting">
              {{ submitting ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import clubApi from '@/api/club'
import type { Club } from '@/api/club'
import { useAuthStore } from '@/stores/auth'

interface ClubForm {
  name: string
  telegram_chat_id: number | ''
  welcome_enabled: boolean
  birthday_enabled: boolean
  race_notify_enabled: boolean
}

const clubs = ref<Club[]>([])
const loading = ref(true)
const error = ref('')
const showModal = ref(false)
const editing = ref<Club | null>(null)
const submitting = ref(false)
const formError = ref('')

const defaultForm = (): ClubForm => ({
  name: '',
  telegram_chat_id: '',
  welcome_enabled: true,
  birthday_enabled: true,
  race_notify_enabled: true,
})

const form = ref<ClubForm>(defaultForm())
const authStore = useAuthStore()

async function loadClubs() {
  loading.value = true
  error.value = ''
  try {
    const res = await clubApi.list()
    clubs.value = res.data
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load clubs'
  } finally {
    loading.value = false
  }
}

function openCreateForm() {
  editing.value = null
  form.value = defaultForm()
  formError.value = ''
  showModal.value = true
}

function openEditForm(club: Club) {
  editing.value = club
  form.value = {
    name: club.name,
    telegram_chat_id: club.telegram_chat_id,
    welcome_enabled: club.welcome_enabled,
    birthday_enabled: club.birthday_enabled,
    race_notify_enabled: club.race_notify_enabled,
  }
  formError.value = ''
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  editing.value = null
}

async function handleSubmit() {
  formError.value = ''
  submitting.value = true
  try {
    const payload = {
      name: form.value.name,
      telegram_chat_id: Number(form.value.telegram_chat_id),
      welcome_enabled: form.value.welcome_enabled,
      birthday_enabled: form.value.birthday_enabled,
      race_notify_enabled: form.value.race_notify_enabled,
    }
    if (editing.value) {
      await clubApi.update(editing.value.id, payload)
    } else {
      await clubApi.create(payload)
    }
    closeModal()
    await loadClubs()
  } catch (e: any) {
    formError.value = e.response?.data?.error || 'Failed to save club'
  } finally {
    submitting.value = false
  }
}

async function handleDelete(club: Club) {
  if (!confirm(`Delete club "${club.name}"?`)) return
  try {
    await clubApi.remove(club.id)
    await loadClubs()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to delete club'
  }
}

onMounted(loadClubs)
</script>
