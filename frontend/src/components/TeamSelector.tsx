import type { Player } from '../types'

interface TeamSelectorProps {
  team: 'A' | 'B'
  players: Player[]
  isSelected: boolean
  onSelect: () => void
}

export default function TeamSelector({ team, players, isSelected, onSelect }: TeamSelectorProps) {
  const bgColor = team === 'A' ? 'bg-team-a' : 'bg-team-b'
  const borderColor = team === 'A' ? 'border-team-a' : 'border-team-b'
  const textColor = team === 'A' ? 'text-team-a' : 'text-team-b'

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
