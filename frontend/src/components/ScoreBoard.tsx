interface ScoreBoardProps {
  teamA: number
  teamB: number
}

export default function ScoreBoard({ teamA, teamB }: ScoreBoardProps) {
  return (
    <div className="flex items-center justify-center gap-4">
      <div className="flex-1 text-center">
        <div className="text-sm font-medium text-team-a mb-1">Команда A</div>
        <div className="text-4xl font-bold text-team-a">{teamA}</div>
      </div>

      <div className="text-2xl text-tg-hint">:</div>

      <div className="flex-1 text-center">
        <div className="text-sm font-medium text-team-b mb-1">Команда B</div>
        <div className="text-4xl font-bold text-team-b">{teamB}</div>
      </div>
    </div>
  )
}
