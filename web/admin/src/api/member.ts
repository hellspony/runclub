import client from './client'

export interface Member {
  id: number
  fio: string
  telegram_username: string
  telegram_id: number
  birth_date?: string
  role: string
  created_at?: string
  updated_at?: string
}

export interface MemberCreate {
  fio: string
  telegram_username: string
  role: string
}

export interface MemberUpdate {
  fio?: string
  telegram_username?: string
  birth_date?: string
}

export default {
  list(clubId: number) {
    return client.get<Member[]>(`/clubs/${clubId}/members`)
  },

  create(clubId: number, data: MemberCreate) {
    return client.post<Member>(`/clubs/${clubId}/members`, data)
  },

  get(id: number) {
    return client.get<Member>(`/members/${id}`)
  },

  update(id: number, data: MemberUpdate) {
    return client.put<Member>(`/members/${id}`, data)
  },

  remove(id: number) {
    return client.delete(`/members/${id}`)
  },

  updateRole(clubId: number, memberId: number, role: string) {
    return client.put(`/clubs/${clubId}/members/${memberId}/role`, { role })
  },
}
