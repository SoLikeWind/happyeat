import { createContext, useCallback, useContext, useState } from 'react'
import { getToken, setToken as saveToken, clearToken } from '../api/client'

interface AuthContextValue {
  token: string | null
  isLoggedIn: boolean
  setToken: (token: string) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setTokenState] = useState<string | null>(getToken)

  const setToken = useCallback((t: string) => {
    saveToken(t)
    setTokenState(t)
  }, [])

  const logout = useCallback(() => {
    clearToken()
    setTokenState(null)
  }, [])

  const value: AuthContextValue = {
    token,
    isLoggedIn: !!token,
    setToken,
    logout,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
