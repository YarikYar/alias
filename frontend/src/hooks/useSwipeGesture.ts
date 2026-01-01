import { useState, useCallback } from 'react'
import { useDrag } from '@use-gesture/react'
import { hapticFeedback } from '../lib/telegram'

interface SwipeConfig {
  threshold?: number
  onSwipeUp?: () => void
  onSwipeDown?: () => void
  onSwipeLeft?: () => void
  onSwipeRight?: () => void
}

interface SwipeState {
  x: number
  y: number
  direction: 'up' | 'down' | 'left' | 'right' | null
  swiping: boolean
}

export function useSwipeGesture(config: SwipeConfig) {
  const {
    threshold = 100,
    onSwipeUp,
    onSwipeDown,
    onSwipeLeft,
    onSwipeRight,
  } = config

  const [state, setState] = useState<SwipeState>({
    x: 0,
    y: 0,
    direction: null,
    swiping: false,
  })

  const handleSwipeComplete = useCallback((direction: 'up' | 'down' | 'left' | 'right') => {
    switch (direction) {
      case 'up':
        hapticFeedback('success')
        onSwipeUp?.()
        break
      case 'down':
        hapticFeedback('error')
        onSwipeDown?.()
        break
      case 'left':
        hapticFeedback('light')
        onSwipeLeft?.()
        break
      case 'right':
        hapticFeedback('warning')
        onSwipeRight?.()
        break
    }
  }, [onSwipeUp, onSwipeDown, onSwipeLeft, onSwipeRight])

  const bind = useDrag(
    ({ movement: [mx, my], down, velocity: [vx, vy] }) => {
      if (!down) {
        // Check if swipe exceeded threshold
        const absX = Math.abs(mx)
        const absY = Math.abs(my)
        const maxVelocity = Math.max(Math.abs(vx), Math.abs(vy))

        // Either distance or velocity can trigger swipe
        const triggerByDistance = Math.max(absX, absY) > threshold
        const triggerByVelocity = maxVelocity > 0.5

        if (triggerByDistance || triggerByVelocity) {
          if (absY > absX) {
            // Vertical swipe
            if (my < 0) {
              handleSwipeComplete('up')
            } else {
              handleSwipeComplete('down')
            }
          } else {
            // Horizontal swipe
            if (mx < 0) {
              handleSwipeComplete('left')
            } else {
              handleSwipeComplete('right')
            }
          }
        }

        // Reset state
        setState({ x: 0, y: 0, direction: null, swiping: false })
        return
      }

      // Determine direction during drag
      let direction: 'up' | 'down' | 'left' | 'right' | null = null
      const absX = Math.abs(mx)
      const absY = Math.abs(my)

      if (absY > absX) {
        direction = my < 0 ? 'up' : 'down'
      } else if (absX > absY) {
        direction = mx < 0 ? 'left' : 'right'
      }

      // Haptic on direction change
      if (direction && direction !== state.direction) {
        hapticFeedback('light')
      }

      setState({
        x: mx,
        y: my,
        direction,
        swiping: true,
      })
    },
    {
      filterTaps: true,
      rubberband: true,
    }
  )

  return {
    bind,
    ...state,
  }
}
