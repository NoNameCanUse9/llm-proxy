<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import axios from 'axios'
import { 
  Settings, Save, ShieldCheck, Globe, Zap, Loader2, 
  User, Lock, Copy, Check 
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAppStore } from '@/stores/app'
import { useAppClipboard } from '@/composables/useAppClipboard'

const { t } = useI18n()
const store = useAppStore() 
const apiUrl = window.location.origin + '/v1'
const isLoading = ref(true)
const isSavingSettings = ref(false)
const isUpdatingUsername = ref(false)
const isUpdatingPassword = ref(false)

// 1. API Endpoint Settings
const settings = ref({
  enable_openai: true,
  enable_anthropic: true,
  enable_gemini: true
})

// 2. Account Settings
const username = ref(localStorage.getItem('username') || 'Admin')
const passwordForm = ref({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const fetchSettings = async () => {
  try {
    const res = await axios.get(`/admin/settings?_t=${Date.now()}`)
    console.log('[DEBUG] Fetched settings:', res.data)
    settings.value.enable_openai = !!res.data.enable_openai
    settings.value.enable_anthropic = !!res.data.enable_anthropic
    settings.value.enable_gemini = !!res.data.enable_gemini
  } catch (error) {
    console.error('Failed to fetch settings', error)
  } finally {
    isLoading.value = false
  }
}


const { copy: copyToClipboard } = useAppClipboard()

const handleCopyUrl = (path: string) => {
  const fullUrl = window.location.origin + path
  copyToClipboard(fullUrl)
}

const handleSaveSettings = async () => {
  isSavingSettings.value = true
  try {
    await axios.put('/admin/settings', settings.value)
    await fetchSettings()
    toast.success(t('settings_saved_success') || 'Settings saved successfully')
  } catch (error) {
    toast.error('Failed to save settings')
  } finally {
    isSavingSettings.value = false
  }
}

const handleUpdateUsername = async () => {
  if (!username.value) return
  isUpdatingUsername.value = true
  try {
    await axios.post('/admin/user/username', { username: username.value })
    localStorage.setItem('username', username.value)
    toast.success(t('username_update_success') || 'Username updated successfully')
  } catch (error) {
    toast.error('Failed to update username')
  } finally {
    isUpdatingUsername.value = false
  }
}

const handleChangePassword = async () => {
  if (!passwordForm.value.old_password || !passwordForm.value.new_password) {
    toast.error('Please fill in all password fields')
    return
  }
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    toast.error('Passwords do not match')
    return
  }
  isUpdatingPassword.value = true
  try {
    await axios.post('/admin/user/password', {
      old_password: passwordForm.value.old_password,
      new_password: passwordForm.value.new_password
    })
    toast.success(t('password_success') || 'Password updated successfully')
    passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
  } catch (error) {
    toast.error(t('password_error') || 'Failed to update password')
  } finally {
    isUpdatingPassword.value = false
  }
}

onMounted(fetchSettings)
</script>

<template>
  <div class="space-y-8 pb-10">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold tracking-tight text-white flex items-center gap-3">
          <Settings class="w-8 h-8 text-cyan-400" />
          {{ t('system_settings') }}
        </h1>
        <p class="text-zinc-400 mt-2">{{ t('manage_system_configurations') }}</p>
      </div>
    </div>

    <div v-if="isLoading" class="flex flex-col items-center justify-center py-20 gap-4">
      <Loader2 class="w-10 h-10 text-cyan-500 animate-spin" />
      <p class="text-zinc-500 animate-pulse">{{ t('loading_settings') }}</p>
    </div>

    <div v-else class="grid gap-8 lg:grid-cols-12">
      <!-- Left Column: Account Settings -->
      <div class="lg:col-span-7 space-y-8">
        <!-- Profile Card -->
        <div class="bg-zinc-900/50 border border-white/5 rounded-2xl p-6 backdrop-blur-xl">
          <div class="flex items-center gap-3 mb-6">
            <div class="p-2 bg-blue-500/10 rounded-lg">
              <User class="w-5 h-5 text-blue-400" />
            </div>
            <h2 class="text-lg font-semibold text-white">{{ t('account_mgmt') }}</h2>
          </div>

          <div class="space-y-4">
            <div class="grid gap-2">
              <Label for="username" class="text-zinc-400">{{ t('username') }}</Label>
              <div class="flex gap-2">
                <Input 
                  id="username" 
                  v-model="username" 
                  class="bg-white/5 border-white/10 text-white focus-visible:ring-cyan-500/50"
                  :placeholder="t('new_username_placeholder')"
                />
                <Button 
                  @click="handleUpdateUsername" 
                  variant="secondary"
                  class="shrink-0 bg-zinc-900 border border-white/10 text-white hover:bg-white hover:text-black transition-all duration-300"
                  :disabled="isUpdatingUsername"
                >
                  <Loader2 v-if="isUpdatingUsername" class="w-4 h-4 animate-spin" />
                  <span v-else>{{ t('update_username') }}</span>
                </Button>
              </div>
            </div>
          </div>
        </div>

        <!-- Password Card -->
        <div class="bg-zinc-900/50 border border-white/5 rounded-2xl p-6 backdrop-blur-xl">
          <div class="flex items-center gap-3 mb-6">
            <div class="p-2 bg-purple-500/10 rounded-lg">
              <Lock class="w-5 h-5 text-purple-400" />
            </div>
            <h2 class="text-lg font-semibold text-white">{{ t('change_password') }}</h2>
          </div>

          <div class="space-y-4">
            <div class="grid gap-2">
              <Label class="text-zinc-400">{{ t('original_password') }}</Label>
              <Input 
                type="password" 
                v-model="passwordForm.old_password"
                class="bg-white/5 border-white/10 text-white focus-visible:ring-cyan-500/50"
                :placeholder="t('old_password_placeholder')"
              />
            </div>
            <div class="grid gap-2">
              <Label class="text-zinc-400">{{ t('new_password') }}</Label>
              <Input 
                type="password" 
                v-model="passwordForm.new_password"
                class="bg-white/5 border-white/10 text-white focus-visible:ring-cyan-500/50"
                :placeholder="t('new_password_placeholder')"
              />
            </div>
            <div class="grid gap-2">
              <Label class="text-zinc-400">{{ t('confirm_password') }}</Label>
              <Input 
                type="password" 
                v-model="passwordForm.confirm_password"
                class="bg-white/5 border-white/10 text-white focus-visible:ring-cyan-500/50"
                :placeholder="t('confirm_password_placeholder')"
              />
            </div>
            <Button 
              @click="handleChangePassword" 
              class="w-full mt-2 bg-zinc-950 border border-emerald-500/40 text-emerald-400 hover:bg-emerald-500 hover:text-black shadow-[0_0_15px_rgba(16,185,129,0.2)] transition-all duration-300 font-mono"
              :disabled="isUpdatingPassword"
            >
              <Loader2 v-if="isUpdatingPassword" class="w-4 h-4 mr-2 animate-spin" />
              {{ t('save_changes') }}
            </Button>
          </div>
        </div>
      </div>

      <!-- Right Column: API Controls & Info -->
      <div class="lg:col-span-5 space-y-8">
        <!-- API Toggles -->
        <div class="bg-zinc-900/50 border border-white/5 rounded-2xl p-6 backdrop-blur-xl shadow-xl">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-cyan-500/10 rounded-lg">
                <Globe class="w-5 h-5 text-cyan-400" />
              </div>
              <h2 class="text-lg font-semibold text-white">{{ t('endpoint_visibility') }}</h2>
            </div>
            <Button 
              size="sm"
              @click="handleSaveSettings" 
              class="bg-zinc-950 border border-cyan-500/40 text-cyan-400 hover:bg-cyan-500 hover:text-black shadow-[0_0_15px_rgba(6,182,212,0.2)] transition-all duration-300 font-mono tracking-wider"
              :disabled="isSavingSettings"
            >
              <Save v-if="!isSavingSettings" class="w-4 h-4 mr-2" />
              <Loader2 v-else class="w-4 h-4 mr-2 animate-spin" />
              {{ t('save') }}
            </Button>
          </div>
          
          <div class="space-y-3">
            <!-- OpenAI -->
            <div class="flex items-center justify-between p-3 rounded-xl hover:bg-white/5 transition-colors group border border-transparent hover:border-white/5">
              <div class="space-y-1">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium text-zinc-300 group-hover:text-white transition-colors">OpenAI Endpoint</span>
                  <Button size="icon" variant="ghost" class="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity" @click="handleCopyUrl('/v1/chat/completions')">
                    <Copy class="w-3 h-3 text-cyan-400" />
                  </Button>
                </div>
                <code class="text-[10px] text-zinc-500 font-mono">/v1/chat/completions</code>
              </div>
              <Switch v-model="settings.enable_openai" />
            </div>

            <!-- Anthropic -->
            <div class="flex items-center justify-between p-3 rounded-xl hover:bg-white/5 transition-colors group border border-transparent hover:border-white/5">
              <div class="space-y-1">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium text-zinc-300 group-hover:text-white transition-colors">Anthropic Endpoint</span>
                  <Button size="icon" variant="ghost" class="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity" @click="handleCopyUrl('/v1/messages')">
                    <Copy class="w-3 h-3 text-cyan-400" />
                  </Button>
                </div>
                <code class="text-[10px] text-zinc-500 font-mono">/v1/messages</code>
              </div>
              <Switch v-model="settings.enable_anthropic" />
            </div>

            <!-- Gemini -->
            <div class="flex items-center justify-between p-3 rounded-xl hover:bg-white/5 transition-colors group border border-transparent hover:border-white/5">
              <div class="space-y-1">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium text-zinc-300 group-hover:text-white transition-colors">Gemini Endpoint</span>
                  <Button size="icon" variant="ghost" class="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity" @click="handleCopyUrl('/v1/models')">
                    <Copy class="w-3 h-3 text-cyan-400" />
                  </Button>
                </div>
                <code class="text-[10px] text-zinc-500 font-mono">/v1/models</code>
              </div>
              <Switch v-model="settings.enable_gemini" />
            </div>
          </div>
        </div>

        <!-- Security Info -->
        <div class="bg-zinc-900/50 border border-white/5 rounded-2xl p-6 backdrop-blur-xl border-l-cyan-500/50 border-l-2">
          <div class="flex items-center gap-3 mb-4">
            <ShieldCheck class="w-5 h-5 text-cyan-400" />
            <h2 class="text-md font-semibold text-white">{{ t('security_policy') }}</h2>
          </div>
          <p class="text-xs text-zinc-400 leading-relaxed mb-4">
            {{ t('security_policy_description') }}
          </p>
          <div class="p-3 rounded-lg bg-cyan-500/5 border border-cyan-500/10 flex gap-2">
            <Zap class="w-4 h-4 text-cyan-400 shrink-0" />
            <p class="text-[10px] text-cyan-400/80 leading-tight">
              {{ t('security_notice_rebuild') }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
