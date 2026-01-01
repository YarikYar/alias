import { useEffect } from 'react'
import { useGameStore } from './stores/gameStore'
import { useWebSocket } from './hooks/useWebSocket'
import { initTelegram, getUser, getStartParam } from './lib/telegram'
import { getRoom, joinRoom, createRoom } from './lib/api'
import Home from './pages/Home'
import Lobby from './pages/Lobby'
import Game from './pages/Game'
import Stats from './pages/Stats'

function App() {
  const { screen, setScreen, setUser, setRoom, setPlayers, room, setSendSwipe } = useGameStore()

  // Single WebSocket connection for the entire app
  const { sendSwipe } = useWebSocket(room?.id || null)

  useEffect(() => {
    setSendSwipe(sendSwipe)
  }, [sendSwipe, setSendSwipe])

  useEffect(() => {
    initTelegram()

    const user = getUser()
    console.log('[App] User:', user)
    if (user) {
      setUser(user)
    }

    const startParam = getStartParam()
    console.log('[App] Start param:', startParam)

    async function init() {
      // Try to get room ID from startParam or localStorage
      const savedRoomId = localStorage.getItem('currentRoomId')
      const roomId = startParam || savedRoomId

      if (roomId) {
        // Join/reconnect to room
        console.log('[App] Attempting to join room:', roomId)
        try {
          const roomData = await getRoom(roomId)
          console.log('[App] Room data:', roomData)
          setRoom(roomData.room)
          setPlayers(roomData.players)

          // Save room ID for reconnection
          localStorage.setItem('currentRoomId', roomId)

          // Check if already in room
          const isInRoom = user && roomData.players.some(p => p.user_id === user.id)
          console.log('[App] Is in room:', isInRoom)
          if (!isInRoom && user) {
            console.log('[App] Joining room...')
            await joinRoom(roomId)
            // Refresh room data
            const updatedRoom = await getRoom(roomId)
            setPlayers(updatedRoom.players)
          }

          if (roomData.room.status === 'playing') {
            setScreen('game')
          } else if (roomData.room.status === 'finished') {
            setScreen('stats')
          } else {
            setScreen('lobby')
          }
        } catch (e) {
          console.error('[App] Failed to join room:', e)
          const errorMsg = e instanceof Error ? e.message : String(e)
          sessionStorage.setItem('debugInfo', `Error joining room: ${errorMsg}`)
          // Clear saved room if it doesn't exist
          localStorage.removeItem('currentRoomId')
          setScreen('home')
        }
      } else {
        console.log('[App] No room to join, showing home')
        setScreen('home')
      }
    }

    init()
  }, [setScreen, setUser, setRoom, setPlayers])

  const handleCreateRoom = async (category: string = 'general', numTeams: number = 2) => {
    try {
      const { room, player } = await createRoom(category, numTeams)
      setRoom(room)
      setPlayers([player])
      // Save room ID for reconnection
      localStorage.setItem('currentRoomId', room.id)
      setScreen('lobby')
    } catch (e) {
      console.error('Failed to create room:', e)
    }
  }

  switch (screen) {
    case 'loading':
      return (
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <div className="animate-spin w-8 h-8 border-4 border-tg-button border-t-transparent rounded-full mx-auto mb-4" />
            <p className="text-tg-hint">Загрузка...</p>
          </div>
        </div>
      )
    case 'home':
      return <Home onCreateRoom={handleCreateRoom} />
    case 'lobby':
      return <Lobby />
    case 'game':
      return <Game />
    case 'stats':
      return <Stats />
    default:
      return null
  }
}

export default App
