/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./ui/html/*.tmpl.html",
    "./ui/html/**/*.tmpl.html",
    "./ui/html/**/**/*.tmpl.html",
    "./ui/dashboard/html/*.tmpl.html",
    "./ui/dashboard/html/**/*.tmpl.html",
    "./ui/dashboard/html/**/**/*.tmpl.html",
  ],
  theme: {
    extend: {
      fontFamily: {
        readex: ['"Readex Pro"', "sans-serif"], // Add your custom font here
      },
    },
  },
  plugins: [],
};
