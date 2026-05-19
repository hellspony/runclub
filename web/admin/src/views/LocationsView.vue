<template>
  <div class="locations-view">
    <div class="page-header">
      <h2>Locations</h2>
      <button class="btn btn--primary" @click="openCreateForm">Add Location</button>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error-message">{{ error }}</div>
    <div v-else>
      <table v-if="locations.length > 0" class="data-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Address</th>
            <th>Map URL</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="location in locations" :key="location.id">
            <td>{{ location.name }}</td>
            <td>{{ location.address }}</td>
            <td>
              <a v-if="location.map_url" :href="location.map_url" target="_blank" rel="noopener">View Map</a>
              <span v-else class="text-muted">--</span>
            </td>
            <td class="actions-cell">
              <button class="btn btn--small" @click="openEditForm(location)">Edit</button>
              <button class="btn btn--small btn--danger" @click="handleDelete(location)">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">No locations found.</div>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal">
        <h2>{{ editing ? 'Edit Location' : 'Add Location' }}</h2>
        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="name">Name</label>
            <input id="name" v-model="form.name" type="text" required />
          </div>
          <div class="form-group">
            <label for="address">Address</label>
            <input id="address" v-model="form.address" type="text" required />
          </div>
          <div class="form-group">
            <label for="map_url">Map URL</label>
            <input id="map_url" v-model="form.map_url" type="url" />
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
import locationApi from '@/api/location'
import type { Location } from '@/api/location'

const route = useRoute()
const clubId = Number(route.params.id)

const locations = ref<Location[]>([])
const loading = ref(true)
const error = ref('')
const showModal = ref(false)
const editing = ref<Location | null>(null)
const submitting = ref(false)
const formError = ref('')

const form = ref({
  name: '',
  address: '',
  map_url: '',
})

async function loadLocations() {
  loading.value = true
  error.value = ''
  try {
    const res = await locationApi.list(clubId)
    locations.value = res.data
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load locations'
  } finally {
    loading.value = false
  }
}

function openCreateForm() {
  editing.value = null
  form.value = { name: '', address: '', map_url: '' }
  formError.value = ''
  showModal.value = true
}

function openEditForm(location: Location) {
  editing.value = location
  form.value = { name: location.name, address: location.address, map_url: location.map_url }
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
      await locationApi.update(editing.value.id, form.value)
    } else {
      await locationApi.create(clubId, form.value)
    }
    closeModal()
    await loadLocations()
  } catch (e: any) {
    formError.value = e.response?.data?.error || 'Failed to save location'
  } finally {
    submitting.value = false
  }
}

async function handleDelete(location: Location) {
  if (!confirm(`Delete location "${location.name}"?`)) return
  try {
    await locationApi.remove(location.id)
    await loadLocations()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to delete location'
  }
}

onMounted(loadLocations)
</script>
