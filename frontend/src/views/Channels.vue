<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import axios from 'axios'
import { Plus, Pencil, Trash2, Loader2, RefreshCw, Search, Copy, Check, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetFooter,
} from '@/components/ui/sheet'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { useAppStore } from '@/stores/app'
import { useAppClipboard } from '@/composables/useAppClipboard'

const { t } = useI18n()
const store = useAppStore()

const isOpen = ref(false)
const isSaving = ref(false)
const isFetchingModels = ref(false)
const editingId = ref<number | null>(null)
const searchQuery = ref('')

const filteredChannels = computed(() => {
  if (!searchQuery.value) return store.channels
  const q = searchQuery.value.toLowerCase()
  return store.channels.filter(c => 
    c.name.toLowerCase().includes(q) || 
    c.allowed_models?.toLowerCase().includes(q)
  )
})

const highlightedId = ref<number | null>(null)
const showSuggestions = ref(false)

const searchSuggestions = computed(() => {
  if (!searchQuery.value || searchQuery.value.length < 1) return []
  const q = searchQuery.value.toLowerCase()
  const suggestions: any[] = []
  
  store.channels.forEach(channel => {
    // Match channel name
    if (channel.name.toLowerCase().includes(q)) {
      suggestions.push({ type: 'channel', text: channel.name, channelId: channel.id })
    }
    // Match models
    if (channel.allowed_models) {
      const models = channel.allowed_models.split(',')
      models.forEach(m => {
        if (m.trim().toLowerCase().includes(q)) {
          suggestions.push({ type: 'model', text: m.trim(), channelId: channel.id, channelName: channel.name })
        }
      })
    }
  })
  
  return suggestions.slice(0, 8) // Limit to 8 suggestions
})

const handleSelectSuggestion = (suggestion: any) => {
  searchQuery.value = suggestion.text
  showSuggestions.value = false
  scrollToChannel(suggestion.channelId)
}

const handleBlur = () => {
  setTimeout(() => {
    showSuggestions.value = false
  }, 200)
}

const scrollToChannel = (id: number) => {
  const element = document.getElementById(`channel-${id}`)
  if (element) {
    element.scrollIntoView({ behavior: 'smooth', block: 'center' })
    highlightedId.value = id
    setTimeout(() => {
      highlightedId.value = null
    }, 3000)
  }
}

const form = ref({
  name: '',
  type: 'openai',
  base_url: '',
  api_keys: [] as string[],
  api_key_input: '', // Temporary input field for the UI
  allowed_models: '',
  denied_models: '',
  is_active: true,
  rpm: 0
})

onMounted(() => {
  store.fetchChannels()
})

const openAdd = () => {
  editingId.value = null
  form.value = {
    name: '',
    type: 'openai',
    base_url: '',
    api_keys: [],
    api_key_input: '',
    allowed_models: '',
    denied_models: '',
    is_active: true,
    rpm: 0
  }
  isOpen.value = true
}

const openEdit = (channel: any) => {
  editingId.value = channel.id
  // Map APIKeys objects from backend to simple string array for the form
  const keys = channel.api_keys ? channel.api_keys.map((k: any) => k.key_value) : []
  form.value = { 
    ...channel, 
    api_keys: keys,
    api_key_input: keys.join('\n')
  }
  isOpen.value = true
}

const { copy: copyToClipboard, copied: isCopied } = useAppClipboard()

const handleSave = async () => {
  isSaving.value = true
  try {
    const data = { ...form.value }
    delete data.api_key_input // Clean up frontend-only field
    if (form.value.api_key_input) {
      // Split by newline or space, filter out empty strings
      data.api_keys = form.value.api_key_input.split(/[\n\s]+/).map(k => k.trim()).filter(Boolean)
    }
    
    if (editingId.value) {
      await axios.put(`/admin/channels/${editingId.value}`, data)
    } else {
      await axios.post('/admin/channels', data)
    }
    await store.fetchChannels()
    isOpen.value = false
  } catch (err) {
    console.error('Save failed', err)
  } finally {
    isSaving.value = false
  }
}

