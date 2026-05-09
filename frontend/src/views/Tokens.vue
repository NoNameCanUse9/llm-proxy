<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import axios from 'axios'
import { Plus, Pencil, Trash2, Key, Shield, Globe, Clock, Copy, Check, Loader2, ListChecks } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { useAppClipboard } from '@/composables/useAppClipboard'
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { Switch } from '@/components/ui/switch'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const store = useAppStore()

const isOpen = ref(false)
const isSaving = ref(false)
const editingId = ref<number | null>(null)
const revealModal = ref(false)
const newlyCreatedToken = ref('')

const form = ref({
  name: '',
  expires_at: 'permanent',
  allowed_channels: '', // Comma separated IDs
  allowed_models: '',
  denied_models: '',
  allowed_ips: '',
  denied_ips: '',
  rpm: 0,
  is_active: true
})

onMounted(() => {
  store.fetchTokens()
  store.fetchChannels()
})

const openAdd = () => {
  editingId.value = null
  form.value = {
    name: '',
    expires_at: 'permanent',
    allowed_channels: '',
    allowed_models: '',
    denied_models: '',
    allowed_ips: '',
    denied_ips: '',
    rpm: 0,
    is_active: true
  }
  isOpen.value = true
}

const openEdit = (token: any) => {
  editingId.value = token.id
  form.value = { ...token, expires_at: 'permanent' }
  isOpen.value = true
}

const handleSave = async () => {
  isSaving.value = true
  try {
    let res;
    if (editingId.value) {
      res = await axios.put(`/admin/tokens/${editingId.value}`, form.value)
    } else {
      res = await axios.post('/admin/tokens', form.value)
      if (res.data.token) {
        newlyCreatedToken.value = res.data.token
        revealModal.value = true
      }
    }
    await store.fetchTokens()
    isOpen.value = false
  } catch (err) {
    console.error('Save failed', err)
  } finally {
    isSaving.value = false
  }
}

const toggleChannel = (id: number) => {
  const channels = form.value.allowed_channels ? form.value.allowed_channels.split(',') : []
  const index = channels.indexOf(id.toString())
  if (index > -1) {
    channels.splice(index, 1)
  } else {
    channels.push(id.toString())
  }
  form.value.allowed_channels = channels.join(',')
}

const { copy: copyToClipboard, copied: isCopied } = useAppClipboard()

