"use client"

import type React from "react"
import type { User } from "@/types/models"
import { useUserStore } from "@/store/userStore"
import { createContext, useContext, useEffect, useState } from "react"



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
  const user = useUserStore((state) => state.user)
  const setUser = useUserStore((state) => state.setUser)
  const removeUser = useUserStore((state) => state.removeUser)
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
    removeUser()
    localStorage.removeItem("streamstudio_user")
  }

  const signInWithGoogle = async () => {
    setIsLoading(true)
    try {
      // Simulate OAuth flow
      await new Promise((resolve) => setTimeout(resolve, 1500))
      const randomSuffix = Math.floor(Math.random() * 10000);
      const mockUser: User = {
        id: `${randomSuffix}`,
        email: `user${randomSuffix}@gmail.com`,
        name: `Google User ${randomSuffix}` ,

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
