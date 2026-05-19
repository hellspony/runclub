<template>
  <div class="members-view">
    <div class="page-header">
      <h2>Members</h2>
      <button class="btn btn--primary" @click="openCreateForm">Add Member</button>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error-message">{{ error }}</div>
    <div v-else>
      <table v-if="members.length > 0" class="data-table">
        <thead>
          <tr>
            <th>FIO</th>
            <th>Username</th>
            <th>Role</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="member in members" :key="member.id">
            <td>{{ member.fio }}</td>
            <td>{{ member.telegram_username }}</td>
            <td>
              <select
                :value="member.role"
                class="role-select"
                @change="handleRoleChange(member, ($event.target as HTMLSelectElement).value)"
              >
                <option value="admin">Admin</option>
                <option value="trainer">Trainer</option>
                <option value="member">Member</option>
              </select>
            </td>
            <td class="actions-cell">
              <button class="btn btn--small" @click="openEditForm(member)">Edit</button>
              <button class="btn btn--small btn--danger" @click="handleDelete(member)">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">No members found.</div>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal">
        <h2>{{ editing ? 'Edit Member' : 'Add Member' }}</h2>
        <form @submit.prevent="handleSubmit">
          <div class="form-group">
            <label for="fio">FIO</label>
            <input id="fio" v-model="form.fio" type="text" required />
          </div>
          <div class="form-group">
            <label for="username">Username</label>
            <input id="username" v-model="form.telegram_username" type="text" required />
          </div>
          <div class="form-group">
            <label for="role">Role</label>
            <select id="role" v-model="form.role">
              <option value="admin">Admin</option>
              <option value="trainer">Trainer</option>
              <option value="member">Member</option>
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import memberApi from '@/api/member'
import type { Member } from '@/api/member'

const route = useRoute()
const clubId = Number(route.params.id)

const members = ref<Member[]>([])
const loading = ref(true)
const error = ref('')
const showModal = ref(false)
const editing = ref<Member | null>(null)
const submitting = ref(false)
const formError = ref('')

const form = ref({
  fio: '',
  telegram_username: '',
  role: 'member',
})

async function loadMembers() {
  loading.value = true
  error.value = ''
  try {
    const res = await memberApi.list(clubId)
    members.value = res.data
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load members'
  } finally {
    loading.value = false
  }
}

function openCreateForm() {
  editing.value = null
  form.value = { fio: '', telegram_username: '', role: 'member' }
  formError.value = ''
  showModal.value = true
}

function openEditForm(member: Member) {
  editing.value = member
  form.value = { fio: member.fio, telegram_username: member.telegram_username, role: member.role }
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
      await memberApi.update(editing.value.id, form.value)
    } else {
      await memberApi.create(clubId, form.value)
    }
    closeModal()
    await loadMembers()
  } catch (e: any) {
    formError.value = e.response?.data?.error || 'Failed to save member'
  } finally {
    submitting.value = false
  }
}

async function handleRoleChange(member: Member, newRole: string) {
  try {
    await memberApi.updateRole(clubId, member.id, newRole)
    member.role = newRole
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to update role'
  }
}

async function handleDelete(member: Member) {
  if (!confirm(`Delete member "${member.fio}"?`)) return
  try {
    await memberApi.remove(member.id)
    await loadMembers()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to delete member'
  }
}

onMounted(loadMembers)
</script>
