/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Exact colors from Kotlin AppTheme.kt
        primary: {
          DEFAULT: '#1976D2',
          50: '#E3F2FD',
          100: '#BBDEFB',
          500: '#1976D2',
          600: '#1565C0',
          700: '#0D47A1'
        },
        secondary: {
          DEFAULT: '#2196F3',
          50: '#E1F5FE',
          100: '#B3E5FC',
          500: '#2196F3',
          600: '#1976D2'
        },
        surface: '#FAFAFA',
        background: '#F5F5F5',
        'on-primary': '#FFFFFF',
        'on-secondary': '#FFFFFF',
        'on-background': '#1A1A1A',
        'on-surface': '#1A1A1A',
        'surface-variant': '#F0F0F0',
        'on-surface-variant': '#424242'
      },
      fontSize: {
        // Exact typography from Kotlin AppTheme.kt
        'title-large': ['16px', { lineHeight: '24px' }],
        'title-medium': ['13px', { lineHeight: '20px' }],
        'body-large': ['13px', { lineHeight: '20px' }],
        'body-medium': ['11px', { lineHeight: '16px' }],
        'label-large': ['12px', { lineHeight: '16px' }],
      },
      spacing: {
        // Exact dimensions from Kotlin Dimensions.kt
        'button-height': '32px',
        'spacing-small': '4px',
        'spacing-medium': '8px',
        'spacing-large': '16px',
        'drop-zone-height': '150px',  // Matches Kotlin 150.dp
      },
      height: {
        'button': '32px',  // Dimensions.buttonHeight
        'drop-zone': '150px',
      },
      fontFamily: {
        'mono': ['Monaco', 'Menlo', 'Ubuntu Mono', 'monospace'], // For console
      },
      animation: {
        'progress': 'progress 2s ease-in-out infinite',
        'pulse-border': 'pulse-border 2s ease-in-out infinite',
      },
      keyframes: {
        'progress': {
          '0%': { transform: 'translateX(-100%)' },
          '100%': { transform: 'translateX(100%)' }
        },
        'pulse-border': {
          '0%, 100%': { borderColor: '#1976D2', borderWidth: '2px' },
          '50%': { borderColor: '#2196F3', borderWidth: '3px' }
        }
      },
      gridTemplateColumns: {
        'main': '40% 60%',  // Exact split from Kotlin Row(weight(0.4f), weight(0.6f))
      }
    },
  },
  darkMode: 'class',
  plugins: [],
}