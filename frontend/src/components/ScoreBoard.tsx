interface ScoreBoardProps {
  teamScores: Record<string, number>
}

const TEAM_COLORS = ['text-blue-500', 'text-red-500', 'text-green-500', 'text-yellow-500', 'text-purple-500']

export default function ScoreBoard({ teamScores }: ScoreBoardProps) {
  const teams = Object.entries(teamScores).sort((a, b) => a[0].localeCompare(b[0]))

  return (
    <div className="flex items-center justify-center gap-2 flex-wrap">
      {teams.map(([team, score], index) => (
        <div key={team} className="flex items-center gap-2">
          <div className="text-center">
            <div className={`text-sm font-medium ${TEAM_COLORS[index % TEAM_COLORS.length]} mb-1`}>
              {team}
            </div>
            <div className={`text-3xl font-bold ${TEAM_COLORS[index % TEAM_COLORS.length]}`}>
              {score}
            </div>
          </div>
          {index < teams.length - 1 && <div className="text-xl text-tg-hint">:</div>}
        </div>
      ))}
    </div>
  )
}
