import { defineStore } from 'pinia'
import axios from 'axios'

// Configure axios interceptor for auth
axios.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Global response interceptor for auth errors
axios.interceptors.response.use(
  response => response,
  error => {
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const useAppStore = defineStore('app', {
  state: () => ({
    channels: [] as any[],
    tokens: [] as any[],
    logs: [] as any[],
    dashboard: {
      total_requests: 0,
      avg_latency: 0,
      total_tokens: 0,
      failed_requests: 0,
      provider_stats: [] as any[]
    },
    loading: false,
  }),
  actions: {
    async fetchDashboard() {
      try {
        const res = await axios.get('/admin/logs')
        const logs = res.data || []
        
        // Calculate real stats from logs
        let totalLatency = 0
        let totalTokens = 0
        let failedRequests = 0
        
        logs.forEach((log: any) => {
          totalLatency += log.latency_ms || 0
          totalTokens += log.total_tokens || 0
          if (log.status_code >= 400) failedRequests++
        })

        this.dashboard = {
          total_requests: logs.length,
          avg_latency: logs.length > 0 ? Math.round(totalLatency / logs.length) : 0,
          total_tokens: totalTokens,
          failed_requests: failedRequests,
          provider_stats: []
        }
      } catch (err) {
        console.error('Dashboard fetch failed', err)
      }
    },
    async fetchChannels() {
      try {
        const res = await axios.get('/admin/channels')
        this.channels = res.data || []
      } catch (err) {
        console.error('Failed to fetch channels', err)
      }
    },
    async fetchTokens() {
      try {
        const res = await axios.get('/admin/tokens')
        this.tokens = res.data || []
      } catch (err) {
        console.error('Failed to fetch tokens', err)
      }
    },
    async fetchLogs() {
      try {
        const res = await axios.get('/admin/logs')
        this.logs = res.data || []
        // Refresh dashboard stats when logs are fetched
        await this.fetchDashboard()
      } catch (err) {
        console.error('Failed to fetch logs', err)
      }
    },
    async deleteChannel(id: number) {
      await axios.delete(`/admin/channels/${id}`)
      await this.fetchChannels()
    },
    async deleteToken(id: number) {
      await axios.delete(`/admin/tokens/${id}`)
      await this.fetchTokens()
    }
  }
})
