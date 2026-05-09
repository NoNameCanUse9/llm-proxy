import { useClipboard as useVueUseClipboard } from '@vueuse/core'
import { toast } from 'vue-sonner'
import { useI18n } from 'vue-i18n'

export function useAppClipboard() {
  const { t } = useI18n()
  const { copy: vueUseCopy, copied, isSupported } = useVueUseClipboard()

  const copyWithToast = (text: string, description?: string) => {
    if (!text) return
    
    vueUseCopy(text)
    
    toast.success(t('copied_success'), {
      description: description || (text.length > 30 ? text.substring(0, 30) + '...' : text),
      duration: 2000,
    })
  }

  return {
    copy: copyWithToast,
    copied,
    isSupported
  }
}
