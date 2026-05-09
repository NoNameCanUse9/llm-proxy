<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Activity, Key, LayoutDashboard, Search, Settings, Server, Languages, LogOut } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { 
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger 
} from '@/components/ui/dropdown-menu'

const router = useRouter()
const { t, locale } = useI18n()

const navItems = [
  { name: 'dashboard', icon: LayoutDashboard, path: '/' },
  { name: 'providers', icon: Server, path: '/channels' },
  { name: 'tokens', icon: Key, path: '/tokens' },
  { name: 'logs', icon: Activity, path: '/logs' },
  { name: 'settings', icon: Settings, path: '/settings' },
]

const toggleLocale = () => {
  locale.value = locale.value === 'zh' ? 'en' : 'zh'
  localStorage.setItem('locale', locale.value)
}

const handleLogout = () => {
  localStorage.removeItem('token')
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen bg-[#0a0c10] text-white flex flex-col font-sans selection:bg-cyan-500/30">
    <!-- Top Navbar -->
    <header class="sticky top-0 z-50 w-full border-b border-white/5 bg-[#0a0c10]/80 backdrop-blur-xl">
      <div class="container flex h-16 items-center justify-between px-4 sm:px-8">
        <!-- Logo -->
        <div class="flex items-center gap-3 cursor-pointer" @click="router.push('/')">
          <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-purple-500 to-cyan-500 shadow-[0_0_15px_rgba(34,211,238,0.3)]">
            <Server class="h-4 w-4 text-white" />
          </div>
          <span class="text-lg font-semibold tracking-tight">LLM Proxy</span>
        </div>

        <!-- Navigation -->
        <nav class="hidden md:flex items-center space-x-1">
          <Button 
            v-for="item in navItems" 
            :key="item.path"
            variant="ghost" 
            class="h-9 px-4 rounded-full text-sm font-medium text-zinc-400 hover:text-white hover:bg-white/5 data-[active=true]:bg-white/10 data-[active=true]:text-white transition-colors"
            :data-active="router.currentRoute.value.path === item.path"
            @click="router.push(item.path)"
          >
            <component :is="item.icon" class="w-4 h-4 mr-2" />
            {{ t(item.name) }}
          </Button>
        </nav>

        <!-- Right Side -->
        <div class="flex items-center gap-4">
          <!-- Locale Toggle -->
          <Button variant="ghost" size="icon" class="rounded-full text-zinc-400 hover:text-white" @click="toggleLocale">
            <Languages class="h-5 w-5" />
          </Button>

          <!-- User Menu -->
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" class="rounded-full text-zinc-400 hover:text-white">
                <Settings class="h-5 w-5" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" class="w-56 bg-[#14171f] border-white/10 text-white">
              <DropdownMenuLabel>{{ t('administrator') }}</DropdownMenuLabel>
              <DropdownMenuSeparator class="bg-white/5" />
              <DropdownMenuItem class="focus:bg-white/5 focus:text-white cursor-pointer" @click="router.push('/settings')">
                <Settings class="mr-2 h-4 w-4" />
                <span>{{ t('settings') }}</span>
              </DropdownMenuItem>
              <DropdownMenuItem class="focus:bg-rose-500/10 focus:text-rose-500 cursor-pointer text-rose-400" @click="handleLogout">
                <LogOut class="mr-2 h-4 w-4" />
                <span>{{ t('logout') }}</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="flex-1 container mx-auto px-4 sm:px-8 py-8">
      <router-view v-slot="{ Component }">
        <transition name="page" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<style scoped>
.page-enter-active,
.page-leave-active {
  transition: opacity 0.15s ease-out;
}

.page-enter-from,
.page-leave-to {
  opacity: 0;
}
</style>
