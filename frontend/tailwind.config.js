/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        tg: {
          bg: 'var(--tg-theme-bg-color)',
          text: 'var(--tg-theme-text-color)',
          hint: 'var(--tg-theme-hint-color)',
          link: 'var(--tg-theme-link-color)',
          button: 'var(--tg-theme-button-color)',
          buttonText: 'var(--tg-theme-button-text-color)',
          secondary: 'var(--tg-theme-secondary-bg-color)',
        },
        team: {
          a: '#ef4444',
          b: '#3b82f6',
        },
      },
      animation: {
        'swipe-up': 'swipeUp 0.3s ease-out',
        'swipe-down': 'swipeDown 0.3s ease-out',
        'pulse-slow': 'pulse 2s infinite',
      },
      keyframes: {
        swipeUp: {
          '0%': { transform: 'translateY(0)', opacity: '1' },
          '100%': { transform: 'translateY(-100px)', opacity: '0' },
        },
        swipeDown: {
          '0%': { transform: 'translateY(0)', opacity: '1' },
          '100%': { transform: 'translateY(100px)', opacity: '0' },
        },
      },
    },
  },
  plugins: [],
}
