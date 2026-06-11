import { createContext, useCallback, useContext, useEffect, useState } from 'react'
import * as api from './api'

type User = { userId: string; email: string }

type AuthContextValue = {
  user: User | null
  accessToken: string | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (payload: {
    email: string
    password: string
    full_name: string
    phone?: string
    vin?: string
  }) => Promise<void>
  logout: () => Promise<void>
  getAccessToken: () => Promise<string | null>
}

const AuthContext = createContext<AuthContextValue | null>(null)

const ACCESS_KEY = 'dealer_client_access_token'
const REFRESH_KEY = 'dealer_client_refresh_token'

function loadTokens() {
  try {
    return {
      access: sessionStorage.getItem(ACCESS_KEY),
      refresh: sessionStorage.getItem(REFRESH_KEY),
    }
  } catch {
    return { access: null, refresh: null }
  }
}

function saveTokens(access: string, refresh: string) {
  sessionStorage.setItem(ACCESS_KEY, access)
  sessionStorage.setItem(REFRESH_KEY, refresh)
}

function clearTokens() {
  sessionStorage.removeItem(ACCESS_KEY)
  sessionStorage.removeItem(REFRESH_KEY)
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [accessToken, setAccessToken] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  const validateOrRefresh = useCallback(async () => {
    const { access, refresh } = loadTokens()
    if (!access && !refresh) return
    if (access) {
      try {
        const data = await api.me(access)
        if (data.valid) {
          setUser({ userId: data.user_id, email: data.email })
          setAccessToken(access)
          return
        }
      } catch {
        /* try refresh */
      }
    }
    if (refresh) {
      try {
        const data = await api.refresh(refresh)
        saveTokens(data.access_token, data.refresh_token)
        const meData = await api.me(data.access_token)
        if (meData.valid) {
          setUser({ userId: meData.user_id, email: meData.email })
          setAccessToken(data.access_token)
          return
        }
      } catch {
        clearTokens()
      }
    }
    setUser(null)
    setAccessToken(null)
  }, [])

  useEffect(() => {
    validateOrRefresh()
      .catch(() => {
        setUser(null)
        setAccessToken(null)
      })
      .finally(() => setLoading(false))
  }, [validateOrRefresh])

  const getAccessToken = useCallback(async () => {
    const { access, refresh } = loadTokens()
    if (access) {
      try {
        const data = await api.me(access)
        if (data.valid) return access
      } catch {
        /* refresh below */
      }
    }
    if (!refresh) return null
    try {
      const data = await api.refresh(refresh)
      saveTokens(data.access_token, data.refresh_token)
      setAccessToken(data.access_token)
      return data.access_token
    } catch {
      clearTokens()
      setUser(null)
      setAccessToken(null)
      return null
    }
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    const data = await api.login(email, password)
    saveTokens(data.access_token, data.refresh_token)
    setUser({ userId: data.user_id, email: data.email })
    setAccessToken(data.access_token)
  }, [])

  const register = useCallback(
    async (payload: { email: string; password: string; full_name: string; phone?: string; vin?: string }) => {
      const data = await api.registerClient(payload)
      saveTokens(data.access_token, data.refresh_token)
      setUser({ userId: data.user_id, email: data.email })
      setAccessToken(data.access_token)
    },
    [],
  )

  const logout = useCallback(async () => {
    const { refresh } = loadTokens()
    if (refresh) {
      try {
        await api.logout(refresh)
      } catch {
        /* ignore */
      }
    }
    clearTokens()
    setUser(null)
    setAccessToken(null)
  }, [])

  return (
    <AuthContext.Provider value={{ user, accessToken, loading, login, register, logout, getAccessToken }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
