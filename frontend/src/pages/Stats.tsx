import { useEffect, useState } from 'react'
import { useGameStore } from '../stores/gameStore'
import { getStats } from '../lib/api'
import type { GameStats } from '../types'

export default function Stats() {
  const { room, teamScores, setScreen } = useGameStore()
  const [stats, setStats] = useState<GameStats | null>(null)

  useEffect(() => {
    if (room) {
      getStats(room.id).then(setStats).catch(console.error)
    }
  }, [room])

  const teams = Object.entries(teamScores).sort((a, b) => b[1] - a[1])
  const winner = teams.length > 0 && teams[0][1] > teams[1][1] ? teams[0][0] : null
  const TEAM_COLORS = ['bg-blue-500', 'bg-red-500', 'bg-green-500', 'bg-yellow-500', 'bg-purple-500']
  const TEAM_TEXT_COLORS = ['text-blue-500', 'text-red-500', 'text-green-500', 'text-yellow-500', 'text-purple-500']

  const handleNewGame = () => {
    setScreen('home')
  }

  return (
    <div className="flex flex-col h-full safe-area-top safe-area-bottom">
      {/* Winner banner */}
      <div className={`p-8 text-center ${winner ? TEAM_COLORS[['A', 'B', 'C', 'D', 'E'].indexOf(winner)] : 'bg-tg-secondary'}`}>
        <h1 className="text-3xl font-bold text-white mb-2">
          {winner ? `Команда ${winner} победила!` : 'Ничья!'}
        </h1>
        <p className="text-white/80 text-xl">
          {teams.map(([team, score]) => `${team}: ${score}`).join(' | ')}
        </p>
      </div>

      {/* Stats */}
      <div className="flex-1 overflow-y-auto p-4">
        {/* Team scores */}
        <div className={`grid gap-4 mb-6 ${teams.length <= 2 ? 'grid-cols-2' : teams.length === 3 ? 'grid-cols-3' : 'grid-cols-2'}`}>
          {teams.map(([team, score], index) => (
            <div key={team} className={`p-4 rounded-xl text-center ${winner === team ? `${TEAM_COLORS[index]}/20 border-2 ${TEAM_COLORS[index].replace('bg-', 'border-')}` : 'bg-tg-secondary'}`}>
              <div className="text-sm text-tg-hint mb-1">Команда {team}</div>
              <div className={`text-3xl font-bold ${TEAM_TEXT_COLORS[index]}`}>{score}</div>
            </div>
          ))}
        </div>

        {/* Player stats */}
        {stats && (
          <div className="mb-6">
            <h3 className="text-lg font-semibold mb-3">Игроки</h3>
            <div className="space-y-2">
              {stats.players
                .sort((a, b) => b.score - a.score)
                .map((player) => (
                  <div
                    key={player.user_id}
                    className="flex items-center justify-between p-3 bg-tg-secondary rounded-lg"
                  >
                    <div className="flex items-center gap-3">
                      <div className={`w-8 h-8 rounded-full flex items-center justify-center text-white font-bold ${TEAM_COLORS[['A', 'B', 'C', 'D', 'E'].indexOf(player.team)]}`}>
                        {player.first_name?.[0] || '?'}
                      </div>
                      <div>
                        <div className="font-medium">{player.first_name}</div>
                        <div className="text-xs text-tg-hint">Команда {player.team}</div>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="font-bold">{player.score}</div>
                      <div className="text-xs text-tg-hint">очков</div>
                    </div>
                  </div>
                ))}
            </div>
          </div>
        )}
      </div>

      {/* New game button */}
      <div className="p-4 border-t border-tg-secondary">
        <button
          onClick={handleNewGame}
          className="w-full py-4 bg-tg-button text-tg-buttonText rounded-xl font-semibold text-lg"
        >
          Новая игра
        </button>
      </div>
    </div>
  )
}
