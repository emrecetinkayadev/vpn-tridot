const config = {
  content: [
    "./app/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}"
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          DEFAULT: "#1e293b",
          accent: "#38bdf8"
        }
      }
    }
  },
  plugins: []
};

export default config;
