import { useEffect } from 'react'
import { useGameStore } from '../stores/gameStore'
import { changeTeam, startGame } from '../lib/api'
import { shareRoom, showMainButton, hideMainButton } from '../lib/telegram'
import PlayerList from '../components/PlayerList'
import TeamSelector from '../components/TeamSelector'

export default function Lobby() {
  const { room, players, isHost, getMyPlayer } = useGameStore()
  const isConnected = true // WebSocket is managed in App.tsx

  const myPlayer = getMyPlayer()
  const canStart = players.filter(p => p.team).length >= 2

  useEffect(() => {
    if (isHost() && room) {
      if (canStart) {
        showMainButton('Начать игру', handleStart)
      } else {
        hideMainButton()
      }
    }

    return () => hideMainButton()
  }, [canStart, room])

  const handleTeamChange = async (team: string) => {
    if (!room) return
    try {
      await changeTeam(room.id, team)
    } catch (e) {
      console.error('Failed to change team:', e)
    }
  }

  const handleStart = async () => {
    if (!room) return
    try {
      await startGame(room.id)
    } catch (e) {
      console.error('Failed to start game:', e)
    }
  }

  const handleShare = () => {
    if (room) {
      shareRoom(room.id)
    }
  }

  const handleLeave = () => {
    localStorage.removeItem('currentRoomId')
    window.location.reload()
  }

  if (!room) return null

  return (
    <div className="flex flex-col h-full safe-area-top safe-area-bottom">
      {/* Header */}
      <div className="p-4 border-b border-tg-secondary">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold">Лобби</h1>
            <p className="text-sm text-tg-hint">
              {isConnected ? '🟢 Подключено' : '🔴 Подключение...'}
            </p>
          </div>
          <div className="flex gap-2">
            <button
              onClick={handleLeave}
              className="px-4 py-2 bg-red-500/10 text-red-500 rounded-lg text-sm font-medium"
            >
              Выйти
            </button>
            <button
              onClick={handleShare}
              className="px-4 py-2 bg-tg-secondary rounded-lg text-sm font-medium"
            >
              Пригласить
            </button>
          </div>
        </div>
      </div>

      {/* Teams */}
      <div className="flex-1 overflow-y-auto p-4">
        <div className={`grid gap-4 mb-6 ${room.num_teams <= 2 ? 'grid-cols-2' : room.num_teams === 3 ? 'grid-cols-3' : 'grid-cols-2'}`}>
          {room.team_names?.map((teamName, i) => (
            <TeamSelector
              key={teamName}
              team={teamName}
              teamIndex={i}
              players={players.filter(p => p.team === teamName)}
              isSelected={myPlayer?.team === teamName}
              onSelect={() => handleTeamChange(teamName)}
            />
          ))}
        </div>

        {/* Unassigned players */}
        <div className="mb-6">
          <h3 className="text-sm font-medium text-tg-hint mb-2">Без команды</h3>
          <PlayerList players={players.filter(p => !p.team)} />
        </div>

        {/* Info */}
        <div className="bg-tg-secondary rounded-xl p-4 text-center">
          <p className="text-sm text-tg-hint">
            {players.length} / 8 игроков
          </p>
          {!canStart && (
            <p className="text-sm text-tg-hint mt-2">
              Минимум 2 игрока в командах для старта
            </p>
          )}
        </div>
      </div>

      {/* Start button (for host) */}
      {isHost() && canStart && (
        <div className="p-4 border-t border-tg-secondary">
          <button
            onClick={handleStart}
            className="w-full py-4 bg-tg-button text-tg-buttonText rounded-xl font-semibold text-lg"
          >
            Начать игру
          </button>
        </div>
      )}
    </div>
  )
}
