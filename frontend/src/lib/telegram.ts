declare global {
  interface Window {
    Telegram: {
      WebApp: TelegramWebApp
    }
  }
}

interface TelegramWebApp {
  initData: string
  initDataUnsafe: {
    user?: {
      id: number
      username?: string
      first_name?: string
      last_name?: string
      photo_url?: string
    }
    start_param?: string
  }
  version: string
  platform: string
  colorScheme: 'light' | 'dark'
  themeParams: {
    bg_color?: string
    text_color?: string
    hint_color?: string
    link_color?: string
    button_color?: string
    button_text_color?: string
    secondary_bg_color?: string
  }
  isExpanded: boolean
  viewportHeight: number
  viewportStableHeight: number
  headerColor: string
  backgroundColor: string
  ready: () => void
  expand: () => void
  close: () => void
  MainButton: {
    text: string
    color: string
    textColor: string
    isVisible: boolean
    isActive: boolean
    isProgressVisible: boolean
    setText: (text: string) => void
    onClick: (callback: () => void) => void
    offClick: (callback: () => void) => void
    show: () => void
    hide: () => void
    enable: () => void
    disable: () => void
    showProgress: (leaveActive?: boolean) => void
    hideProgress: () => void
  }
  BackButton: {
    isVisible: boolean
    onClick: (callback: () => void) => void
    offClick: (callback: () => void) => void
    show: () => void
    hide: () => void
  }
  HapticFeedback: {
    impactOccurred: (style: 'light' | 'medium' | 'heavy' | 'rigid' | 'soft') => void
    notificationOccurred: (type: 'error' | 'success' | 'warning') => void
    selectionChanged: () => void
  }
  showPopup: (params: {
    title?: string
    message: string
    buttons?: Array<{ id?: string; type?: string; text?: string }>
  }) => void
  showAlert: (message: string, callback?: () => void) => void
  showConfirm: (message: string, callback?: (ok: boolean) => void) => void
  openLink: (url: string) => void
  openTelegramLink: (url: string) => void
  switchInlineQuery: (query: string, choose_chat_types?: string[]) => void
}

export const tg = typeof window !== 'undefined' ? window.Telegram?.WebApp : null

export function initTelegram() {
  if (tg) {
    tg.ready()
    tg.expand()

    // Apply theme colors
    const root = document.documentElement
    if (tg.themeParams.bg_color) {
      root.style.setProperty('--tg-theme-bg-color', tg.themeParams.bg_color)
    }
    if (tg.themeParams.text_color) {
      root.style.setProperty('--tg-theme-text-color', tg.themeParams.text_color)
    }
    if (tg.themeParams.hint_color) {
      root.style.setProperty('--tg-theme-hint-color', tg.themeParams.hint_color)
    }
    if (tg.themeParams.link_color) {
      root.style.setProperty('--tg-theme-link-color', tg.themeParams.link_color)
    }
    if (tg.themeParams.button_color) {
      root.style.setProperty('--tg-theme-button-color', tg.themeParams.button_color)
    }
    if (tg.themeParams.button_text_color) {
      root.style.setProperty('--tg-theme-button-text-color', tg.themeParams.button_text_color)
    }
    if (tg.themeParams.secondary_bg_color) {
      root.style.setProperty('--tg-theme-secondary-bg-color', tg.themeParams.secondary_bg_color)
    }
  }
}

export function getInitData(): string {
  const initData = tg?.initData || ''

  // Debug logging
  console.log('[Telegram] Init Data:', initData ? 'present' : 'empty')
  console.log('[Telegram] Platform:', tg?.platform)
  console.log('[Telegram] Version:', tg?.version)

  return initData
}

export function getUser() {
  return tg?.initDataUnsafe?.user || null
}

export function getStartParam(): string | undefined {
  return tg?.initDataUnsafe?.start_param
}

export function shareRoom(roomId: string) {
  const botUsername = import.meta.env.VITE_BOT_USERNAME || 'zyaliasbot'

  // Create a direct link to the Mini App with startapp parameter
  const shareUrl = `https://t.me/${botUsername}?startapp=${roomId}`
  const shareText = 'Присоединяйся к игре в Alias!'

  if (tg) {
    // Use share URL to let user share the link
    tg.openTelegramLink(`https://t.me/share/url?url=${encodeURIComponent(shareUrl)}&text=${encodeURIComponent(shareText)}`)
  } else {
    navigator.clipboard.writeText(shareUrl)
    alert('Ссылка скопирована в буфер обмена!')
  }
}

export function hapticFeedback(type: 'success' | 'error' | 'warning' | 'light' | 'medium' | 'heavy') {
  if (!tg?.HapticFeedback) return

  switch (type) {
    case 'success':
    case 'error':
    case 'warning':
      tg.HapticFeedback.notificationOccurred(type)
      break
    case 'light':
    case 'medium':
    case 'heavy':
      tg.HapticFeedback.impactOccurred(type)
      break
  }
}

export function showMainButton(text: string, onClick: () => void) {
  if (!tg?.MainButton) return

  tg.MainButton.setText(text)
  tg.MainButton.onClick(onClick)
  tg.MainButton.show()
}

export function hideMainButton() {
  tg?.MainButton?.hide()
}
