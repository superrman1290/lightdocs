import { reactive } from 'vue'

export interface SecuritySettings {
  maxLoginFailures: number
  lockMinutes: number
  sessionDays: number
  rememberLogin: boolean
  logoutOtherDevicesOnPasswordChange: boolean
  requirePasswordReauthOnSave: boolean
}

export interface AdminSettings {
  username: string
  password: string
}

export interface SiteSettings {
  siteName: string
  siteTitle: string
  logoUrl: string
  security: SecuritySettings
  admin: AdminSettings
}

const storageKey = 'lightdocs-site-settings'
let storageListenerRegistered = false

export const defaultSecuritySettings: SecuritySettings = {
  maxLoginFailures: 5,
  lockMinutes: 15,
  sessionDays: 7,
  rememberLogin: false,
  logoutOtherDevicesOnPasswordChange: true,
  requirePasswordReauthOnSave: false,
}

export const defaultAdminSettings: AdminSettings = {
  username: 'admin',
  password: '1234',
}

export const siteSettings = reactive<SiteSettings>({
  siteName: '轻文档',
  siteTitle: '轻文档 - 专注技术教程的个人文档网站',
  logoUrl: '/favicon.svg',
  security: { ...defaultSecuritySettings },
  admin: { ...defaultAdminSettings },
})

export const applySiteSettings = (settings: Partial<SiteSettings>) => {
  if (settings.siteName !== undefined) {
    siteSettings.siteName = settings.siteName
  }

  if (settings.siteTitle !== undefined) {
    siteSettings.siteTitle = settings.siteTitle
  }

  if (settings.logoUrl !== undefined) {
    siteSettings.logoUrl = settings.logoUrl
  }

  Object.assign(
    siteSettings.security,
    defaultSecuritySettings,
    settings.security ?? {},
  )

  Object.assign(
    siteSettings.admin,
    defaultAdminSettings,
    settings.admin ?? {},
  )

  if (typeof document !== 'undefined') {
    document.title = siteSettings.siteTitle

    const faviconLinks = document.querySelectorAll<HTMLLinkElement>(
      'link[rel="icon"], link[rel="shortcut icon"]',
    )

    if (faviconLinks.length === 0) {
      const favicon = document.createElement('link')
      favicon.rel = 'icon'
      favicon.href = siteSettings.logoUrl
      document.head.appendChild(favicon)
    } else {
      faviconLinks.forEach(favicon => {
        favicon.href = siteSettings.logoUrl
      })
    }
  }
}

export const loadSiteSettings = () => {
  if (typeof window === 'undefined') {
    return
  }

  try {
    if (!storageListenerRegistered) {
      window.addEventListener('storage', event => {
        if (event.key !== storageKey || !event.newValue) {
          return
        }

        try {
          applySiteSettings(
            JSON.parse(event.newValue) as Partial<SiteSettings>,
          )
        } catch (error) {
          console.error('同步网站设置失败：', error)
        }
      })

      storageListenerRegistered = true
    }

    const saved = window.localStorage.getItem(storageKey)

    if (saved) {
      const parsed = JSON.parse(saved) as Partial<SiteSettings>

      // 迁移早期演示默认值，保留用户后来主动修改的密码。
      if (
        parsed.admin?.username === 'admin' &&
        parsed.admin.password === 'admin1234'
      ) {
        parsed.admin.password = defaultAdminSettings.password
        window.localStorage.setItem(storageKey, JSON.stringify(parsed))
      }

      applySiteSettings(parsed)
    } else {
      applySiteSettings(siteSettings)
    }
  } catch (error) {
    console.error('读取网站设置失败：', error)
    applySiteSettings(siteSettings)
  }
}

export const saveSiteSettings = (settings: SiteSettings) => {
  applySiteSettings(settings)

  if (typeof window !== 'undefined') {
    window.localStorage.setItem(
      storageKey,
      JSON.stringify(settings),
    )
  }
}
