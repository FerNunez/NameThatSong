
export const content = [
  "web/templates/**/*.templ",
  "web/templates/*.templ",
  "web/templates/*.go",
  "web/templates/*.templ.txt",
  "static/**/*.js",
];

export const theme = {
  extend: {
    colors: {
      'forest': {
        50: '#96c5a9',   // Light green text
        100: '#366348',  // Border color
        200: '#264532',  // Accent background
        300: '#1b3124',  // Secondary background
        400: '#122118',  // Primary background
      },
      'accent': '#38e07b', // Bright green
    },
    fontFamily: {
      'sans': ['Spline Sans', 'Noto Sans', 'sans-serif'],
    },
    animation: {
      'fade-in': 'fadeIn 0.3s ease-in-out',
      'slide-up': 'slideUp 0.3s ease-out',
      'pulse-soft': 'pulseSoft 2s ease-in-out infinite',
    },
    keyframes: {
      fadeIn: {
        '0%': { opacity: '0' },
        '100%': { opacity: '1' },
      },
      slideUp: {
        '0%': { transform: 'translateY(10px)', opacity: '0' },
        '100%': { transform: 'translateY(0)', opacity: '1' },
      },
      pulseSoft: {
        '0%, 100%': { opacity: '1' },
        '50%': { opacity: '0.8' },
      },
    },
  }
};

export const plugins = [require("@tailwindcss/forms"), require("@tailwindcss/typography")];
