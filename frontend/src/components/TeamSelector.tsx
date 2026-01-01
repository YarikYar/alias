import type { Player } from '../types'

interface TeamSelectorProps {
  team: string
  teamIndex: number
  players: Player[]
  isSelected: boolean
  onSelect: () => void
}

const TEAM_COLORS = [
  { bg: 'bg-blue-500', border: 'border-blue-500', text: 'text-blue-500' },
  { bg: 'bg-red-500', border: 'border-red-500', text: 'text-red-500' },
  { bg: 'bg-green-500', border: 'border-green-500', text: 'text-green-500' },
  { bg: 'bg-yellow-500', border: 'border-yellow-500', text: 'text-yellow-500' },
  { bg: 'bg-purple-500', border: 'border-purple-500', text: 'text-purple-500' },
]

export default function TeamSelector({ team, teamIndex, players, isSelected, onSelect }: TeamSelectorProps) {
  const colors = TEAM_COLORS[teamIndex % TEAM_COLORS.length]
  const bgColor = colors.bg
  const borderColor = colors.border
  const textColor = colors.text

  return (
    <button
      onClick={onSelect}
      className={`p-4 rounded-xl border-2 transition-all ${
        isSelected
          ? `${borderColor} ${bgColor}/10`
          : 'border-tg-secondary bg-tg-secondary'
      }`}
    >
      <div className="flex items-center justify-between mb-3">
        <h3 className={`font-bold text-lg ${isSelected ? textColor : ''}`}>
          Команда {team}
        </h3>
        <span className="text-sm text-tg-hint">{players.length}</span>
      </div>

      <div className="space-y-1">
        {players.length === 0 ? (
          <p className="text-sm text-tg-hint">Нажми, чтобы присоединиться</p>
        ) : (
          players.map((player) => (
            <div
              key={player.user_id}
              className="flex items-center gap-2 text-sm"
            >
              <div className={`w-5 h-5 rounded-full ${bgColor} flex items-center justify-center text-white text-xs font-bold`}>
                {player.first_name?.[0] || '?'}
              </div>
              <span className="truncate">
                {player.first_name || player.username || 'Игрок'}
              </span>
              {player.is_host && <span className="text-xs">👑</span>}
            </div>
          ))
        )}
      </div>
    </button>
  )
}
