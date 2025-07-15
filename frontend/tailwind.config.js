/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./src/**/*.{astro,html,md,mdx,ts,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                border: "hsl(var(--border))",
                input: "hsl(var(--input))",
                ring: "hsl(var(--ring))",
                background: "hsl(var(--background))",
                foreground: "hsl(var(--foreground))",
                primary: {
                    DEFAULT: "hsl(var(--primary))",
                    foreground: "hsl(var(--primary-foreground))",
                },
                secondary: {
                    DEFAULT: "hsl(var(--secondary))",
                    foreground: "hsl(var(--secondary-foreground))",
                },
                destructive: {
                    DEFAULT: "hsl(var(--destructive))",
                    foreground: "hsl(var(--destructive-foreground))",
                },
                muted: {
                    DEFAULT: "hsl(var(--muted))",
                    foreground: "hsl(var(--muted-foreground))",
                },
                accent: {
                    DEFAULT: "hsl(var(--accent))",
                    foreground: "hsl(var(--accent-foreground))",
                },
                popover: {
                    DEFAULT: "hsl(var(--popover))",
                    foreground: "hsl(var(--popover-foreground))",
                },
                card: {
                    DEFAULT: "hsl(var(--card))",
                    foreground: "hsl(var(--card-foreground))",
                },
                shamrock: {
                    50: 'hsl(var(--shamrock-50) / <alpha-value>)',
                    100: 'hsl(var(--shamrock-100) / <alpha-value>)',
                    200: 'hsl(var(--shamrock-200) / <alpha-value>)',
                    300: 'hsl(var(--shamrock-300) / <alpha-value>)',
                    400: 'hsl(var(--shamrock-400) / <alpha-value>)',
                    500: 'hsl(var(--shamrock-500) / <alpha-value>)',
                    600: 'hsl(var(--shamrock-600) / <alpha-value>)',
                    700: 'hsl(var(--shamrock-700) / <alpha-value>)',
                    800: 'hsl(var(--shamrock-800) / <alpha-value>)',
                    900: 'hsl(var(--shamrock-900) / <alpha-value>)',
                    950: 'hsl(var(--shamrock-950) / <alpha-value>)',
                },
            },
            borderRadius: {
                lg: "var(--radius)",
                md: "calc(var(--radius) - 2px)",
                sm: "calc(var(--radius) - 4px)",
            },
        },
    },
    plugins: [],
} 