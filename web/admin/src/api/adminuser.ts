import client from './client'

export interface AdminUser {
  id: number
  username: string
  role: string
  created_at: string
}

export interface AdminUserCreate {
  username: string
  password: string
  role: string
}

export interface AdminUserClub {
  club_id: number
  club_name: string
}

export default {
  list() {
    return client.get<AdminUser[]>('/admin-users')
  },

  create(data: AdminUserCreate) {
    return client.post<AdminUser>('/admin-users', data)
  },

  remove(id: number) {
    return client.delete(`/admin-users/${id}`)
  },

  listClubs(id: number) {
    return client.get<AdminUserClub[]>(`/admin-users/${id}/clubs`)
  },

  assignClub(id: number, clubId: number) {
    return client.post(`/admin-users/${id}/clubs/${clubId}`)
  },

  unassignClub(id: number, clubId: number) {
    return client.delete(`/admin-users/${id}/clubs/${clubId}`)
  },
}
