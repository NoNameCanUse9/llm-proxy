<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Server, Loader2 } from 'lucide-vue-next'

const router = useRouter()
const { t } = useI18n()

const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

const handleLogin = async () => {
  loading.value = true
  error.value = ''
  try {
    const res = await axios.post('/auth/login', {
      username: username.value,
      password: password.value
    })
    
    if (res.data.token) {
      localStorage.setItem('token', res.data.token)
      // Configure axios default header for future requests
      axios.defaults.headers.common['Authorization'] = `Bearer ${res.data.token}`
      router.push('/')
    }
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-[#0a0c10] flex items-center justify-center p-4 selection:bg-cyan-500/30">
    <div class="max-w-md w-full space-y-8 p-8 border border-white/10 rounded-2xl bg-zinc-900/50 backdrop-blur-xl shadow-2xl relative overflow-hidden">
      <!-- Glow effect -->
      <div class="absolute -top-24 -left-24 w-48 h-48 bg-purple-500/20 blur-[80px]"></div>
      <div class="absolute -bottom-24 -right-24 w-48 h-48 bg-cyan-500/20 blur-[80px]"></div>

      <div class="text-center relative">
        <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-purple-500 to-cyan-500 mb-4 shadow-lg shadow-cyan-500/20">
          <Server class="h-6 w-6 text-white" />
        </div>
        <h2 class="text-2xl font-bold tracking-tight text-white">{{ t('welcome_back') }}</h2>
        <p class="text-zinc-400 mt-2 text-sm">{{ t('sign_in_desc') }}</p>
      </div>

      <form class="mt-8 space-y-6 relative" @submit.prevent="handleLogin">
        <div class="space-y-4">
          <div class="space-y-2">
            <Label for="username" class="text-zinc-400 text-sm">{{ t('username') }}</Label>
            <Input 
              id="username" 
              v-model="username" 
              required 
              class="bg-white/5 border-white/10 text-white focus:ring-cyan-500/50" 
            />
          </div>
          <div class="space-y-2">
            <Label for="password" class="text-zinc-400 text-sm">{{ t('password') }}</Label>
            <Input 
              id="password" 
              v-model="password" 
              type="password" 
              required 
              class="bg-white/5 border-white/10 text-white focus:ring-cyan-500/50" 
            />
          </div>
        </div>

        <div v-if="error" class="text-rose-500 text-xs bg-rose-500/10 p-3 rounded-lg border border-rose-500/20">
          {{ error }}
        </div>

        <Button 
          type="submit" 
          :disabled="loading"
          class="w-full h-11"
        >
          <Loader2 v-if="loading" class="mr-2 h-4 w-4 animate-spin" />
          {{ t('login_btn') }}
        </Button>
      </form>
    </div>
  </div>
</template>
