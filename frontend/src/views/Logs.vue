<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { 
  Terminal, Search, Shield, Cpu, Activity, Globe, Zap, Hash, 
  ChevronUp, ChevronDown, Filter, X 
} from 'lucide-vue-next'
import { Input } from '@/components/ui/input'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const store = useAppStore()

// State
const searchQuery = ref('')
const sortOrder = ref<'asc' | 'desc'>('desc')
const filters = ref({
  provider: '',
  model: '',
  ip_address: '',
  key_hint: ''
})

onMounted(() => {
  store.fetchLogs()
})

// Dynamic Filter Options (extracted from current logs)
const filterOptions = computed(() => {
  const providers = new Set<string>()
  const models = new Set<string>()
  const ips = new Set<string>()
  const keys = new Set<string>()

  store.logs.forEach(log => {
    if (log.provider) providers.add(log.provider)
    if (log.model) models.add(log.model)
    if (log.ip_address) ips.add(log.ip_address)
    if (log.key_hint) keys.add(log.key_hint)
  })

  return {
    providers: Array.from(providers).sort(),
    models: Array.from(models).sort(),
    ips: Array.from(ips).sort(),
    keys: Array.from(keys).sort()
  }
})

// Filtered and Sorted Logs
const filteredLogs = computed(() => {
  let result = [...store.logs]

  // Apply Search
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    result = result.filter(log => 
      log.model?.toLowerCase().includes(q) || 
      log.provider?.toLowerCase().includes(q) ||
      log.ip_address?.includes(q) ||
      log.request_id?.includes(q)
    )
  }

  // Apply Specific Filters
  if (filters.value.provider) result = result.filter(l => l.provider === filters.value.provider)
  if (filters.value.model) result = result.filter(l => l.model === filters.value.model)
  if (filters.value.ip_address) result = result.filter(l => l.ip_address === filters.value.ip_address)
  if (filters.value.key_hint) result = result.filter(l => l.key_hint === filters.value.key_hint)

  // Apply Sorting
  result.sort((a, b) => {
    const timeA = new Date(a.created_at).getTime()
    const timeB = new Date(b.created_at).getTime()
    return sortOrder.value === 'asc' ? timeA - timeB : timeB - timeA
  })

  return result
})

const toggleSort = () => {
  sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
}

const clearFilters = () => {
  filters.value = { provider: '', model: '', ip_address: '', key_hint: '' }
  searchQuery.value = ''
}

