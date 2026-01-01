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

  const winner = teamScores.A > teamScores.B ? 'A' : teamScores.B > teamScores.A ? 'B' : null

  const handleNewGame = () => {
    setScreen('home')
  }

  return (
    <div className="flex flex-col h-full safe-area-top safe-area-bottom">
      {/* Winner banner */}
      <div className={`p-8 text-center ${winner === 'A' ? 'bg-team-a' : winner === 'B' ? 'bg-team-b' : 'bg-tg-secondary'}`}>
        <h1 className="text-3xl font-bold text-white mb-2">
          {winner ? `Команда ${winner} победила!` : 'Ничья!'}
        </h1>
        <p className="text-white/80 text-xl">
          {teamScores.A} : {teamScores.B}
        </p>
      </div>

      {/* Stats */}
      <div className="flex-1 overflow-y-auto p-4">
        {/* Team scores */}
        <div className="grid grid-cols-2 gap-4 mb-6">
          <div className={`p-4 rounded-xl text-center ${winner === 'A' ? 'bg-team-a/20 border-2 border-team-a' : 'bg-tg-secondary'}`}>
            <div className="text-sm text-tg-hint mb-1">Команда A</div>
            <div className="text-3xl font-bold text-team-a">{teamScores.A}</div>
          </div>
          <div className={`p-4 rounded-xl text-center ${winner === 'B' ? 'bg-team-b/20 border-2 border-team-b' : 'bg-tg-secondary'}`}>
            <div className="text-sm text-tg-hint mb-1">Команда B</div>
            <div className="text-3xl font-bold text-team-b">{teamScores.B}</div>
          </div>
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
                      <div className={`w-8 h-8 rounded-full flex items-center justify-center text-white font-bold ${player.team === 'A' ? 'bg-team-a' : 'bg-team-b'}`}>
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
