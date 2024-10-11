module.exports = {
  content: [
    './internal/templates/**/*.html'
  ],
  theme: {
    extend: {
      colors: {
        neutral: {
          DEFAULT: '#f5f5f5',  // This will be used for bg-neutral
          light: '#fafafa',
          dark: '#e5e5e5',
        },
        // You can add more custom colors here
        customBlue: {
          DEFAULT: '#3490dc',
          light: '#6cb2eb',
          dark: '#2779bd',
        },
        // ... other custom colors
      },
    },
  },
  plugins: [
  require('daisyui'),
  ],
daisyui: {
    themes: [
      {
        mytheme: {
          "primary": "#570DF8",
          "secondary": "#F000B8",
          "accent": "#37CDBE",
          "neutral": "#3D4451",
          "base-100": "#FFFFFF",
          "base-200": "#F0F8FF",
          "base-300": "#E6F3FF",
        },
      },
    ],
  },
};
