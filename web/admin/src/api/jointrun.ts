import client from './client'

export interface JointRun {
  id: number
  club_id: number
  date: string
  location_id: number
  creator_id: number
  created_at?: string
  updated_at?: string
}

export interface JointRunCreate {
  date: string
  location_id: number
  creator_id: number
}

export interface JointRunParticipant {
  id: number
  joint_run_id: number
  member_id: number
  member_fio?: string
  created_at?: string
}

export default {
  list(clubId: number) {
    return client.get<JointRun[]>(`/clubs/${clubId}/joint-runs`)
  },

  create(clubId: number, data: JointRunCreate) {
    return client.post<JointRun>(`/clubs/${clubId}/joint-runs`, data)
  },

  get(id: number) {
    return client.get<JointRun>(`/joint-runs/${id}`)
  },

  remove(id: number) {
    return client.delete(`/joint-runs/${id}`)
  },

  participants(runId: number) {
    return client.get<JointRunParticipant[]>(`/joint-runs/${runId}/participants`)
  },
}
