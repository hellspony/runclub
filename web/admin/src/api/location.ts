import client from './client'

export interface Location {
  id: number
  club_id: number
  name: string
  address: string
  map_url: string
  created_at?: string
  updated_at?: string
}

export interface LocationCreate {
  name: string
  address: string
  map_url: string
}

export interface LocationUpdate {
  name?: string
  address?: string
  map_url?: string
}

export default {
  list(clubId: number) {
    return client.get<Location[]>(`/clubs/${clubId}/locations`)
  },

  create(clubId: number, data: LocationCreate) {
    return client.post<Location>(`/clubs/${clubId}/locations`, data)
  },

  get(id: number) {
    return client.get<Location>(`/locations/${id}`)
  },

  update(id: number, data: LocationUpdate) {
    return client.put<Location>(`/locations/${id}`, data)
  },

  remove(id: number) {
    return client.delete(`/locations/${id}`)
  },
}
