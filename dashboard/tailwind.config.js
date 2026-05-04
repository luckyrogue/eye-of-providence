import preset from "@eop/ui/tailwind.preset.js";

/** @type {import('tailwindcss').Config} */
export default {
  presets: [preset],
  content: [
    "./index.html",
    "./src/**/*.{ts,tsx}",
    "../ui/src/**/*.{ts,tsx}",
  ],
};
