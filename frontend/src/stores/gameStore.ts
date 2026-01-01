import { create } from 'zustand'
import type { Room, Player, Word, TelegramUser } from '../types'

interface GameStore {
  // User
  user: TelegramUser | null
  setUser: (user: TelegramUser | null) => void

  // Room state
  room: Room | null
  players: Player[]
  setRoom: (room: Room | null) => void
  setPlayers: (players: Player[]) => void
  addPlayer: (player: Player) => void
  removePlayer: (userId: number) => void
  updatePlayerTeam: (userId: number, team: 'A' | 'B') => void

  // Game state
  currentWord: Word | null
  secondsLeft: number
  teamScores: { A: number; B: number }
  setCurrentWord: (word: Word | null) => void
  setSecondsLeft: (seconds: number) => void
  setTeamScores: (scores: { A: number; B: number }) => void

  // UI state
  screen: 'loading' | 'home' | 'lobby' | 'game' | 'stats'
  setScreen: (screen: 'loading' | 'home' | 'lobby' | 'game' | 'stats') => void

  // WebSocket
  sendSwipe: ((action: 'up' | 'down' | 'left' | 'right') => void) | null
  setSendSwipe: (fn: ((action: 'up' | 'down' | 'left' | 'right') => void) | null) => void

  // Helpers
  isHost: () => boolean
  isExplainer: () => boolean
  getMyPlayer: () => Player | undefined
  getTeamPlayers: (team: 'A' | 'B') => Player[]

  // Reset
  reset: () => void
}

const initialState = {
  user: null,
  room: null,
  players: [],
  currentWord: null,
  secondsLeft: 60,
  teamScores: { A: 0, B: 0 },
  screen: 'loading' as const,
  sendSwipe: null,
}

export const useGameStore = create<GameStore>((set, get) => ({
  ...initialState,

  setUser: (user) => set({ user }),

  setRoom: (room) => set({ room }),

  setPlayers: (players) => set({ players }),

  addPlayer: (player) => set((state) => ({
    players: [...state.players.filter(p => p.user_id !== player.user_id), player],
  })),

  removePlayer: (userId) => set((state) => ({
    players: state.players.filter(p => p.user_id !== userId),
  })),

  updatePlayerTeam: (userId, team) => set((state) => ({
    players: state.players.map(p =>
      p.user_id === userId ? { ...p, team } : p
    ),
  })),

  setCurrentWord: (word) => set({ currentWord: word }),

  setSecondsLeft: (seconds) => set({ secondsLeft: seconds }),

  setTeamScores: (scores) => set({ teamScores: scores }),

  setScreen: (screen) => set({ screen }),

  setSendSwipe: (fn) => set({ sendSwipe: fn }),

  isHost: () => {
    const { user, players } = get()
    if (!user) return false
    const myPlayer = players.find(p => p.user_id === user.id)
    return myPlayer?.is_host ?? false
  },

  isExplainer: () => {
    const { user, room } = get()
    if (!user || !room) return false
    return room.current_explainer_id === user.id
  },

  getMyPlayer: () => {
    const { user, players } = get()
    if (!user) return undefined
    return players.find(p => p.user_id === user.id)
  },

  getTeamPlayers: (team) => {
    const { players } = get()
    return players.filter(p => p.team === team)
  },

  reset: () => set(initialState),
}))
