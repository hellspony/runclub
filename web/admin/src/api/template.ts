import client from './client'

export interface Template {
  id: number
  club_id: number
  name: string
  type: string
  content: string
  created_at?: string
  updated_at?: string
}

export interface TemplateCreate {
  name: string
  type: string
  content: string
}

export interface TemplateUpdate {
  name?: string
  type?: string
  content?: string
}

export const TEMPLATE_TYPES = [
  'greeting',
  'training_reminder',
  'race_announcement',
  'results_summary',
  'joint_run_invite',
  'weekly_report',
] as const

export type TemplateType = (typeof TEMPLATE_TYPES)[number]

export default {
  list(clubId: number) {
    return client.get<Template[]>(`/clubs/${clubId}/templates`)
  },

  create(clubId: number, data: TemplateCreate) {
    return client.post<Template>(`/clubs/${clubId}/templates`, data)
  },

  update(id: number, data: TemplateUpdate) {
    return client.put<Template>(`/templates/${id}`, data)
  },

  remove(id: number) {
    return client.delete(`/templates/${id}`)
  },
}
