<template>
  <div class="races-view">
    <div class="page-header">
      <h2>Races</h2>
      <button class="btn btn--primary" @click="openCreateForm">Add Race</button>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error-message">{{ error }}</div>
    <div v-else>
      <table v-if="races.length > 0" class="data-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Date</th>
            <th>Type</th>
            <th>Place</th>
            <th>Distances</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="race in races" :key="race.id">
            <td>{{ race.name }}</td>
            <td>{{ formatDate(race.date) }}</td>
            <td>{{ race.type }}</td>
            <td>{{ race.place }}</td>
            <td>{{ race.distances }}</td>
            <td class="actions-cell">
              <button class="btn btn--small" @click="openEditForm(race)">Edit</button>
              <button class="btn btn--small btn--danger" @click="handleDelete(race)">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">No races found.</div>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal">
        <h2>{{ editing ? 'Edit Race' : 'Add Race' }}</h2>
        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="name">Name</label>
            <input id="name" v-model="form.name" type="text" required />
          </div>
          <div class="form-group">
            <label for="date">Date</label>
            <input id="date" v-model="form.date" type="datetime-local" required />
          </div>
          <div class="form-group">
            <label for="type">Type</label>
            <input id="type" v-model="form.type" type="text" required />
          </div>
          <div class="form-group">
            <label for="place">Place</label>
            <input id="place" v-model="form.place" type="text" required />
          </div>
          <div class="form-group">
            <label for="distances">Distances</label>
            <input id="distances" v-model="form.distances" type="text" required placeholder="e.g. 5km, 10km, 21.1km" />
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
import { useRoute } from 'vue-router'
import raceApi from '@/api/race'
import type { Race } from '@/api/race'

const route = useRoute()
const clubId = Number(route.params.id)

const races = ref<Race[]>([])
const loading = ref(true)
const error = ref('')
const showModal = ref(false)
const editing = ref<Race | null>(null)
const submitting = ref(false)
const formError = ref('')

const form = ref({
  name: '',
  date: '',
  type: '',
  place: '',
  distances: '',
})

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' })
}

async function loadRaces() {
  loading.value = true
  error.value = ''
  try {
    const res = await raceApi.list(clubId)
    races.value = res.data
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load races'
  } finally {
    loading.value = false
  }
}

function openCreateForm() {
  editing.value = null
  form.value = { name: '', date: '', type: '', place: '', distances: '' }
  formError.value = ''
  showModal.value = true
}

function openEditForm(race: Race) {
  editing.value = race
  const dateVal = race.date ? race.date.slice(0, 16) : ''
  form.value = {
    name: race.name,
    date: dateVal,
    type: race.type,
    place: race.place,
    distances: race.distances,
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
    if (editing.value) {
      await raceApi.update(editing.value.id, form.value)
    } else {
      await raceApi.create(clubId, form.value)
    }
    closeModal()
    await loadRaces()
  } catch (e: any) {
    formError.value = e.response?.data?.error || 'Failed to save race'
  } finally {
    submitting.value = false
  }
}

async function handleDelete(race: Race) {
  if (!confirm(`Delete race "${race.name}"?`)) return
  try {
    await raceApi.remove(race.id)
    await loadRaces()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to delete race'
  }
}

onMounted(loadRaces)
</script>
