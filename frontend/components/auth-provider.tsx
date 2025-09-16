"use client"

import type React from "react"

import { AuthProvider } from "@/lib/auth"

export function AuthProviderWrapper({ children }: { children: React.ReactNode }) {
  return <AuthProvider>{children}</AuthProvider>
}

