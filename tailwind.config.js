const colors = require("tailwindcss/colors");

module.exports = {
  content: [

    "internal/templates/**/*.templ",
    "internal/templates/*.templ",
    "internal/templates/*.go",
    "internal/templates/*.templ.txt",
  ],
  theme: {
    extend: {
      colors: {
        ...colors
      }
    }
  },
  plugins: [require("@tailwindcss/forms"), require("@tailwindcss/typography")],
}
