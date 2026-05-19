import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import AppLayout from '@/components/AppLayout.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
  },
  {
    path: '/',
    component: AppLayout,
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('@/views/DashboardView.vue'),
      },
      {
        path: 'clubs',
        name: 'clubs',
        component: () => import('@/views/ClubsView.vue'),
      },
      {
        path: 'admin-users',
        name: 'admin-users',
        component: () => import('@/views/AdminUsersView.vue'),
      },
      {
        path: 'clubs/:id',
        name: 'club-detail',
        component: () => import('@/views/ClubDetailView.vue'),
        children: [
          {
            path: 'members',
            name: 'club-members',
            component: () => import('@/views/MembersView.vue'),
          },
          {
            path: 'locations',
            name: 'club-locations',
            component: () => import('@/views/LocationsView.vue'),
          },
          {
            path: 'races',
            name: 'club-races',
            component: () => import('@/views/RacesView.vue'),
          },
          {
            path: 'templates',
            name: 'club-templates',
            component: () => import('@/views/TemplatesView.vue'),
          },
          {
            path: 'trainings',
            name: 'club-trainings',
            component: () => import('@/views/TrainingsView.vue'),
          },
          {
            path: 'joint-runs',
            name: 'club-joint-runs',
            component: () => import('@/views/JointRunsView.vue'),
          },
        ],
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  if (to.path !== '/login' && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
