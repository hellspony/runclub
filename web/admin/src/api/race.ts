import client from './client'

export interface Race {
  id: number
  club_id: number
  name: string
  date: string
  type: string
  place: string
  distances: string
  created_at?: string
  updated_at?: string
}

export interface RaceCreate {
  name: string
  date: string
  type: string
  place: string
  distances: string
}

export interface RaceUpdate {
  name?: string
  date?: string
  type?: string
  place?: string
  distances?: string
}

export interface Registration {
  id: number
  race_id: number
  member_id: number
  distance: string
  created_at?: string
}

export interface RegistrationCreate {
  member_id: number
  distance: string
}

export default {
  list(clubId: number) {
    return client.get<Race[]>(`/clubs/${clubId}/races`)
  },

  create(clubId: number, data: RaceCreate) {
    return client.post<Race>(`/clubs/${clubId}/races`, data)
  },

  get(id: number) {
    return client.get<Race>(`/races/${id}`)
  },

  update(id: number, data: RaceUpdate) {
    return client.put<Race>(`/races/${id}`, data)
  },

  remove(id: number) {
    return client.delete(`/races/${id}`)
  },

  registrations(raceId: number) {
    return client.get<Registration[]>(`/races/${raceId}/registrations`)
  },

  register(raceId: number, data: RegistrationCreate) {
    return client.post(`/races/${raceId}/registrations`, data)
  },

  unregister(raceId: number, data: { member_id: number }) {
    return client.delete(`/races/${raceId}/registrations`, { data })
  },
}
