# DESIGN.md - LLM Proxy Gateway

This document defines the design system for the LLM Proxy Gateway. AI agents should refer to this file when generating or modifying any UI components.

## 1. Design Principles
- **Modern Minimalist**: Focus on whitespace, clarity, and essential data.
- **Glassmorphism**: Subtle translucency and blurs for depth.
- **Micro-interactions**: Every action should have a visual response.
- **Responsive**: Mobile-first, desktop-optimized.

## 2. Visual Tokens

### Colors (Dark Mode Optimized)
- `bg-primary`: `#0f1117` (Deep Midnight)
- `bg-secondary`: `#1a1d26` (Sidebar/Cards)
- `accent-purple`: `#8b5cf6` (Primary Action)
- `accent-cyan`: `#06b6d4` (Success/Secondary Action)
- `text-primary`: `#f8fafc` (High Contrast)
- `text-secondary`: `#94a3b8` (Muted/Subtle)
- `border-color`: `rgba(255, 255, 255, 0.08)`
- `status-success`: `#10b981`
- `status-error`: `#ef4444`

### Typography
- **Primary Font**: 'Inter', sans-serif (Google Fonts)
- **Monospace**: 'JetBrains Mono', monospace (for API keys/logs)
- **Scale**:
  - `h1`: 32px / 700 weight
  - `h2`: 24px / 600 weight
  - `body`: 14px / 400 weight

### Geometry
- **Radius**:
  - `card`: 20px
  - `button`: 12px
  - `input`: 12px
- **Padding**:
  - `section`: 40px
  - `card`: 32px

## 3. Component Specifications

### Buttons
- **Primary**: Gradient `accent-purple` to `accent-cyan`, hover scale 1.02.
- **Secondary**: Ghost style with border, subtle glow on hover.

### Modals (Popups)
- **Overlay**: `rgba(0, 0, 0, 0.8)` with `backdrop-filter: blur(12px)`.
- **Card**: Max-width 650px, centered or top-aligned, vertical scrolling on overlay.

### Tables
- **Row**: Subtle border-bottom, highlight on hover.
- **Badges**: Pill-shaped, semi-transparent background of status color.

## 4. Layout Rules
- **Sidebar**: Fixed 260px, integrated search bar at top.
- **Main Content**: Max-width 1200px, centered horizontally.
- **Gaps**: Standard 24px between cards.
