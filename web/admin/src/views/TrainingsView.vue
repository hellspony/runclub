<template>
  <div class="trainings-view">
    <div class="page-header">
      <h2>Trainings</h2>
      <button class="btn btn--primary" @click="openCreateForm">Add Training</button>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error-message">{{ error }}</div>
    <div v-else>
      <table v-if="trainings.length > 0" class="data-table">
        <thead>
          <tr>
            <th>Date</th>
            <th>Location</th>
            <th>Duration (min)</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="training in trainings" :key="training.id">
            <td>{{ formatDate(training.date) }}</td>
            <td>{{ getLocationName(training.location_id) }}</td>
            <td>{{ training.duration }}</td>
            <td>
              <span class="status-badge" :class="`status--${training.status}`">{{ training.status }}</span>
            </td>
            <td class="actions-cell">
              <button class="btn btn--small" @click="viewParticipants(training)">Participants</button>
              <button class="btn btn--small" @click="openEditForm(training)">Edit</button>
              <button class="btn btn--small btn--danger" @click="handleDelete(training)">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">No trainings found.</div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal">
        <h2>{{ editing ? 'Edit Training' : 'Add Training' }}</h2>
        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="date">Date</label>
            <input id="date" v-model="form.date" type="datetime-local" required />
          </div>
          <div class="form-group">
            <label for="location_id">Location</label>
            <select id="location_id" v-model.number="form.location_id" required>
              <option value="">Select location</option>
              <option v-for="loc in locations" :key="loc.id" :value="loc.id">{{ loc.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label for="duration">Duration (minutes)</label>
            <input id="duration" v-model.number="form.duration" type="number" min="1" required />
          </div>
          <div class="form-group">
            <label for="status">Status</label>
            <select id="status" v-model="form.status">
              <option value="planned">Planned</option>
              <option value="active">Active</option>
              <option value="completed">Completed</option>
              <option value="cancelled">Cancelled</option>
            </select>
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

    <!-- Participants Modal -->
    <div v-if="showParticipants" class="modal-overlay" @click.self="showParticipants = false">
      <div class="modal">
        <h2>Training Participants</h2>
        <div v-if="participantsLoading" class="loading">Loading...</div>
        <div v-else-if="participants.length === 0" class="empty-state">No participants yet.</div>
        <ul v-else class="participant-list">
          <li v-for="p in participants" :key="p.id">{{ p.member_fio || p.member_id }}</li>
        </ul>
        <div class="modal-actions">
          <button class="btn" @click="showParticipants = false">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import trainingApi from '@/api/training'
import type { Training, Participant } from '@/api/training'
import locationApi from '@/api/location'
import type { Location } from '@/api/location'

const route = useRoute()
const clubId = Number(route.params.id)

const trainings = ref<Training[]>([])
const locations = ref<Location[]>([])
const loading = ref(true)
const error = ref('')
const showModal = ref(false)
const editing = ref<Training | null>(null)
const submitting = ref(false)
const formError = ref('')

const showParticipants = ref(false)
const participants = ref<Participant[]>([])
const participantsLoading = ref(false)

const form = ref({
  date: '',
  location_id: '' as number | '',
  duration: 60,
  status: 'planned',
})

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function getLocationName(locationId: number): string {
  const loc = locations.value.find((l) => l.id === locationId)
  return loc ? loc.name : String(locationId)
}

async function loadData() {
  loading.value = true
  error.value = ''
  try {
    const [trainingsRes, locationsRes] = await Promise.all([
      trainingApi.list(clubId),
      locationApi.list(clubId),
    ])
    trainings.value = trainingsRes.data
    locations.value = locationsRes.data
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load data'
  } finally {
    loading.value = false
  }
}

function openCreateForm() {
  editing.value = null
  form.value = { date: '', location_id: '', duration: 60, status: 'planned' }
  formError.value = ''
  showModal.value = true
}

function openEditForm(training: Training) {
  editing.value = training
  const dateVal = training.date ? training.date.slice(0, 16) : ''
  form.value = {
    date: dateVal,
    location_id: training.location_id,
    duration: training.duration,
    status: training.status,
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
      date: form.value.date,
      location_id: Number(form.value.location_id),
      duration: form.value.duration,
      status: form.value.status,
    }
    if (editing.value) {
      await trainingApi.update(editing.value.id, payload)
    } else {
      await trainingApi.create(clubId, payload)
    }
    closeModal()
    await loadData()
  } catch (e: any) {
    formError.value = e.response?.data?.error || 'Failed to save training'
  } finally {
    submitting.value = false
  }
}

async function handleDelete(training: Training) {
  if (!confirm('Delete this training?')) return
  try {
    await trainingApi.remove(training.id)
    await loadData()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to delete training'
  }
}

async function viewParticipants(training: Training) {
  showParticipants.value = true
  participantsLoading.value = true
  participants.value = []
  try {
    const res = await trainingApi.participants(training.id)
    participants.value = res.data
  } catch (e: any) {
    // silently fail
  } finally {
    participantsLoading.value = false
  }
}

onMounted(loadData)
</script>
