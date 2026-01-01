import { getInitData } from './telegram'
import type { Room, Player, GameStats } from '../types'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const initData = getInitData()

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...options.headers as Record<string, string>,
  }

  // Send init data in both formats for compatibility
  if (initData) {
    headers['X-Telegram-Init-Data'] = initData
    headers['Authorization'] = `tma ${initData}`
  }

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Unknown error' }))
    throw new Error(error.error || 'Request failed')
  }

  return response.json()
}

export async function createRoom(category: string = 'general'): Promise<{ room: Room; player: Player }> {
  return request('/api/rooms', {
    method: 'POST',
    body: JSON.stringify({ category }),
  })
}

export async function getRoom(roomId: string): Promise<{ room: Room; players: Player[] }> {
  return request(`/api/rooms/${roomId}`)
}

export async function joinRoom(roomId: string): Promise<{ player: Player }> {
  return request(`/api/rooms/${roomId}/join`, { method: 'POST' })
}

export async function changeTeam(roomId: string, team: 'A' | 'B'): Promise<{ player: Player }> {
  return request(`/api/rooms/${roomId}/team`, {
    method: 'POST',
    body: JSON.stringify({ team }),
  })
}

export async function startGame(roomId: string): Promise<{ status: string }> {
  return request(`/api/rooms/${roomId}/start`, { method: 'POST' })
}

export async function getStats(roomId: string): Promise<GameStats> {
  return request(`/api/rooms/${roomId}/stats`)
}
