<template>
  <div class="admin-users-view">
    <div class="page-header">
      <h1>Admin Users</h1>
      <button class="btn btn--primary" @click="openCreateForm">Add User</button>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error-message">{{ error }}</div>
    <div v-else>
      <table v-if="users.length > 0" class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Username</th>
            <th>Role</th>
            <th>Clubs</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id">
            <td>{{ user.id }}</td>
            <td>{{ user.username }}</td>
            <td><span class="role-badge" :class="`role-badge--${user.role}`">{{ user.role }}</span></td>
            <td>
              <div class="club-tags">
                <span v-for="club in userClubs[user.id] || []" :key="club.club_id" class="club-tag">
                  {{ club.club_name }}
                  <button class="club-tag__remove" @click="handleUnassignClub(user.id, club.club_id)" title="Remove">&times;</button>
                </span>
                <button class="btn btn--small" @click="openAssignForm(user)">+ Club</button>
              </div>
            </td>
            <td class="actions-cell">
              <button class="btn btn--small btn--danger" @click="handleDelete(user)" :disabled="user.id === authStore.userId">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">No admin users found.</div>
    </div>

    <!-- Create User Modal -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="closeCreateModal">
      <div class="modal">
        <h2>Create Admin User</h2>
        <form @submit.prevent="handleCreate">
          <div class="form-group">
            <label for="username">Username</label>
            <input id="username" v-model="createForm.username" type="text" required />
          </div>
          <div class="form-group">
            <label for="password">Password</label>
            <input id="password" v-model="createForm.password" type="password" required />
          </div>
          <div class="form-group">
            <label for="role">Role</label>
            <select id="role" v-model="createForm.role">
              <option value="admin">admin</option>
              <option value="superadmin">superadmin</option>
            </select>
          </div>
          <div v-if="formError" class="error-message">{{ formError }}</div>
          <div class="modal-actions">
            <button type="button" class="btn" @click="closeCreateModal">Cancel</button>
            <button type="submit" class="btn btn--primary" :disabled="submitting">
              {{ submitting ? 'Creating...' : 'Create' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Assign Club Modal -->
    <div v-if="showAssignModal" class="modal-overlay" @click.self="closeAssignModal">
      <div class="modal">
        <h2>Assign Club to {{ assignUser?.username }}</h2>
        <form @submit.prevent="handleAssign">
          <div class="form-group">
            <label for="club">Club</label>
            <select id="club" v-model.number="assignClubId" required>
              <option :value="0" disabled>Select a club</option>
              <option v-for="club in availableClubs" :key="club.id" :value="club.id">{{ club.name }}</option>
            </select>
          </div>
          <div v-if="formError" class="error-message">{{ formError }}</div>
          <div class="modal-actions">
            <button type="button" class="btn" @click="closeAssignModal">Cancel</button>
            <button type="submit" class="btn btn--primary" :disabled="submitting || assignClubId === 0">
              {{ submitting ? 'Assigning...' : 'Assign' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import adminUserApi from '@/api/adminuser'
import type { AdminUser, AdminUserClub } from '@/api/adminuser'
import clubApi from '@/api/club'
import type { Club } from '@/api/club'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const users = ref<AdminUser[]>([])
const userClubs = reactive<Record<number, AdminUserClub[]>>({})
const allClubs = ref<Club[]>([])
const loading = ref(true)
const error = ref('')

const showCreateModal = ref(false)
const showAssignModal = ref(false)
const submitting = ref(false)
const formError = ref('')

const createForm = ref({ username: '', password: '', role: 'admin' })
const assignUser = ref<AdminUser | null>(null)
const assignClubId = ref<number>(0)

const availableClubs = ref<Club[]>([])

async function loadUsers() {
  loading.value = true
  error.value = ''
  try {
    const res = await adminUserApi.list()
    users.value = res.data
    // Load clubs for each user
    for (const user of users.value) {
      try {
        const clubsRes = await adminUserApi.listClubs(user.id)
        userClubs[user.id] = clubsRes.data
      } catch {
        userClubs[user.id] = []
      }
    }
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to load admin users'
  } finally {
    loading.value = false
  }
}

async function loadAllClubs() {
  try {
    const res = await clubApi.list()
    allClubs.value = res.data
  } catch {
    // ignore
  }
}

function openCreateForm() {
  createForm.value = { username: '', password: '', role: 'admin' }
  formError.value = ''
  showCreateModal.value = true
}

function closeCreateModal() {
  showCreateModal.value = false
}

async function handleCreate() {
  formError.value = ''
  submitting.value = true
  try {
    await adminUserApi.create(createForm.value)
    closeCreateModal()
    await loadUsers()
  } catch (e: any) {
    formError.value = e.response?.data?.error || 'Failed to create admin user'
  } finally {
    submitting.value = false
  }
}

async function handleDelete(user: AdminUser) {
  if (user.id === authStore.userId) return
  if (!confirm(`Delete admin user "${user.username}"?`)) return
  try {
    await adminUserApi.remove(user.id)
    await loadUsers()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to delete admin user'
  }
}

function openAssignForm(user: AdminUser) {
  assignUser.value = user
  assignClubId.value = 0
  formError.value = ''
  // Filter out clubs already assigned
  const assigned = (userClubs[user.id] || []).map(c => c.club_id)
  availableClubs.value = allClubs.value.filter(c => !assigned.includes(c.id))
  showAssignModal.value = true
}

function closeAssignModal() {
  showAssignModal.value = false
  assignUser.value = null
}

async function handleAssign() {
  if (!assignUser.value || assignClubId.value === 0) return
  formError.value = ''
  submitting.value = true
  try {
    await adminUserApi.assignClub(assignUser.value.id, assignClubId.value)
    closeAssignModal()
    await loadUsers()
  } catch (e: any) {
    formError.value = e.response?.data?.error || 'Failed to assign club'
  } finally {
    submitting.value = false
  }
}

async function handleUnassignClub(userId: number, clubId: number) {
  try {
    await adminUserApi.unassignClub(userId, clubId)
    await loadUsers()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to unassign club'
  }
}

onMounted(async () => {
  await Promise.all([loadUsers(), loadAllClubs()])
})
</script>

<style scoped>
.role-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}
.role-badge--superadmin {
  background: #e8f5e9;
  color: #2e7d32;
}
.role-badge--admin {
  background: #e3f2fd;
  color: #1565c0;
}
.club-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  align-items: center;
}
.club-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  background: #f0f0f0;
  border-radius: 4px;
  font-size: 12px;
}
.club-tag__remove {
  background: none;
  border: none;
  color: #999;
  cursor: pointer;
  font-size: 14px;
  padding: 0;
  line-height: 1;
}
.club-tag__remove:hover {
  color: #e53935;
}
</style>
