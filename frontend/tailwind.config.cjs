/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          blue: '#002FA7',
          'blue-dark': '#001f7a',
          'blue-light': '#e6f0ff',
          bg: '#f8faff',
        },
      },
    },
  },
  plugins: [],
}