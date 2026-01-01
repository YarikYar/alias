import { useSwipeGesture } from '../hooks/useSwipeGesture'

interface SwipeCardProps {
  word: string
  onSwipe: (direction: 'up' | 'down' | 'left' | 'right') => void
}

export default function SwipeCard({ word, onSwipe }: SwipeCardProps) {
  const { bind, x, y, direction, swiping } = useSwipeGesture({
    threshold: 80,
    onSwipeUp: () => onSwipe('up'),
    onSwipeDown: () => onSwipe('down'),
    onSwipeLeft: () => onSwipe('left'),
    onSwipeRight: () => onSwipe('right'),
  })

  // Calculate rotation based on x movement
  const rotate = x * 0.1

  // Calculate opacity based on swipe distance
  const opacity = swiping ? Math.min(Math.abs(y) / 100, 1) : 0

  // Background color based on direction
  const getBgOverlay = () => {
    if (!swiping) return 'transparent'
    if (direction === 'up') return `rgba(34, 197, 94, ${opacity})`
    if (direction === 'down') return `rgba(239, 68, 68, ${opacity})`
    return 'transparent'
  }

  return (
    <div className="relative w-full max-w-sm">
      {/* Swipe indicators */}
      <div
        className={`absolute inset-x-0 -top-16 text-center transition-opacity ${
          direction === 'up' && swiping ? 'opacity-100' : 'opacity-0'
        }`}
      >
        <span className="text-4xl">✅</span>
      </div>
      <div
        className={`absolute inset-x-0 -bottom-16 text-center transition-opacity ${
          direction === 'down' && swiping ? 'opacity-100' : 'opacity-0'
        }`}
      >
        <span className="text-4xl">❌</span>
      </div>

      {/* Card */}
      <div
        {...bind()}
        className="relative touch-none cursor-grab active:cursor-grabbing"
        style={{
          transform: `translate(${x}px, ${y}px) rotate(${rotate}deg)`,
          transition: swiping ? 'none' : 'transform 0.3s ease-out',
        }}
      >
        <div
          className="bg-white dark:bg-gray-800 rounded-3xl shadow-2xl p-8 aspect-[3/4] flex items-center justify-center"
          style={{
            background: getBgOverlay(),
          }}
        >
          <span className="text-4xl font-bold text-center break-words">
            {word}
          </span>
        </div>

        {/* Corner hints */}
        <div className="absolute top-4 left-4 text-2xl opacity-50">👆</div>
        <div className="absolute bottom-4 right-4 text-2xl opacity-50">👇</div>
      </div>
    </div>
  )
}
