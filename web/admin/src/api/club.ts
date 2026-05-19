import client from './client'

export interface Club {
  id: number
  name: string
  telegram_chat_id: number
  welcome_enabled: boolean
  birthday_enabled: boolean
  race_notify_enabled: boolean
  created_at?: string
  updated_at?: string
}

export interface ClubCreate {
  name: string
  telegram_chat_id: number
  welcome_enabled?: boolean
  birthday_enabled?: boolean
  race_notify_enabled?: boolean
}

export interface ClubUpdate {
  name?: string
  telegram_chat_id?: number
  welcome_enabled?: boolean
  birthday_enabled?: boolean
  race_notify_enabled?: boolean
}

export default {
  list() {
    return client.get<Club[]>('/clubs')
  },

  create(data: ClubCreate) {
    return client.post<Club>('/clubs', data)
  },

  get(id: number) {
    return client.get<Club>(`/clubs/${id}`)
  },

  update(id: number, data: ClubUpdate) {
    return client.put<Club>(`/clubs/${id}`, data)
  },

  remove(id: number) {
    return client.delete(`/clubs/${id}`)
  },
}
