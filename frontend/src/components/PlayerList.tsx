import type { Player } from '../types'

interface PlayerListProps {
  players: Player[]
}

export default function PlayerList({ players }: PlayerListProps) {
  if (players.length === 0) {
    return (
      <div className="text-center py-4 text-tg-hint text-sm">
        Нет игроков
      </div>
    )
  }

  return (
    <div className="flex flex-wrap gap-2">
      {players.map((player) => (
        <div
          key={player.user_id}
          className="flex items-center gap-2 px-3 py-2 bg-tg-secondary rounded-full"
        >
          <div className="w-6 h-6 rounded-full bg-tg-button flex items-center justify-center text-tg-buttonText text-xs font-bold">
            {player.first_name?.[0] || player.username?.[0] || '?'}
          </div>
          <span className="text-sm font-medium">
            {player.first_name || player.username || 'Игрок'}
          </span>
          {player.is_host && (
            <span className="text-xs">👑</span>
          )}
        </div>
      ))}
    </div>
  )
}
