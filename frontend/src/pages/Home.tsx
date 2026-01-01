import { useState, useEffect } from 'react'
import { getStartParam } from '../lib/telegram'

interface HomeProps {
  onCreateRoom: (category: string, numTeams: number) => void
}

const CATEGORIES = [
  { id: 'general', name: 'Общие', icon: '🎯' },
  { id: 'animals', name: 'Животные', icon: '🦁' },
  { id: 'food', name: 'Еда', icon: '🍕' },
  { id: 'countries', name: 'Страны', icon: '🌍' },
  { id: 'professions', name: 'Профессии', icon: '👨‍💼' },
]

export default function Home({ onCreateRoom }: HomeProps) {
  const [debugInfo, setDebugInfo] = useState<string>('')
  const [selectedCategory, setSelectedCategory] = useState<string>('general')
  const [numTeams, setNumTeams] = useState<number>(2)
  const startParam = getStartParam()

  useEffect(() => {
    // Get debug info from session storage
    const info = sessionStorage.getItem('debugInfo')
    if (info) {
      setDebugInfo(info)
    }
  }, [])

  return (
    <div className="flex flex-col items-center justify-center h-full p-6 safe-area-top safe-area-bottom">
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold mb-2">Elias</h1>
        <p className="text-tg-hint text-lg">Объясняй слова друзьям</p>
        {startParam && (
          <p className="text-xs text-red-500 mt-2">Debug: startParam = {startParam}</p>
        )}
        {debugInfo && (
          <p className="text-xs text-orange-500 mt-2 max-w-xs break-words">{debugInfo}</p>
        )}
      </div>

      <div className="w-full max-w-md space-y-4">
        {/* Category selection */}
        <div>
          <p className="text-center text-tg-hint text-sm mb-3">Выбери тематику:</p>
          <div className="grid grid-cols-2 gap-2">
            {CATEGORIES.map((cat) => (
              <button
                key={cat.id}
                onClick={() => setSelectedCategory(cat.id)}
                className={`py-3 px-4 rounded-xl font-medium text-sm transition-all ${
                  selectedCategory === cat.id
                    ? 'bg-tg-button text-tg-buttonText shadow-lg scale-105'
                    : 'bg-tg-secondary text-tg-text'
                }`}
              >
                <div className="text-2xl mb-1">{cat.icon}</div>
                {cat.name}
              </button>
            ))}
          </div>
        </div>

        {/* Number of teams selection */}
        <div>
          <p className="text-center text-tg-hint text-sm mb-3">Количество команд:</p>
          <div className="flex gap-2 justify-center">
            {[2, 3, 4, 5].map((n) => (
              <button
                key={n}
                onClick={() => setNumTeams(n)}
                className={`w-12 h-12 rounded-xl font-bold text-lg transition-all ${
                  numTeams === n
                    ? 'bg-tg-button text-tg-buttonText shadow-lg scale-110'
                    : 'bg-tg-secondary text-tg-text'
                }`}
              >
                {n}
              </button>
            ))}
          </div>
        </div>

        <button
          onClick={() => onCreateRoom(selectedCategory, numTeams)}
          className="w-full py-4 px-6 bg-tg-button text-tg-buttonText rounded-xl font-semibold text-lg shadow-lg active:scale-95 transition-transform"
        >
          Создать комнату
        </button>

        <p className="text-center text-tg-hint text-sm">
          Или присоединись по ссылке от друга
        </p>
      </div>

      <div className="mt-12 text-center">
        <div className="grid grid-cols-2 gap-6 text-sm">
          <div>
            <div className="text-2xl mb-1">👆</div>
            <p className="text-tg-hint">Свайп вверх</p>
            <p className="font-medium text-green-500">Угадал!</p>
          </div>
          <div>
            <div className="text-2xl mb-1">👇</div>
            <p className="text-tg-hint">Свайп вниз</p>
            <p className="font-medium text-red-500">Не угадал</p>
          </div>
        </div>
      </div>
    </div>
  )
}