const getStatusColor = (code: number) => {
  if (code < 300) return 'text-emerald-400 bg-emerald-400/10'
  if (code < 500) return 'text-amber-400 bg-amber-400/10'
  return 'text-rose-400 bg-rose-400/10'
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-3xl font-bold tracking-tight text-white">{{ t('logs_title') }}</h2>
        <p class="text-zinc-400 mt-2">{{ t('logs_desc') }}</p>
      </div>
      <div class="flex items-center gap-4">
        <button 
          v-if="Object.values(filters).some(v => v !== '') || searchQuery"
          @click="clearFilters"
          class="flex items-center gap-2 text-xs text-zinc-500 hover:text-white transition-colors"
        >
          <X class="w-3 h-3" />
          {{ t('reset') }}
        </button>
        <div class="relative w-72">
          <Search class="absolute left-3 top-2.5 h-4 w-4 text-zinc-500" />
          <Input 
            v-model="searchQuery"
            class="pl-10 bg-zinc-900/50 border-white/10 focus:ring-cyan-500/30" 
            :placeholder="t('search_placeholder')" 
          />
        </div>
      </div>
    </div>

    <!-- Enhanced Terminal View -->
    <div class="rounded-2xl border border-white/5 bg-zinc-950/80 backdrop-blur-xl overflow-hidden shadow-2xl">
      <!-- Header -->
      <div class="flex items-center justify-between px-4 py-3 bg-white/[0.03] border-b border-white/5">
        <div class="flex items-center gap-2">
          <Terminal class="w-4 h-4 text-cyan-500" />
          <span class="text-zinc-400 text-[10px] tracking-[0.2em] font-bold uppercase">System Audit Stream</span>
        </div>
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-1.5">
            <div class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></div>
            <span class="text-[10px] text-zinc-500 uppercase font-mono">Live</span>
          </div>
        </div>
      </div>

      <!-- Logs Container -->
      <div class="p-2 overflow-x-auto">
        <table class="w-full text-left border-separate border-spacing-y-1">
          <thead>
            <tr class="text-[10px] text-zinc-500 uppercase font-mono">
              <!-- Sortable Timestamp -->
              <th class="px-3 py-2 font-medium cursor-pointer hover:text-white transition-colors" @click="toggleSort">
                <div class="flex items-center gap-1">
                  {{ t('timestamp') }}
                  <ChevronUp v-if="sortOrder === 'asc'" class="w-3 h-3" />
                  <ChevronDown v-else class="w-3 h-3" />
                </div>
              </th>
              
              <th class="px-3 py-2 font-medium text-center">{{ t('status') }}</th>

              <!-- Filterable Provider -->
              <th class="px-3 py-2 font-medium">
                <div class="flex items-center gap-2">
                  {{ t('channel') }}
                  <select v-model="filters.provider" class="bg-transparent border-none text-zinc-500 focus:ring-0 cursor-pointer hover:text-white p-0 text-[10px] uppercase max-w-[80px]">
                    <option value="">ALL</option>
                    <option v-for="p in filterOptions.providers" :key="p" :value="p">{{ p }}</option>
                  </select>
                </div>
              </th>

              <!-- Filterable Model -->
              <th class="px-3 py-2 font-medium">
                <div class="flex items-center gap-2">
                  {{ t('models') }}
                  <select v-model="filters.model" class="bg-transparent border-none text-zinc-500 focus:ring-0 cursor-pointer hover:text-white p-0 text-[10px] uppercase max-w-[120px]">
                    <option value="">ALL</option>
                    <option v-for="m in filterOptions.models" :key="m" :value="m">{{ m }}</option>
                  </select>
                </div>
              </th>

              <th class="px-3 py-2 font-medium text-right">{{ t('latency') }}</th>
              <th class="px-3 py-2 font-medium text-right">{{ t('log_usage') }}</th>

              <!-- Filterable Key -->
              <th class="px-3 py-2 font-medium">
                <div class="flex items-center gap-2">
                  {{ t('log_key') }}
                  <select v-model="filters.key_hint" class="bg-transparent border-none text-zinc-500 focus:ring-0 cursor-pointer hover:text-white p-0 text-[10px] uppercase max-w-[80px]">
                    <option value="">ALL</option>
                    <option v-for="k in filterOptions.keys" :key="k" :value="k">{{ k }}</option>
                  </select>
                </div>
              </th>

              <!-- Filterable IP -->
              <th class="px-3 py-2 font-medium">
                <div class="flex items-center gap-2">
                  {{ t('ip_address') }}
                  <select v-model="filters.ip_address" class="bg-transparent border-none text-zinc-500 focus:ring-0 cursor-pointer hover:text-white p-0 text-[10px] uppercase max-w-[100px]">
                    <option value="">ALL</option>
                    <option v-for="ip in filterOptions.ips" :key="ip" :value="ip">{{ ip }}</option>
                  </select>
                </div>
              </th>
            </tr>
          </thead>
          <tbody class="font-mono text-xs">
            <tr v-for="log in filteredLogs" :key="log.id" class="group hover:bg-white/[0.03] transition-colors rounded-lg overflow-hidden">
              <td class="px-3 py-2 text-zinc-500 whitespace-nowrap">
                {{ new Date(log.created_at).toLocaleTimeString() }}
              </td>
              <td class="px-3 py-2 whitespace-nowrap text-center">
                <span :class="['px-2 py-0.5 rounded text-[10px] font-bold', getStatusColor(log.status_code)]">
                  {{ log.status_code }}
                </span>
              </td>
              <td class="px-3 py-2 whitespace-nowrap">
                <div class="flex items-center gap-1.5 text-zinc-300">
                  <Shield class="w-3 h-3 text-cyan-500/50" />
                  {{ log.provider || 'system' }}
                </div>
              </td>
              <td class="px-3 py-2 whitespace-nowrap">
                <div class="flex items-center gap-1.5 text-zinc-400">
                  <Cpu class="w-3 h-3 text-zinc-600" />
                  <span class="truncate max-w-[200px] group-hover:text-white transition-colors">{{ log.model }}</span>
                </div>
              </td>
              <td class="px-3 py-2 text-right whitespace-nowrap">
                <div class="flex items-center justify-end gap-1 text-zinc-400">
                  <Zap class="w-3 h-3" :class="log.latency_ms > 5000 ? 'text-amber-500' : 'text-purple-500'" />
                  {{ log.latency_ms }}ms
                </div>
              </td>
              <td class="px-3 py-2 text-right whitespace-nowrap">
                <div class="flex items-center justify-end gap-1 text-emerald-500/80">
                  <Activity class="w-3 h-3" />
                  {{ log.total_tokens }}
                </div>
              </td>
              <td class="px-3 py-2 whitespace-nowrap">
                <div class="flex items-center gap-1 text-zinc-500 text-[10px]">
                  <Hash class="w-2.5 h-2.5 opacity-50" />
                  {{ log.key_hint || 'N/A' }}
                </div>
              </td>
              <td class="px-3 py-2 whitespace-nowrap">
                <div class="flex items-center gap-1.5 text-zinc-600 text-[10px]">
                  <Globe class="w-2.5 h-2.5 opacity-50" />
                  {{ log.ip_address }}
                </div>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Empty State -->
        <div v-if="filteredLogs.length === 0" class="py-20 flex flex-col items-center justify-center space-y-4 text-zinc-600">
          <Terminal class="w-12 h-12 opacity-10" />
          <p class="italic text-sm font-mono">{{ t('no_logs') }}</p>
          <button @click="clearFilters" class="text-xs text-cyan-500 underline decoration-cyan-500/30 underline-offset-4">{{ t('reset') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Reset select styles for minimal terminal look */
select {
  appearance: none;
  background-image: none;
  outline: none;
}
select option {
  background-color: #09090b;
  color: #a1a1aa;
}
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 10px;
}
</style>
