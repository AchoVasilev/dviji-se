/** @type {import('tailwindcss').Config} */
export default {
  darkMode: "class",
  content: ["./web/templates/**/*.templ", "./web/templates/*.templ"],
  theme: {
    extend: {
      colors: {
        primary: "#DC2626",
        accent: "#2563EB",
        "bg-light": "#F8FAFC",
        "bg-dark": "#0F172A",
        "card-dark": "#1E293B",
      },
      fontFamily: {
        sans: ["Inter", "sans-serif"],
        display: ["Inter", "sans-serif"],
      },
      borderRadius: {
        DEFAULT: "0.75rem",
      },
    },
  },
  plugins: [],
};
