<template>
  <div class="templates-view">
    <div class="page-header">
      <h2>Templates</h2>
      <button class="btn btn--primary" @click="openCreateForm">Add Template</button>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error-message">{{ error }}</div>
    <div v-else>
      <table v-if="templates.length > 0" class="data-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Content</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="tmpl in templates" :key="tmpl.id">
            <td>{{ tmpl.name }}</td>
            <td>{{ formatType(tmpl.type) }}</td>
            <td class="content-cell">{{ truncate(tmpl.content, 80) }}</td>
            <td class="actions-cell">
              <button class="btn btn--small" @click="openEditForm(tmpl)">Edit</button>
              <button class="btn btn--small btn--danger" @click="handleDelete(tmpl)">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">No templates found.</div>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal modal--wide">
        <h2>{{ editing ? 'Edit Template' : 'Add Template' }}</h2>
        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="name">Name</label>
            <input id="name" v-model="form.name" type="text" required />
          </div>
          <div class="form-group">
            <label for="type">Type</label>
            <select id="type" v-model="form.type">
              <option v-for="t in templateTypes" :key="t.value" :value="t.value">{{ t.label }}</option>
            </select>
          </div>
          <div class="form-group">
            <label for="content">Content</label>
            <textarea id="content" v-model="form.content" rows="8" required></textarea>
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
import templateApi, { TEMPLATE_TYPES } from '@/api/template'
import type { Template } from '@/api/template'

const route = useRoute()
const clubId = Number(route.params.id)

const templateTypes = TEMPLATE_TYPES.map((t) => ({
  value: t,
  label: t.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase()),
}))

const templates = ref<Template[]>([])
const loading = ref(true)
const error = ref('')
const showModal = ref(false)
const editing = ref<Template | null>(null)
const submitting = ref(false)
const formError = ref('')

const form = ref({
  name: '',
  type: 'greeting',
  content: '',
})

function formatType(type: string): string {
  return type.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}

function truncate(str: string, len: number): string {
  if (!str) return ''
  return str.length > len ? str.slice(0, len) + '...' : str
}

async function loadTemplates() {
  loading.value = true
  error.value = ''
  try {
    const res = await templateApi.list(clubId)
    templates.value = res.data
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load templates'
  } finally {
    loading.value = false
  }
}

function openCreateForm() {
  editing.value = null
  form.value = { name: '', type: 'greeting', content: '' }
  formError.value = ''
  showModal.value = true
}

function openEditForm(tmpl: Template) {
  editing.value = tmpl
  form.value = { name: tmpl.name, type: tmpl.type, content: tmpl.content }
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
      await templateApi.update(editing.value.id, form.value)
    } else {
      await templateApi.create(clubId, form.value)
    }
    closeModal()
    await loadTemplates()
  } catch (e: any) {
    formError.value = e.response?.data?.error || 'Failed to save template'
  } finally {
    submitting.value = false
  }
}

async function handleDelete(tmpl: Template) {
  if (!confirm(`Delete template "${tmpl.name}"?`)) return
  try {
    await templateApi.remove(tmpl.id)
    await loadTemplates()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to delete template'
  }
}

onMounted(loadTemplates)
</script>
