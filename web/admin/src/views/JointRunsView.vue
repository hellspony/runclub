<template>
  <div class="joint-runs-view">
    <div class="page-header">
      <h2>Joint Runs</h2>
      <button class="btn btn--primary" @click="openCreateForm">Add Joint Run</button>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error-message">{{ error }}</div>
    <div v-else>
      <table v-if="jointRuns.length > 0" class="data-table">
        <thead>
          <tr>
            <th>Date</th>
            <th>Location</th>
            <th>Creator</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="run in jointRuns" :key="run.id">
            <td>{{ formatDate(run.date) }}</td>
            <td>{{ getLocationName(run.location_id) }}</td>
            <td>{{ getCreatorName(run.creator_id) }}</td>
            <td class="actions-cell">
              <button class="btn btn--small" @click="viewParticipants(run)">Participants</button>
              <button class="btn btn--small btn--danger" @click="handleDelete(run)">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">No joint runs found.</div>
    </div>

    <!-- Create Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal">
        <h2>Add Joint Run</h2>
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
            <label for="creator_id">Creator</label>
            <select id="creator_id" v-model.number="form.creator_id" required>
              <option value="">Select creator</option>
              <option v-for="member in members" :key="member.id" :value="member.id">{{ member.fio }}</option>
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
        <h2>Joint Run Participants</h2>
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
import jointRunApi from '@/api/jointrun'
import type { JointRun, JointRunParticipant } from '@/api/jointrun'
import locationApi from '@/api/location'
import type { Location } from '@/api/location'
import memberApi from '@/api/member'
import type { Member } from '@/api/member'

const route = useRoute()
const clubId = Number(route.params.id)

const jointRuns = ref<JointRun[]>([])
const locations = ref<Location[]>([])
const members = ref<Member[]>([])
const loading = ref(true)
const error = ref('')
const showModal = ref(false)
const submitting = ref(false)
const formError = ref('')

const showParticipants = ref(false)
const participants = ref<JointRunParticipant[]>([])
const participantsLoading = ref(false)

const form = ref({
  date: '',
  location_id: '' as number | '',
  creator_id: '' as number | '',
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

function getCreatorName(creatorId: number): string {
  const m = members.value.find((m) => m.id === creatorId)
  return m ? m.fio : String(creatorId)
}

async function loadData() {
  loading.value = true
  error.value = ''
  try {
    const [runsRes, locationsRes, membersRes] = await Promise.all([
      jointRunApi.list(clubId),
      locationApi.list(clubId),
      memberApi.list(clubId),
    ])
    jointRuns.value = runsRes.data
    locations.value = locationsRes.data
    members.value = membersRes.data
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load data'
  } finally {
    loading.value = false
  }
}

function openCreateForm() {
  form.value = { date: '', location_id: '', creator_id: '' }
  formError.value = ''
  showModal.value = true
}

function closeModal() {
  showModal.value = false
}

async function handleSubmit() {
  formError.value = ''
  submitting.value = true
  try {
    const payload = {
      date: form.value.date,
      location_id: Number(form.value.location_id),
      creator_id: Number(form.value.creator_id),
    }
    await jointRunApi.create(clubId, payload)
    closeModal()
    await loadData()
  } catch (e: any) {
    formError.value = e.response?.data?.error || 'Failed to create joint run'
  } finally {
    submitting.value = false
  }
}

async function handleDelete(run: JointRun) {
  if (!confirm('Delete this joint run?')) return
  try {
    await jointRunApi.remove(run.id)
    await loadData()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to delete joint run'
  }
}

async function viewParticipants(run: JointRun) {
  showParticipants.value = true
  participantsLoading.value = true
  participants.value = []
  try {
    const res = await jointRunApi.participants(run.id)
    participants.value = res.data
  } catch {
    // silently fail
  } finally {
    participantsLoading.value = false
  }
}

onMounted(loadData)
</script>