const fetchModels = async () => {
  if (!form.value.base_url || !form.value.api_key_input) {
    alert('Please provide Base URL and API Key first')
    return
  }
  isFetchingModels.value = true
  try {
    const firstKey = form.value.api_key_input.split(/[\n\s]+/)[0]
    const res = await axios.post('/admin/channels/fetch-models', {
      type: form.value.type,
      base_url: form.value.base_url,
      api_key: firstKey
    })
    if (res.data && res.data.models) {
      form.value.allowed_models = res.data.models
    } else if (Array.isArray(res.data)) {
      form.value.allowed_models = res.data.join(',')
    }
  } catch (err) {
    console.error('Fetch models failed', err)
  } finally {
    isFetchingModels.value = false
  }
}

const removeModel = (model: string) => {
  const models = form.value.allowed_models.split(',').filter(Boolean)
  form.value.allowed_models = models.filter(m => m !== model).join(',')
}

const removeDeniedModel = (model: string) => {
  const models = form.value.denied_models.split(',').filter(Boolean)
  form.value.denied_models = models.filter(m => m !== model).join(',')
}

const handleDelete = async (id: number) => {
  if (confirm('Are you sure?')) {
    await store.deleteChannel(id)
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-10">
      <div class="space-y-2">
        <h1 class="text-4xl font-bold tracking-tight bg-gradient-to-r from-white to-white/60 bg-clip-text text-transparent">
          {{ t('providers') }}
        </h1>
        <p class="text-zinc-500 text-sm font-medium">
          {{ t('manage_providers_desc') }}
        </p>
      </div>
      
      <div class="flex items-center gap-4">
        <div class="relative group">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500 group-focus-within:text-cyan-400 transition-colors" />
          <Input 
            v-model="searchQuery" 
            :placeholder="t('search_placeholder')" 
            class="pl-10 w-[300px] bg-zinc-950 border-white/10 text-white focus-visible:ring-cyan-500/50 h-10 rounded-xl transition-all"
            @focus="showSuggestions = true"
            @blur="handleBlur"
          />
          
          <!-- Google-style Autocomplete Dropdown -->
          <transition name="fade">
            <div v-if="showSuggestions && searchSuggestions.length > 0" class="absolute top-full left-0 right-0 mt-2 bg-[#14171f] border border-white/10 rounded-xl shadow-2xl z-[60] overflow-hidden backdrop-blur-xl">
              <div 
                v-for="(s, i) in searchSuggestions" 
                :key="i"
                class="px-4 py-3 hover:bg-white/5 cursor-pointer border-b border-white/5 last:border-0 flex items-center justify-between group/item transition-colors"
                @mousedown="handleSelectSuggestion(s)"
              >
                <div class="flex items-center gap-3">
                  <div :class="['w-1.5 h-1.5 rounded-full', s.type === 'channel' ? 'bg-purple-500' : 'bg-cyan-500']"></div>
                  <span class="text-sm font-medium text-zinc-300 group-hover/item:text-white transition-colors">{{ s.text }}</span>
                </div>
                <span v-if="s.type === 'model'" class="text-[10px] text-zinc-500 font-mono italic">
                  in {{ s.channelName }}
                </span>
                <span v-else class="text-[10px] text-zinc-500 uppercase tracking-wider font-semibold">
                  Channel
                </span>
              </div>
            </div>
          </transition>
        </div>
        <Button @click="openAdd" class="px-6 h-10 rounded-xl">
          <Plus class="w-4 h-4 mr-2" />
          {{ t('add_provider') }}
        </Button>
      </div>
    </div>

    <div class="rounded-2xl border border-white/5 bg-zinc-900/40 backdrop-blur-xl overflow-hidden">
      <Table>
        <TableHeader class="bg-white/5">
          <TableRow class="hover:bg-transparent border-white/5">
            <TableHead class="text-zinc-400 font-medium">{{ t('name') }}</TableHead>
            <TableHead class="text-zinc-400 font-medium">{{ t('channel_type') }}</TableHead>
            <TableHead class="text-zinc-400 font-medium">{{ t('status') }}</TableHead>
            <TableHead class="text-zinc-400 font-medium">{{ t('models') }}</TableHead>
            <TableHead class="text-right text-zinc-400 font-medium">{{ t('actions') }}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow 
            v-for="channel in filteredChannels" 
            :key="channel.id" 
            :id="`channel-${channel.id}`"
            class="group border-white/5 hover:bg-white/[0.02] transition-colors relative h-[96px]"
            :class="{ 'pulse-highlight': highlightedId === channel.id }"
          >
            <TableCell class="font-medium text-white">{{ channel.name }}</TableCell>
            <TableCell>
              <span class="px-2 py-0.5 rounded-md bg-white/5 text-[10px] font-mono text-zinc-400 uppercase border border-white/5">{{ channel.type }}</span>
            </TableCell>
            <TableCell>
              <div class="flex items-center gap-2">
                <div :class="['w-1.5 h-1.5 rounded-full', channel.is_active ? 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]' : 'bg-zinc-600']"></div>
                <span class="text-xs text-zinc-400">{{ channel.is_active ? t('active') : t('inactive') }}</span>
              </div>
            </TableCell>
            <TableCell>
              <div class="flex flex-wrap gap-1 max-w-[500px] max-h-[64px] overflow-hidden py-1">
                <template v-if="searchQuery">
                  <!-- Highlighted matching models first -->
                  <span 
                    v-for="m in channel.allowed_models?.split(',').filter(m => m.trim().toLowerCase().includes(searchQuery.toLowerCase()))" 
                    :key="'match-' + m" 
                    @click="copyToClipboard(channel.name + '/' + m.trim())"
                    class="px-1.5 py-0.5 rounded border border-cyan-500/30 bg-cyan-500/10 text-[9px] text-cyan-400 whitespace-nowrap shadow-[0_0_8px_rgba(6,182,212,0.2)] cursor-pointer hover:bg-cyan-500/20 hover:border-cyan-500/50 transition-all active:scale-95"
                    title="Click to copy: channel/model"
                  >
                    {{ m }}
                  </span>
                  <!-- Other models (if search is partial) -->
                  <span 
                    v-for="m in channel.allowed_models?.split(',').filter(m => !m.trim().toLowerCase().includes(searchQuery.toLowerCase())).slice(0, 5)" 
                    :key="'other-' + m" 
                    @click="copyToClipboard(channel.name + '/' + m.trim())"
                    class="px-1.5 py-0.5 rounded border border-white/5 bg-white/5 text-[9px] text-zinc-500 whitespace-nowrap opacity-50 cursor-pointer hover:opacity-100 hover:bg-white/10 transition-all active:scale-95"
                    title="Click to copy: channel/model"
                  >
                    {{ m }}
                  </span>
                </template>
                <template v-else>
                  <span 
                    v-for="m in channel.allowed_models?.split(',').filter(Boolean).slice(0, 10)" 
                    :key="m" 
                    @click="copyToClipboard(channel.name + '/' + m.trim())"
                    class="px-1.5 py-0.5 rounded border border-white/5 bg-white/5 text-[9px] text-zinc-400 whitespace-nowrap cursor-pointer hover:bg-white/10 hover:text-white transition-all active:scale-95"
                    title="Click to copy: channel/model"
                  >
                    {{ m }}
                  </span>
                </template>
                
                <span v-if="!searchQuery && channel.allowed_models?.split(',').filter(Boolean).length > 10" class="text-[9px] text-zinc-600 flex items-center">
                  +{{ channel.allowed_models.split(',').filter(Boolean).length - 10 }}
                </span>
              </div>
            </TableCell>
            <TableCell class="text-right">
              <div class="flex justify-end gap-1">
                <Button 
                  v-if="channel.api_keys && channel.api_keys.length > 0"
                  variant="ghost" 
                  size="icon" 
                  @click="copyToClipboard(channel.api_keys[0].key_value, 'table_key')" 
                  class="h-8 w-8 text-zinc-500 hover:text-cyan-400 hover:bg-cyan-400/10"
                  title="Copy API Key"
                >
                  <Check v-if="isCopied === 'table_key'" class="h-3.5 w-3.5 text-emerald-400" />
                  <Copy v-else class="h-3.5 w-3.5" />
                </Button>
                <Button variant="ghost" size="icon" @click="openEdit(channel)" class="h-8 w-8 text-zinc-500 hover:text-white hover:bg-white/5">
                  <Pencil class="w-3.5 h-3.5" />
                </Button>
                <Button variant="ghost" size="icon" @click="handleDelete(channel.id)" class="h-8 w-8 text-zinc-500 hover:text-rose-500 hover:bg-rose-500/10">
                  <Trash2 class="w-3.5 h-3.5" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <!-- Sheet -->
    <Sheet v-model:open="isOpen">
      <SheetContent side="right" class="w-full sm:max-w-2xl bg-[#0a0c10] border-l border-white/10 text-white shadow-2xl overflow-y-auto px-10 z-[100]">
        <SheetHeader class="pb-6 border-b border-white/5">
          <SheetTitle class="text-xl font-bold text-white">{{ editingId ? t('edit_channel') : t('add_channel') }}</SheetTitle>
          <SheetDescription class="text-zinc-500 text-sm">
            {{ t('channels_desc') }}
          </SheetDescription>
        </SheetHeader>

        <div class="grid gap-8 py-10">
          <div class="flex items-center justify-between">
            <Label class="text-zinc-400 text-xs uppercase font-semibold">{{ t('status') }}</Label>
            <Switch v-model="form.is_active" />
          </div>

          <div class="grid gap-2">
            <Label class="text-zinc-400 text-xs uppercase font-semibold">{{ t('channel_name') }}</Label>
            <Input v-model="form.name" class="bg-white/5 border-white/10 h-10" />
          </div>

          <div class="grid gap-2">
            <Label class="text-zinc-400 text-xs uppercase font-semibold">{{ t('channel_type') }}</Label>
            <Select v-model="form.type">
              <SelectTrigger class="bg-white/5 border-white/10 h-10">
                <SelectValue placeholder="Select type" />
              </SelectTrigger>
              <SelectContent class="bg-zinc-900 border-white/10 text-white">
                <SelectItem value="openai">OpenAI</SelectItem>
                <SelectItem value="anthropic">Anthropic</SelectItem>
                <SelectItem value="gemini">Gemini</SelectItem>
                <SelectItem value="azure">Azure</SelectItem>
                <SelectItem value="custom">Custom</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="grid gap-2">
            <Label class="text-zinc-400 text-xs uppercase font-semibold">{{ t('channel_base_url') }}</Label>
            <div class="relative group/input">
              <Input v-model="form.base_url" placeholder="https://api.openai.com/v1" class="bg-white/5 border-white/10 h-10 pr-10" />
              <Button 
                variant="ghost" 
                size="icon" 
                @click="copyToClipboard(form.base_url, 'base_url')" 
                class="absolute right-1 top-1/2 -translate-y-1/2 h-8 w-8 text-zinc-500 hover:text-white"
              >
                <Check v-if="isCopied === 'base_url'" class="h-3.5 w-3.5 text-emerald-400" />
                <Copy v-else class="h-3.5 w-3.5" />
              </Button>
            </div>
          </div>
          
          <div class="grid gap-2">
            <Label class="text-zinc-400 text-xs uppercase font-semibold">{{ t('channel_api_key') }}</Label>
            <div class="relative group/input w-full">
              <Textarea 
                v-model="form.api_key_input" 
                placeholder="Enter one or more API keys (one per line or separated by spaces)" 
                class="bg-white/5 border-white/10 min-h-[120px] text-xs font-mono py-3 w-full resize-none focus:ring-cyan-500/30" 
              />
              <div class="absolute right-12 bottom-2 text-[10px] text-zinc-500 font-mono">
                {{ form.api_key_input ? form.api_key_input.split(/[\n\s]+/).filter(Boolean).length : 0 }} keys
              </div>
              <Button 
                variant="ghost" 
                size="icon" 
                @click="copyToClipboard(form.api_key_input, 'api_key')" 
                class="absolute right-1 top-1.5 h-8 w-8 text-zinc-500 hover:text-white"
              >
                <Check v-if="isCopied === 'api_key'" class="h-3.5 w-3.5 text-emerald-400" />
                <Copy v-else class="h-3.5 w-3.5" />
              </Button>
            </div>
          </div>
          
          <div class="space-y-4">
             <div class="flex items-center justify-between">
                <Label class="text-zinc-400 text-xs uppercase font-semibold">{{ t('channel_models') }}</Label>
                <Button variant="ghost" size="sm" @click="fetchModels" :disabled="isFetchingModels" class="h-7 text-[10px] text-cyan-400 hover:text-cyan-300 hover:bg-cyan-400/10 transition-all">
                  <RefreshCw :class="['w-3 h-3 mr-1', isFetchingModels ? 'animate-spin' : '']" />
                  获取可用模型
                </Button>
             </div>
             
             <!-- Tag Style Display -->
             <div class="space-y-3">
               <Input v-model="form.allowed_models" :placeholder="t('allowed_models_placeholder')" class="bg-white/5 border-white/10 text-xs font-mono focus:ring-cyan-500/30" />
               <div v-if="form.allowed_models" class="flex flex-wrap gap-1.5 p-3 rounded-xl border border-dashed border-white/10 bg-white/[0.02] max-h-60 overflow-y-auto custom-scrollbar">
                 <span v-for="m in form.allowed_models.split(',').filter(Boolean)" :key="m" class="group/tag px-2 py-0.5 rounded-md bg-cyan-500/10 border border-cyan-500/20 text-[10px] text-cyan-300 font-mono flex items-center gap-1 shadow-[0_0_10px_rgba(34,211,238,0.05)] transition-all hover:bg-cyan-500/20">
                   {{ m }} <button @click="removeModel(m)" class="ml-1 opacity-0 group-hover/tag:opacity-100 hover:text-white transition-opacity"><X class="w-2.5 h-2.5" /></button>
                 </span>
               </div>
             </div>
          </div>

          <div class="space-y-3">
            <Label class="text-zinc-400 text-xs uppercase font-semibold">{{ t('channel_denied_models') }}</Label>
            <Input v-model="form.denied_models" :placeholder="t('denied_models_placeholder')" class="bg-white/5 border-white/10 text-xs font-mono" />
            <div v-if="form.denied_models" class="flex flex-wrap gap-1.5 p-3 rounded-xl border border-dashed border-white/10 bg-white/[0.02] max-h-40 overflow-y-auto custom-scrollbar">
              <span v-for="m in form.denied_models.split(',').filter(Boolean)" :key="m" class="group/tag px-2 py-0.5 rounded-md bg-rose-500/10 border border-rose-500/20 text-[10px] text-rose-300 font-mono flex items-center gap-1 transition-all hover:bg-rose-500/20">
                {{ m }}
                <button @click="removeDeniedModel(m)" class="ml-1 opacity-0 group-hover/tag:opacity-100 hover:text-white transition-opacity">
                  <X class="w-2.5 h-2.5" />
                </button>
              </span>
            </div>
          </div>
        </div>

        <SheetFooter class="pt-6 border-t border-white/5">
          <Button variant="ghost" @click="isOpen = false" class="text-zinc-500 hover:text-white">{{ t('cancel') }}</Button>
          <Button @click="handleSave" :disabled="isSaving" class="px-8">
             <Loader2 v-if="isSaving" class="mr-2 h-4 w-4 animate-spin" />
             {{ t('save_channel') }}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  </div>
</template>

<style scoped>
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
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.2);
}

/* Pulse Highlight Animation */
.pulse-highlight {
  animation: pulse-bg 3s ease-in-out;
}

@keyframes pulse-bg {
  0% { background-color: transparent; }
  10% { background-color: rgba(34, 211, 238, 0.15); box-shadow: inset 0 0 20px rgba(34, 211, 238, 0.1); }
  50% { background-color: rgba(34, 211, 238, 0.05); }
  100% { background-color: transparent; }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
