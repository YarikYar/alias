import { useEffect, useRef, useCallback, useState } from 'react'
import { getInitData } from '../lib/telegram'
import { useGameStore } from '../stores/gameStore'
import type {
  WSMessage,
  PlayerJoinedPayload,
  PlayerLeftPayload,
  TeamChangedPayload,
  GameStartedPayload,
  NewWordPayload,
  TimerPayload,
  RoundEndPayload,
  GameEndPayload,
} from '../types'

const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080'

export function useWebSocket(roomId: string | null) {
  const wsRef = useRef<WebSocket | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const reconnectTimeoutRef = useRef<number | null>(null)

  const {
    addPlayer,
    removePlayer,
    updatePlayerTeam,
    setRoom,
    setCurrentWord,
    setSecondsLeft,
    setTeamScores,
    setScreen,
    room,
  } = useGameStore()

  const connect = useCallback(() => {
    if (!roomId) return

    const initData = encodeURIComponent(getInitData())
    const ws = new WebSocket(`${WS_URL}/ws/${roomId}?init_data=${initData}`)

    ws.onopen = () => {
      console.log('WebSocket connected')
      setIsConnected(true)
    }

    ws.onclose = () => {
      console.log('WebSocket disconnected')
      setIsConnected(false)

      // Reconnect after 3 seconds
      reconnectTimeoutRef.current = window.setTimeout(() => {
        connect()
      }, 3000)
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }

    ws.onmessage = (event) => {
      try {
        const message: WSMessage = JSON.parse(event.data)
        handleMessage(message)
      } catch (e) {
        console.error('Failed to parse message:', e)
      }
    }

    wsRef.current = ws
  }, [roomId])

  const handleMessage = useCallback((message: WSMessage) => {
    switch (message.type) {
      case 'player_joined': {
        const payload = message.payload as PlayerJoinedPayload
        addPlayer(payload.player)
        break
      }
      case 'player_left': {
        const payload = message.payload as PlayerLeftPayload
        removePlayer(payload.user_id)
        break
      }
      case 'team_changed': {
        const payload = message.payload as TeamChangedPayload
        updatePlayerTeam(payload.user_id, payload.team)
        break
      }
      case 'game_started': {
        const payload = message.payload as GameStartedPayload
        if (room) {
          setRoom({
            ...room,
            status: 'playing',
            current_explainer_id: payload.explainer_id,
          })
        }
        setScreen('game')
        break
      }
      case 'new_word': {
        const payload = message.payload as NewWordPayload
        setCurrentWord({ id: payload.word_id, word: payload.word })
        break
      }
      case 'word_result': {
        // Word was guessed or missed, clear current word
        setCurrentWord(null)
        break
      }
      case 'timer': {
        const payload = message.payload as TimerPayload
        setSecondsLeft(payload.seconds_left)
        break
      }
      case 'round_end': {
        const payload = message.payload as RoundEndPayload
        setTeamScores(payload.team_scores)
        if (room) {
          setRoom({
            ...room,
            current_round: payload.round + 1,
            current_explainer_id: payload.next_explainer,
          })
        }
        setSecondsLeft(60)
        setCurrentWord(null)
        break
      }
      case 'game_end': {
        const payload = message.payload as GameEndPayload
        setTeamScores(payload.team_scores)
        if (room) {
          setRoom({ ...room, status: 'finished' })
        }
        setScreen('stats')
        break
      }
      case 'score_update': {
        const payload = message.payload as { team_scores: { A: number; B: number } }
        setTeamScores(payload.team_scores)
        break
      }
      case 'error': {
        const payload = message.payload as { message: string }
        console.error('Server error:', payload.message)
        break
      }
    }
  }, [room, addPlayer, removePlayer, updatePlayerTeam, setRoom, setCurrentWord, setSecondsLeft, setTeamScores, setScreen])

  const send = useCallback((type: string, payload?: Record<string, unknown>) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, ...(payload || {}) }))
    }
  }, [])

  const sendSwipe = useCallback((action: 'up' | 'down' | 'left' | 'right') => {
    send('swipe', { action })
  }, [send])

  useEffect(() => {
    connect()

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      wsRef.current?.close()
    }
  }, [connect])

  return {
    isConnected,
    send,
    sendSwipe,
  }
}
