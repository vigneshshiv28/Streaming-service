"use client"

import type React from "react"

import { createContext, useContext, useEffect, useState } from "react"

interface User {
  id: string
  email: string
  name: string
  avatar?: string
}

interface AuthContextType {
  user: User | null
  isLoading: boolean
  signIn: (email: string, password: string) => Promise<void>
  signUp: (email: string, password: string, name: string) => Promise<void>
  signOut: () => Promise<void>
  signInWithGoogle: () => Promise<void>
  signInWithGitHub: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    // Simulate checking for existing session
    const checkAuth = async () => {
      try {
        const savedUser = localStorage.getItem("streamstudio_user")
        if (savedUser) {
          setUser(JSON.parse(savedUser))
        }
      } catch (error) {
        console.error("Auth check failed:", error)
      } finally {
        setIsLoading(false)
      }
    }

    checkAuth()
  }, [])

  const signIn = async (email: string, password: string) => {
    setIsLoading(true)
    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 1000))

      const mockUser: User = {
        id: "1",
        email,
        name: email.split("@")[0],
        avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=${email}`,
      }

      setUser(mockUser)
      localStorage.setItem("streamstudio_user", JSON.stringify(mockUser))
    } catch (error) {
      throw new Error("Invalid credentials")
    } finally {
      setIsLoading(false)
    }
  }

  const signUp = async (email: string, password: string, name: string) => {
    setIsLoading(true)
    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 1000))

      const mockUser: User = {
        id: "1",
        email,
        name,
        avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=${email}`,
      }

      setUser(mockUser)
      localStorage.setItem("streamstudio_user", JSON.stringify(mockUser))
    } catch (error) {
      throw new Error("Registration failed")
    } finally {
      setIsLoading(false)
    }
  }

  const signOut = async () => {
    setUser(null)
    localStorage.removeItem("streamstudio_user")
  }

  const signInWithGoogle = async () => {
    setIsLoading(true)
    try {
      // Simulate OAuth flow
      await new Promise((resolve) => setTimeout(resolve, 1500))

      const mockUser: User = {
        id: "1",
        email: "user@gmail.com",
        name: "Google User",
        avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=google",
      }

      setUser(mockUser)
      localStorage.setItem("streamstudio_user", JSON.stringify(mockUser))
    } catch (error) {
      throw new Error("Google sign-in failed")
    } finally {
      setIsLoading(false)
    }
  }

  const signInWithGitHub = async () => {
    setIsLoading(true)
    try {
      // Simulate OAuth flow
      await new Promise((resolve) => setTimeout(resolve, 1500))

      const mockUser: User = {
        id: "1",
        email: "user@github.com",
        name: "GitHub User",
        avatar: "https://api.dicebear.com/7.x/avataaars/svg?seed=github",
      }

      setUser(mockUser)
      localStorage.setItem("streamstudio_user", JSON.stringify(mockUser))
    } catch (error) {
      throw new Error("GitHub sign-in failed")
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        signIn,
        signUp,
        signOut,
        signInWithGoogle,
        signInWithGitHub,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider")
  }
  return context
}
