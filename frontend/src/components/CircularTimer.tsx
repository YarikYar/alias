interface CircularTimerProps {
  seconds: number
  total: number
}

export default function CircularTimer({ seconds, total }: CircularTimerProps) {
  const radius = 60
  const circumference = 2 * Math.PI * radius
  const progress = seconds / total
  const strokeDashoffset = circumference * (1 - progress)

  const getColor = () => {
    if (seconds <= 10) return '#ef4444' // red
    if (seconds <= 30) return '#f59e0b' // amber
    return '#22c55e' // green
  }

  return (
    <div className="relative inline-flex items-center justify-center">
      <svg className="transform -rotate-90" width="140" height="140">
        {/* Background circle */}
        <circle
          cx="70"
          cy="70"
          r={radius}
          fill="transparent"
          stroke="currentColor"
          strokeWidth="8"
          className="text-tg-secondary"
        />
        {/* Progress circle */}
        <circle
          cx="70"
          cy="70"
          r={radius}
          fill="transparent"
          stroke={getColor()}
          strokeWidth="8"
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={strokeDashoffset}
          className="transition-all duration-1000 ease-linear"
        />
      </svg>
      {/* Time display */}
      <div className="absolute flex flex-col items-center">
        <span
          className="text-4xl font-bold tabular-nums"
          style={{ color: getColor() }}
        >
          {seconds}
        </span>
        <span className="text-sm text-tg-hint">сек</span>
      </div>
    </div>
  )
}
