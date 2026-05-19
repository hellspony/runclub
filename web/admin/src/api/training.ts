import client from './client'

export interface Training {
  id: number
  club_id: number
  date: string
  location_id: number
  duration: number
  status: string
  created_at?: string
  updated_at?: string
}

export interface TrainingCreate {
  date: string
  location_id: number
  duration: number
  status: string
}

export interface TrainingUpdate {
  date?: string
  location_id?: number
  duration?: number
  status?: string
}

export interface Participant {
  id: number
  training_id: number
  member_id: number
  member_fio?: string
  created_at?: string
}

export default {
  list(clubId: number) {
    return client.get<Training[]>(`/clubs/${clubId}/trainings`)
  },

  create(clubId: number, data: TrainingCreate) {
    return client.post<Training>(`/clubs/${clubId}/trainings`, data)
  },

  get(id: number) {
    return client.get<Training>(`/trainings/${id}`)
  },

  update(id: number, data: TrainingUpdate) {
    return client.put<Training>(`/trainings/${id}`, data)
  },

  remove(id: number) {
    return client.delete(`/trainings/${id}`)
  },

  participants(trainingId: number) {
    return client.get<Participant[]>(`/trainings/${trainingId}/participants`)
  },
}
