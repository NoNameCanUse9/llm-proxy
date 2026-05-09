import { ref } from 'vue'
import { useClipboard as useVueUseClipboard } from '@vueuse/core'
import { toast } from 'vue-sonner'
import { useI18n } from 'vue-i18n'

export function useAppClipboard() {
  const { t } = useI18n()
  const { copy: vueUseCopy, isSupported } = useVueUseClipboard()
  const copied = ref<string | boolean>(false)

  const copyWithToast = (text: string, id: string | boolean = true) => {
    if (!text) return
    
    vueUseCopy(text)
    copied.value = id
    
    toast.success(t('copied_success'), {
      description: typeof id === 'string' ? id : (text.length > 30 ? text.substring(0, 30) + '...' : text),
      duration: 2000,
    })

    setTimeout(() => {
      copied.value = false
    }, 2000)
  }

  return {
    copy: copyWithToast,
    copied,
    isSupported
  }
}
