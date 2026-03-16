import React, { createContext, useCallback, useContext, useEffect, useState } from 'react'
import * as api from './api'

type User = { userId: string; email: string }

type AuthContextValue = {
  user: User | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, name?: string, phone?: string) => Promise<void>
  logout: () => Promise<void>
  error: string | null
  clearError: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

const ACCESS_KEY = 'dealer_access_token'
const REFRESH_KEY = 'dealer_refresh_token'

function loadTokens(): { access: string | null; refresh: string | null } {
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
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const validateOrRefresh = useCallback(async () => {
    const { access, refresh } = loadTokens()
    if (!access && !refresh) return
    if (access) {
      try {
        const data = await api.me(access)
        if (data.valid) {
          setUser({ userId: data.user_id, email: data.email })
          return
        }
      } catch {
        // access invalid, try refresh
      }
    }
    if (refresh) {
      try {
        const data = await api.refresh(refresh)
        saveTokens(data.access_token, data.refresh_token)
        const meData = await api.me(data.access_token)
        setUser({ userId: meData.user_id, email: meData.email })
        return
      } catch {
        clearTokens()
      }
    }
    setUser(null)
  }, [])

  useEffect(() => {
    validateOrRefresh().catch(() => setUser(null)).finally(() => setLoading(false))
  }, [validateOrRefresh])

  const login = useCallback(async (email: string, password: string) => {
    setError(null)
    const data = await api.login({ email, password })
    saveTokens(data.access_token, data.refresh_token)
    setUser({ userId: data.user_id, email: data.email })
  }, [])

  const register = useCallback(
    async (email: string, password: string, name?: string, phone?: string) => {
      setError(null)
      const data = await api.register({ email, password, name, phone })
      saveTokens(data.access_token, data.refresh_token)
      setUser({ userId: data.user_id, email: data.email })
    },
    [],
  )

  const logout = useCallback(async () => {
    const refresh = sessionStorage.getItem(REFRESH_KEY)
    if (refresh) {
      try {
        await api.logout(refresh)
      } catch {
        /* ignore */
      }
    }
    clearTokens()
    setUser(null)
  }, [])

  const value: AuthContextValue = {
    user,
    loading,
    login,
    register,
    logout,
    error,
    clearError: () => setError(null),
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