const handleDelete = async (id: number) => {
  if (confirm(t('confirm_delete_token'))) {
    await store.deleteToken(id)
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-3xl font-bold tracking-tight text-white">{{ t('tokens_title') }}</h2>
        <p class="text-zinc-400 mt-2">{{ t('tokens_desc') }}</p>
      </div>
      <Button @click="openAdd">
        <Plus class="w-4 h-4 mr-2" />
        {{ t('add_token') }}
      </Button>
    </div>

    <div class="rounded-2xl border border-white/5 bg-zinc-900/40 backdrop-blur-xl overflow-hidden">
      <Table>
        <TableHeader class="bg-white/5">
          <TableRow class="hover:bg-transparent border-white/5">
            <TableHead class="text-zinc-400 font-medium">{{ t('name') }}</TableHead>
            <TableHead class="text-zinc-400 font-medium">{{ t('status') }}</TableHead>
            <TableHead class="text-zinc-400 font-medium">{{ t('policy') }}</TableHead>
            <TableHead class="text-zinc-400 font-medium">{{ t('created_at') }}</TableHead>
            <TableHead class="text-right text-zinc-400 font-medium">{{ t('actions') }}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-if="store.tokens.length === 0">
            <TableCell colspan="5" class="h-24 text-center text-zinc-500 border-white/5 bg-white/[0.01]">
              <div class="flex flex-col items-center gap-2">
                <div class="w-8 h-8 rounded-full bg-zinc-500/10 flex items-center justify-center">
                  <Key class="w-4 h-4 text-zinc-600" />
                </div>
                <span>{{ t('no_tokens_found') }}</span>
              </div>
            </TableCell>
          </TableRow>
          <TableRow v-for="token in store.tokens" :key="token.id" class="border-white/5 hover:bg-white/[0.02]" v-else>
            <TableCell class="font-medium text-white">
              <div class="flex flex-col group/key relative">
                <span class="flex items-center gap-2">
                  <Key class="w-3.5 h-3.5 text-cyan-400" />
                  {{ token.name }}
                </span>
                <div class="flex items-center gap-2 mt-1">
                  <span class="text-[10px] text-zinc-500 font-mono">sk-{{ token.token ? token.token.slice(3, 8) + '••••' + token.token.slice(-4) : '••••••••••' }}</span>
                  <Button 
                    v-if="token.token"
                    variant="ghost" 
                    size="icon" 
                    @click="copyToClipboard(token.token)" 
                    class="h-6 w-6 text-zinc-500 hover:text-cyan-400 hover:bg-cyan-400/10 transition-all rounded-md"
                  >
                    <Copy class="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            </TableCell>
            <TableCell>
              <span :class="['px-2 py-0.5 rounded-full text-[10px] border', token.is_active ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-400' : 'border-zinc-500/20 bg-zinc-500/10 text-zinc-400']">
                {{ token.is_active ? t('active') : t('inactive') }}
              </span>
            </TableCell>
            <TableCell>
               <div class="flex gap-3 text-zinc-500">
                 <Shield class="w-4 h-4" v-if="token.allowed_models" />
                 <Globe class="w-4 h-4" v-if="token.allowed_ips" />
                 <ListChecks class="w-4 h-4" v-if="token.allowed_channels" />
                 <Clock class="w-4 h-4" v-if="token.expires_at !== 'permanent'" />
               </div>
            </TableCell>
            <TableCell class="text-zinc-500 text-xs font-mono">
              {{ new Date(token.created_at).toLocaleDateString() }}
            </TableCell>
            <TableCell class="text-right">
              <div class="flex justify-end gap-1">
                <Button variant="ghost" size="icon" @click="openEdit(token)" class="h-8 w-8 text-zinc-500 hover:text-white hover:bg-white/5">
                  <Pencil class="w-3.5 h-3.5" />
                </Button>
                <Button variant="ghost" size="icon" @click="handleDelete(token.id)" class="h-8 w-8 text-zinc-500 hover:text-rose-500 hover:bg-rose-500/10">
                  <Trash2 class="w-3.5 h-3.5" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <!-- Token Sheet -->
    <Sheet v-model:open="isOpen">
      <SheetContent side="right" class="w-full sm:max-w-2xl bg-[#0a0c10] border-l border-white/10 text-white shadow-2xl overflow-y-auto px-10 z-[100]">
        <SheetHeader class="pb-6 border-b border-white/5">
          <SheetTitle class="text-xl font-bold text-white">{{ editingId ? t('edit_token') : t('add_token') }}</SheetTitle>
          <SheetDescription class="text-zinc-500">
            {{ t('tokens_desc') }}
          </SheetDescription>
        </SheetHeader>

        <div class="grid gap-8 py-10">
          <div class="flex items-center justify-between">
            <Label class="text-zinc-400 text-xs uppercase font-semibold">{{ t('status') }}</Label>
            <Switch v-model="form.is_active" />
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="grid gap-2">
              <Label class="text-zinc-400 text-xs uppercase font-semibold">{{ t('token_name') }}</Label>
              <Input v-model="form.name" class="bg-white/5 border-white/10 h-10" />
            </div>
            <div class="grid gap-2">
              <Label class="text-zinc-400 text-xs uppercase font-semibold">{{ t('expiry') }}</Label>
              <select v-model="form.expires_at" class="bg-white/5 border border-white/10 rounded-md h-10 px-3 text-sm focus:outline-none focus:ring-1 focus:ring-cyan-500/30">
                <option value="permanent" class="bg-zinc-900">{{ t('permanent') }}</option>
                <option value="7d" class="bg-zinc-900">{{ t('7_days') }}</option>
                <option value="30d" class="bg-zinc-900">{{ t('30_days') }}</option>
                <option value="180d" class="bg-zinc-900">{{ t('180_days') }}</option>
                <option value="365d" class="bg-zinc-900">{{ t('365_days') }}</option>
              </select>
            </div>
          </div>

          <div class="grid gap-2">
            <Label class="text-zinc-400 text-xs uppercase font-semibold">{{ t('rpm_limit') }}</Label>
            <Input v-model.number="form.rpm" type="number" class="bg-white/5 border-white/10 h-10" placeholder="0 = unlimited" />
          </div>

          <div class="space-y-4">
            <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-600 border-b border-white/5 pb-2">{{ t('access_control') }}</h3>
            
            <Accordion type="single" collapsible class="w-full">
              <!-- Channels Section -->
              <AccordionItem value="channels" class="border-white/5">
                <AccordionTrigger class="text-sm hover:no-underline py-4 text-zinc-300">
                  <div class="flex items-center gap-2">
                    <ListChecks class="w-4 h-4 text-emerald-400" />
                    {{ t('allowed_channels') }}
                  </div>
                </AccordionTrigger>
                <AccordionContent class="pb-4">
                  <div class="grid grid-cols-2 gap-3 max-h-48 overflow-y-auto pr-2 custom-scrollbar">
                    <div v-for="ch in store.channels" :key="ch.id" class="flex items-center space-x-2 bg-white/5 p-2 rounded border border-white/5">
                      <Checkbox :id="'ch-'+ch.id" :checked="form.allowed_channels.split(',').includes(ch.id.toString())" @update:checked="toggleChannel(ch.id)" />
                      <label :for="'ch-'+ch.id" class="text-xs text-zinc-300 truncate cursor-pointer">{{ ch.name }}</label>
                    </div>
                  </div>
                </AccordionContent>
              </AccordionItem>

              <!-- Models Section -->
              <AccordionItem value="models" class="border-white/5">
                <AccordionTrigger class="text-sm hover:no-underline py-4 text-zinc-300">
                  <div class="flex items-center gap-2">
                    <Shield class="w-4 h-4 text-purple-400" />
                    {{ t('models') }}
                  </div>
                </AccordionTrigger>
                <AccordionContent class="pb-4 space-y-4">
                  <div class="grid gap-2">
                    <Label class="text-[10px] text-zinc-500 uppercase">{{ t('allowed_models') }}</Label>
                    <Input v-model="form.allowed_models" placeholder="gpt-4,*" class="bg-white/5 border-white/10 font-mono text-xs" />
                  </div>
                  <div class="grid gap-2">
                    <Label class="text-[10px] text-zinc-500 uppercase">{{ t('denied_models') }}</Label>
                    <Input v-model="form.denied_models" placeholder="dall-e-3" class="bg-white/5 border-white/10 font-mono text-xs" />
                  </div>
                </AccordionContent>
              </AccordionItem>

              <!-- IPs Section -->
              <AccordionItem value="ips" class="border-white/5">
                <AccordionTrigger class="text-sm hover:no-underline py-4 text-zinc-300">
                  <div class="flex items-center gap-2">
                    <Globe class="w-4 h-4 text-cyan-400" />
                    {{ t('ip_address') }}
                  </div>
                </AccordionTrigger>
                <AccordionContent class="pb-4 space-y-4">
                  <div class="grid gap-2">
                    <Label class="text-[10px] text-zinc-500 uppercase">{{ t('allowed_ips') }}</Label>
                    <Input v-model="form.allowed_ips" placeholder="192.168.1.100" class="bg-white/5 border-white/10 font-mono text-xs" />
                  </div>
                  <div class="grid gap-2">
                    <Label class="text-[10px] text-zinc-500 uppercase">{{ t('denied_ips') }}</Label>
                    <Input v-model="form.denied_ips" placeholder="220.181.38.*" class="bg-white/5 border-white/10 font-mono text-xs" />
                  </div>
                </AccordionContent>
              </AccordionItem>
            </Accordion>
          </div>
        </div>

        <SheetFooter class="pt-6 border-t border-white/5">
          <Button variant="ghost" @click="isOpen = false" class="text-zinc-500 hover:text-white">{{ t('cancel') }}</Button>
          <Button @click="handleSave" :disabled="isSaving" class="px-8">
             <Loader2 v-if="isSaving" class="mr-2 h-4 w-4 animate-spin" />
             {{ t('save_token') }}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>

    <!-- Reveal Token Modal -->
    <Dialog v-model:open="revealModal">
      <DialogContent class="bg-[#0a0c10] border-white/10 text-white max-w-md">
        <DialogHeader>
          <DialogTitle class="text-xl font-bold flex items-center gap-2">
            <div class="h-8 w-8 rounded bg-emerald-500/20 flex items-center justify-center">
              <Check class="h-4 w-4 text-emerald-400" />
            </div>
            {{ t('token_created') }}
          </DialogTitle>
          <DialogDescription class="text-zinc-400 pt-2">
            {{ t('token_reveal_desc') }}
          </DialogDescription>
        </DialogHeader>

        <div class="mt-6 p-4 rounded-xl bg-emerald-500/5 border border-emerald-500/20 flex items-center gap-3">
          <code class="flex-1 text-sm font-mono text-emerald-400 break-all select-all">{{ newlyCreatedToken }}</code>
          <Button variant="ghost" size="icon" @click="copyToClipboard" class="text-zinc-400 hover:text-white hover:bg-white/10">
            <Check v-if="isCopied" class="h-4 w-4 text-emerald-400" />
            <Copy v-else class="h-4 w-4" />
          </Button>
        </div>

        <div class="mt-6 flex justify-end">
          <Button @click="revealModal = false" class="bg-white text-black hover:bg-white/90">{{ t('done') }}</Button>
        </div>
      </DialogContent>
    </Dialog>
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
</style>
