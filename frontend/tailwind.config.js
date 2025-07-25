/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./src/**/*.{astro,html,md,mdx,ts,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                // Material 3 Token 映射
                border: "hsl(var(--border))",
                input: "hsl(var(--input))",
                ring: "hsl(var(--ring))",
                background: "hsl(var(--background))",
                foreground: "hsl(var(--foreground))",
                
                // 主色调系统
                primary: {
                    DEFAULT: "hsl(var(--primary))",
                    foreground: "hsl(var(--primary-foreground))",
                },
                secondary: {
                    DEFAULT: "hsl(var(--secondary))",
                    foreground: "hsl(var(--secondary-foreground))",
                },
                
                // 语义化状态色彩
                destructive: {
                    DEFAULT: "hsl(var(--destructive))",
                    foreground: "hsl(var(--destructive-foreground))",
                },
                success: {
                    DEFAULT: "hsl(var(--success))",
                    foreground: "hsl(var(--success-foreground))",
                },
                warning: {
                    DEFAULT: "hsl(var(--warning))",
                    foreground: "hsl(var(--warning-foreground))",
                },
                info: {
                    DEFAULT: "hsl(var(--info))",
                    foreground: "hsl(var(--info-foreground))",
                },
                
                // Surface 和 交互状态
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
                
                // Material 3 原生色彩系统
                'md-primary': {
                    10: 'hsl(var(--md-primary-10) / <alpha-value>)',
                    20: 'hsl(var(--md-primary-20) / <alpha-value>)',
                    30: 'hsl(var(--md-primary-30) / <alpha-value>)',
                    40: 'hsl(var(--md-primary-40) / <alpha-value>)',
                    50: 'hsl(var(--md-primary-50) / <alpha-value>)',
                    60: 'hsl(var(--md-primary-60) / <alpha-value>)',
                    70: 'hsl(var(--md-primary-70) / <alpha-value>)',
                    80: 'hsl(var(--md-primary-80) / <alpha-value>)',
                    90: 'hsl(var(--md-primary-90) / <alpha-value>)',
                    95: 'hsl(var(--md-primary-95) / <alpha-value>)',
                    99: 'hsl(var(--md-primary-99) / <alpha-value>)',
                },
                'md-secondary': {
                    10: 'hsl(var(--md-secondary-10) / <alpha-value>)',
                    20: 'hsl(var(--md-secondary-20) / <alpha-value>)',
                    30: 'hsl(var(--md-secondary-30) / <alpha-value>)',
                    40: 'hsl(var(--md-secondary-40) / <alpha-value>)',
                    50: 'hsl(var(--md-secondary-50) / <alpha-value>)',
                    60: 'hsl(var(--md-secondary-60) / <alpha-value>)',
                    70: 'hsl(var(--md-secondary-70) / <alpha-value>)',
                    80: 'hsl(var(--md-secondary-80) / <alpha-value>)',
                    90: 'hsl(var(--md-secondary-90) / <alpha-value>)',
                },
                'md-tertiary': {
                    10: 'hsl(var(--md-tertiary-10) / <alpha-value>)',
                    20: 'hsl(var(--md-tertiary-20) / <alpha-value>)',
                    30: 'hsl(var(--md-tertiary-30) / <alpha-value>)',
                    40: 'hsl(var(--md-tertiary-40) / <alpha-value>)',
                    80: 'hsl(var(--md-tertiary-80) / <alpha-value>)',
                },
                'md-neutral': {
                    4: 'hsl(var(--md-neutral-4) / <alpha-value>)',
                    6: 'hsl(var(--md-neutral-6) / <alpha-value>)',
                    10: 'hsl(var(--md-neutral-10) / <alpha-value>)',
                    12: 'hsl(var(--md-neutral-12) / <alpha-value>)',
                    17: 'hsl(var(--md-neutral-17) / <alpha-value>)',
                    22: 'hsl(var(--md-neutral-22) / <alpha-value>)',
                    24: 'hsl(var(--md-neutral-24) / <alpha-value>)',
                    87: 'hsl(var(--md-neutral-87) / <alpha-value>)',
                    90: 'hsl(var(--md-neutral-90) / <alpha-value>)',
                    94: 'hsl(var(--md-neutral-94) / <alpha-value>)',
                    96: 'hsl(var(--md-neutral-96) / <alpha-value>)',
                    98: 'hsl(var(--md-neutral-98) / <alpha-value>)',
                },
                'md-neutral-variant': {
                    30: 'hsl(var(--md-neutral-variant-30) / <alpha-value>)',
                    50: 'hsl(var(--md-neutral-variant-50) / <alpha-value>)',
                    80: 'hsl(var(--md-neutral-variant-80) / <alpha-value>)',
                    90: 'hsl(var(--md-neutral-variant-90) / <alpha-value>)',
                },
                
                // 保留的项目特色色彩（向后兼容）
                'lavender': {
                    50: 'hsl(var(--md-primary-99) / <alpha-value>)',
                    100: 'hsl(var(--md-primary-95) / <alpha-value>)',
                    200: 'hsl(var(--md-primary-90) / <alpha-value>)',
                    300: 'hsl(var(--md-primary-80) / <alpha-value>)',
                    400: 'hsl(var(--md-primary-70) / <alpha-value>)',
                    500: 'hsl(var(--md-primary-60) / <alpha-value>)',
                    600: 'hsl(var(--md-primary-50) / <alpha-value>)',
                    700: 'hsl(var(--md-primary-40) / <alpha-value>)',
                    800: 'hsl(var(--md-primary-30) / <alpha-value>)',
                    900: 'hsl(var(--md-primary-20) / <alpha-value>)',
                    950: 'hsl(var(--md-primary-10) / <alpha-value>)',
                },
            },
            
            // Material 3 间距系统
            spacing: {
                'md-0': 'var(--md-space-0)',
                'md-1': 'var(--md-space-1)',
                'md-2': 'var(--md-space-2)',
                'md-3': 'var(--md-space-3)',
                'md-4': 'var(--md-space-4)',
                'md-5': 'var(--md-space-5)',
                'md-6': 'var(--md-space-6)',
                'md-8': 'var(--md-space-8)',
                'md-10': 'var(--md-space-10)',
                'md-12': 'var(--md-space-12)',
            },
            
            // Material 3 圆角系统
            borderRadius: {
                'md-none': 'var(--md-corner-none)',
                'md-xs': 'var(--md-corner-xs)',
                'md-sm': 'var(--md-corner-sm)',
                'md-md': 'var(--md-corner-md)',
                'md-lg': 'var(--md-corner-lg)',
                'md-xl': 'var(--md-corner-xl)',
                'md-full': 'var(--md-corner-full)',
                
                // 兼容现有系统
                lg: "var(--md-corner-lg)",
                md: "var(--md-corner-md)",
                sm: "var(--md-corner-sm)",
            },
            
            // Material 3 阴影系统
            boxShadow: {
                'md-0': 'var(--md-elevation-0)',
                'md-1': 'var(--md-elevation-1)',
                'md-2': 'var(--md-elevation-2)',
                'md-3': 'var(--md-elevation-3)',
                'md-4': 'var(--md-elevation-4)',
                'md-5': 'var(--md-elevation-5)',
            },
            
            // Material 3 字体系统
            fontFamily: {
                'md-display': ['Inter', 'Roboto', 'system-ui', 'sans-serif'],
                'md-body': ['Inter', 'Roboto', 'system-ui', 'sans-serif'],
                'md-label': ['Inter', 'Roboto', 'system-ui', 'sans-serif'],
            },
            
            fontSize: {
                'md-display-large': ['3.563rem', { lineHeight: '4rem', fontWeight: '400' }],
                'md-display-medium': ['2.813rem', { lineHeight: '3.25rem', fontWeight: '400' }],
                'md-display-small': ['2.25rem', { lineHeight: '2.75rem', fontWeight: '400' }],
                
                'md-headline-large': ['2rem', { lineHeight: '2.5rem', fontWeight: '400' }],
                'md-headline-medium': ['1.75rem', { lineHeight: '2.25rem', fontWeight: '400' }],
                'md-headline-small': ['1.5rem', { lineHeight: '2rem', fontWeight: '400' }],
                
                'md-title-large': ['1.375rem', { lineHeight: '1.75rem', fontWeight: '500' }],
                'md-title-medium': ['1rem', { lineHeight: '1.5rem', fontWeight: '500' }],
                'md-title-small': ['0.875rem', { lineHeight: '1.25rem', fontWeight: '500' }],
                
                'md-body-large': ['1rem', { lineHeight: '1.5rem', fontWeight: '400' }],
                'md-body-medium': ['0.875rem', { lineHeight: '1.25rem', fontWeight: '400' }],
                'md-body-small': ['0.75rem', { lineHeight: '1rem', fontWeight: '400' }],
                
                'md-label-large': ['0.875rem', { lineHeight: '1.25rem', fontWeight: '500' }],
                'md-label-medium': ['0.75rem', { lineHeight: '1rem', fontWeight: '500' }],
                'md-label-small': ['0.6875rem', { lineHeight: '1rem', fontWeight: '500' }],
            },
            
            // Material 3 动效系统
            transitionDuration: {
                'md-short1': 'var(--md-motion-duration-short1)',
                'md-short2': 'var(--md-motion-duration-short2)',
                'md-short3': 'var(--md-motion-duration-short3)',
                'md-short4': 'var(--md-motion-duration-short4)',
                'md-medium1': 'var(--md-motion-duration-medium1)',
                'md-medium2': 'var(--md-motion-duration-medium2)',
                'md-medium3': 'var(--md-motion-duration-medium3)',
                'md-medium4': 'var(--md-motion-duration-medium4)',
                'md-long1': 'var(--md-motion-duration-long1)',
                'md-long2': 'var(--md-motion-duration-long2)',
                'md-long3': 'var(--md-motion-duration-long3)',
                'md-long4': 'var(--md-motion-duration-long4)',
            },
            
            transitionTimingFunction: {
                'md-linear': 'var(--md-motion-easing-linear)',
                'md-standard': 'var(--md-motion-easing-standard)',
                'md-emphasized': 'var(--md-motion-easing-emphasized)',
                'md-decelerated': 'var(--md-motion-easing-decelerated)',
                'md-accelerated': 'var(--md-motion-easing-accelerated)',
            },
            
            // 动画关键帧
            keyframes: {
                'fade-in': {
                    'from': { opacity: '0' },
                    'to': { opacity: '1' }
                },
                'slide-up': {
                    'from': { 
                        opacity: '0', 
                        transform: 'translateY(16px)' 
                    },
                    'to': { 
                        opacity: '1', 
                        transform: 'translateY(0)' 
                    }
                },
                'pulse-gentle': {
                    '0%, 100%': { opacity: '1' },
                    '50%': { opacity: '0.7' }
                }
            },
            
            animation: {
                'fade-in': 'fade-in var(--md-motion-duration-medium2) var(--md-motion-easing-standard)',
                'slide-up': 'slide-up var(--md-motion-duration-medium3) var(--md-motion-easing-emphasized)',
                'pulse-gentle': 'pulse-gentle 2s var(--md-motion-easing-standard) infinite',
            }
        },
    },
    plugins: [],
}