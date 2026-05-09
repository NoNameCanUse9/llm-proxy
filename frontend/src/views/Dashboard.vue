<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Activity, Server, Zap, ShieldCheck } from 'lucide-vue-next'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const store = useAppStore()

onMounted(() => {
  store.fetchDashboard()
  store.fetchLogs()
})
</script>

<template>
  <div class="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
    <div>
      <h2 class="text-3xl font-bold tracking-tight text-white">
        {{ t('dashboard') }}
      </h2>
      <p class="text-zinc-400 mt-2">{{ t('real_time_desc') }}</p>
    </div>

    <!-- Stats Grid -->
    <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
      <!-- Total Requests -->
      <div class="group relative overflow-hidden rounded-2xl border border-white/5 bg-zinc-900/40 p-6 backdrop-blur-xl transition-all hover:border-white/20">
        <div class="flex items-center justify-between relative z-10">
          <p class="text-sm font-medium text-zinc-400">{{ t('total_requests') }}</p>
          <Activity class="h-5 w-5 text-cyan-400" />
        </div>
        <div class="mt-4 relative z-10">
          <h3 class="text-3xl font-bold font-mono tracking-tight text-white">{{ store.dashboard.total_requests }}</h3>
        </div>
      </div>

      <!-- Avg Latency -->
      <div class="group relative overflow-hidden rounded-2xl border border-white/5 bg-zinc-900/40 p-6 backdrop-blur-xl transition-all hover:border-white/20">
        <div class="flex items-center justify-between relative z-10">
          <p class="text-sm font-medium text-zinc-400">{{ t('avg_latency') }}</p>
          <Zap class="h-5 w-5 text-purple-400" />
        </div>
        <div class="mt-4 relative z-10">
          <h3 class="text-3xl font-bold font-mono tracking-tight text-white">{{ store.dashboard.avg_latency }}ms</h3>
        </div>
      </div>

      <!-- Total Tokens -->
      <div class="group relative overflow-hidden rounded-2xl border border-white/5 bg-zinc-900/40 p-6 backdrop-blur-xl transition-all hover:border-white/20">
        <div class="flex items-center justify-between relative z-10">
          <p class="text-sm font-medium text-zinc-400">{{ t('total_tokens') }}</p>
          <ShieldCheck class="h-5 w-5 text-rose-400" />
        </div>
        <div class="mt-4 relative z-10">
          <h3 class="text-3xl font-bold font-mono tracking-tight text-white">{{ store.dashboard.total_tokens }}</h3>
        </div>
      </div>

      <!-- Failed Requests -->
      <div class="group relative overflow-hidden rounded-2xl border border-white/5 bg-zinc-900/40 p-6 backdrop-blur-xl transition-all hover:border-white/20">
        <div class="flex items-center justify-between relative z-10">
          <p class="text-sm font-medium text-zinc-400">{{ t('failed_requests') }}</p>
          <Server class="h-5 w-5 text-emerald-400" />
        </div>
        <div class="mt-4 relative z-10">
          <h3 class="text-3xl font-bold font-mono tracking-tight text-white">{{ store.dashboard.failed_requests }}</h3>
        </div>
      </div>
    </div>

    <!-- Activity Section -->
    <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-7">
      <div class="lg:col-span-4 rounded-2xl border border-white/5 bg-zinc-900/40 p-6 backdrop-blur-xl">
        <h3 class="text-lg font-medium mb-4 text-zinc-200">{{ t('provider_perf') }}</h3>
        <div class="h-[300px] flex items-center justify-center border border-dashed border-white/10 rounded-xl bg-black/20">
          <p class="text-zinc-500 font-mono text-sm">Visualizing performance data...</p>
        </div>
      </div>
      <div class="lg:col-span-3 rounded-2xl border border-white/5 bg-zinc-900/40 p-6 backdrop-blur-xl">
        <h3 class="text-lg font-medium mb-4 text-zinc-200">{{ t('recent_activity') }}</h3>
        <div class="space-y-4">
          <div v-for="log in store.logs.slice(0, 8)" :key="log.id" class="flex items-center gap-4 text-sm font-mono border-b border-white/5 pb-4 last:border-0 hover:bg-white/[0.02] transition-colors rounded-lg px-2 -mx-2">
            <span class="text-zinc-500 text-xs">{{ new Date(log.created_at).toLocaleTimeString() }}</span>
            <span :class="['font-bold', log.status_code < 400 ? 'text-emerald-500' : 'text-rose-500']">{{ log.status_code }}</span>
            <span class="text-zinc-300 truncate max-w-[100px]">{{ log.model }}</span>
            <span class="text-zinc-500 ml-auto text-xs">{{ log.latency_ms }}ms</span>
          </div>
          <div v-if="store.logs.length === 0" class="text-center py-10 text-zinc-600 italic text-sm">
            {{ t('no_logs') }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
