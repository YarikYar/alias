import { useGameStore } from '../stores/gameStore'
import SwipeCard from '../components/SwipeCard'
import CircularTimer from '../components/CircularTimer'
import ScoreBoard from '../components/ScoreBoard'

export default function Game() {
  const {
    room,
    currentWord,
    secondsLeft,
    teamScores,
    isExplainer,
    sendSwipe,
  } = useGameStore()

  const amExplainer = isExplainer()

  const handleSwipe = (direction: 'up' | 'down' | 'left' | 'right') => {
    if (amExplainer && sendSwipe) {
      sendSwipe(direction)
    }
  }

  const handleLeave = () => {
    localStorage.removeItem('currentRoomId')
    window.location.reload()
  }

  if (!room) return null

  return (
    <div className="flex flex-col h-full safe-area-top safe-area-bottom">
      {/* Header with exit button */}
      <div className="p-2 flex justify-end">
        <button
          onClick={handleLeave}
          className="px-3 py-1 bg-red-500/10 text-red-500 rounded-lg text-xs font-medium"
        >
          Выйти
        </button>
      </div>

      {/* Score */}
      <div className="px-4 pb-4">
        <ScoreBoard teamA={teamScores.A} teamB={teamScores.B} />
      </div>

      {/* Timer */}
      <div className="flex-shrink-0 flex justify-center py-4">
        <CircularTimer seconds={secondsLeft} total={60} />
      </div>

      {/* Role indicator */}
      <div className="text-center py-2">
        {amExplainer ? (
          <div className="inline-block px-4 py-2 bg-green-500 text-white rounded-full font-bold">
            Ты объясняешь!
          </div>
        ) : (
          <div className="inline-block px-4 py-2 bg-tg-secondary rounded-full">
            Отгадывай!
          </div>
        )}
      </div>

      {/* Word card (only for explainer) */}
      <div className="flex-1 flex items-center justify-center p-4">
        {amExplainer ? (
          <SwipeCard
            word={currentWord?.word || 'Ждём слово...'}
            onSwipe={handleSwipe}
          />
        ) : (
          <div className="text-center">
            <div className="text-6xl mb-4">🤔</div>
            <p className="text-xl text-tg-hint">
              Слушай объяснение и угадывай!
            </p>
          </div>
        )}
      </div>

      {/* Swipe hints */}
      {amExplainer && (
        <div className="p-4 grid grid-cols-2 gap-4 text-center text-sm">
          <div className="bg-green-500/10 text-green-500 rounded-lg py-2">
            ↑ Угадал
          </div>
          <div className="bg-red-500/10 text-red-500 rounded-lg py-2">
            ↓ Не угадал
          </div>
        </div>
      )}
    </div>
  )
}
