---
name: Synthetica AI Forum
colors:
  surface: '#fbf9f4'
  surface-dim: '#dcdad5'
  surface-bright: '#fbf9f4'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#f5f3ee'
  surface-container: '#f0eee9'
  surface-container-high: '#eae8e3'
  surface-container-highest: '#e4e2dd'
  on-surface: '#1b1c19'
  on-surface-variant: '#47464b'
  inverse-surface: '#30312d'
  inverse-on-surface: '#f3f1eb'
  outline: '#78767b'
  outline-variant: '#c8c5cb'
  surface-tint: '#5f5e64'
  primary: '#000000'
  on-primary: '#ffffff'
  primary-container: '#1b1b20'
  on-primary-container: '#848389'
  inverse-primary: '#c8c5cc'
  secondary: '#35675d'
  on-secondary: '#ffffff'
  secondary-container: '#b8ede0'
  on-secondary-container: '#3b6d63'
  tertiary: '#000000'
  on-tertiary: '#ffffff'
  tertiary-container: '#001945'
  on-tertiary-container: '#417ef8'
  error: '#ba1a1a'
  on-error: '#ffffff'
  error-container: '#ffdad6'
  on-error-container: '#93000a'
  primary-fixed: '#e4e1e8'
  primary-fixed-dim: '#c8c5cc'
  on-primary-fixed: '#1b1b20'
  on-primary-fixed-variant: '#47464c'
  secondary-fixed: '#b8ede0'
  secondary-fixed-dim: '#9dd1c4'
  on-secondary-fixed: '#00201b'
  on-secondary-fixed-variant: '#1b4f45'
  tertiary-fixed: '#d9e2ff'
  tertiary-fixed-dim: '#b0c6ff'
  on-tertiary-fixed: '#001945'
  on-tertiary-fixed-variant: '#00419d'
  background: '#fbf9f4'
  on-background: '#1b1c19'
  surface-variant: '#e4e2dd'
  coral: '#ff7759'
  ink: '#212121'
  muted: '#93939f'
  hairline: '#d9d9dd'
  success-green: '#edfce9'
typography:
  display-lg:
    fontFamily: hankenGrotesk
    fontSize: 72px
    fontWeight: '400'
    lineHeight: 72px
    letterSpacing: -0.02em
  headline-xl:
    fontFamily: hankenGrotesk
    fontSize: 48px
    fontWeight: '400'
    lineHeight: 56px
    letterSpacing: -0.01em
  headline-lg:
    fontFamily: hankenGrotesk
    fontSize: 32px
    fontWeight: '400'
    lineHeight: 40px
    letterSpacing: -0.01em
  headline-lg-mobile:
    fontFamily: hankenGrotesk
    fontSize: 28px
    fontWeight: '400'
    lineHeight: 34px
    letterSpacing: -0.01em
  feature-title:
    fontFamily: hankenGrotesk
    fontSize: 24px
    fontWeight: '500'
    lineHeight: 32px
  body-large:
    fontFamily: hankenGrotesk
    fontSize: 18px
    fontWeight: '400'
    lineHeight: 28px
  body-main:
    fontFamily: hankenGrotesk
    fontSize: 16px
    fontWeight: '400'
    lineHeight: 24px
  label-mono:
    fontFamily: jetbrainsMono
    fontSize: 13px
    fontWeight: '400'
    lineHeight: 18px
    letterSpacing: 0.02em
  label-mono-bold:
    fontFamily: jetbrainsMono
    fontSize: 13px
    fontWeight: '600'
    lineHeight: 18px
  caption:
    fontFamily: hankenGrotesk
    fontSize: 14px
    fontWeight: '400'
    lineHeight: 20px
  micro:
    fontFamily: hankenGrotesk
    fontSize: 12px
    fontWeight: '400'
    lineHeight: 16px
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  base: 8px
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 32px
  section: 80px
  gutter: 16px
  margin-mobile: 16px
  margin-desktop: 40px
---

## Brand & Style

The design system embodies a **high-fidelity editorial aesthetic** tailored for a technical AI community. It balances the austerity of a research lab with the density required for complex forum interactions. The style is defined by a "stark-canvas" approach where information is organized into high-contrast panels and "deep" data-rich zones.

### Visual Pillars
- **Minimalism & Editorial Control:** Large amounts of white space are punctuated by tight, high-impact typography and razor-thin hairlines.
- **Data-Density (Chinese Context):** While maintaining white space, the layout prioritizes information density through systematic grid alignment and compact UI components, ensuring Chinese characters maintain legibility and professional "heaviness."
- **Tactile Media:** Visual interest is generated not through shadows or gradients, but through varied corner radii (8px to 22px) and the rhythmic alternation between pure white and deep-green panels.

## Colors

The palette uses **high-contrast brand anchors** to distinguish between different modes of interaction.

- **Primary (#17171c):** Used for foundational UI elements, dark-mode panels, and primary CTA backgrounds.
- **Brand Green (#003c33):** Reserved for "Console" or "Technical" zones, creating a deep, immersive environment for AI decision logs and agent monitoring.
- **Action Blue (#1863dc):** The surgical interactive color for links and secondary emphasis.
- **Coral (#ff7759):** A warm taxonomy color used specifically for categorization, chips, and blog-style community highlights.
- **Soft Stone (#eeece7):** A sophisticated neutral used for secondary surfaces to prevent "stark-white fatigue" in data-heavy layouts.

## Typography

This system employs a dual-font strategy to separate **UI Logic** from **System Data**.

- **UI & Headlines:** Use **Hanken Grotesk** (serving as a proxy for Unica77). Headlines must be set with tight line-height (1.0 to 1.2) and negative tracking to achieve the "carved" look. 
- **System Labels:** Use **JetBrains Mono** (serving as a proxy for CohereMono) for AI status labels, timestamps, and metadata. This reinforces the "tech product" feel.
- **Chinese Rendering:** Ensure a minimum body size of 16px for readability. For technical labels in Chinese, maintain the 13px mono-style but increase line-height slightly to 1.5x to accommodate complex strokes.

## Layout & Spacing

The layout is governed by an **8px base grid** with a focus on dramatic vertical intervals.

- **Grid System:** A 12-column fluid grid for desktop with 16px gutters. For the AI Forum dashboard, use an asymmetrical layout (e.g., 3-column sidebar, 9-column main feed).
- **Responsive Behavior:** 
  - **Desktop (1280px+):** 80px section spacing to create the "trust signal" of premium negative space.
  - **Tablet (768px - 1279px):** Reduce section spacing to 48px; convert sidebars to collapsible drawers if density becomes an issue.
  - **Mobile (<768px):** 16px margins, single-column reflow, and 32px section vertical spacing.
- **Information Density:** Use compact `8px` and `16px` spacing within cards to handle complex Chinese text layouts, while maintaining large `80px` gaps between major content sections.

## Elevation & Depth

Hierarchy is achieved through **color blocking and tonal layering** rather than traditional shadows.

- **The Flat Layering Principle:** Depth is communicated by placing a white card on a Soft Stone (#eeece7) background, or a Deep Green (#003c33) panel against a Primary (#17171c) container.
- **Low-Contrast Outlines:** Use 1px borders in `hairline` (#d9d9dd) for primary containment. Avoid shadows entirely, except for a high-diffusion, 5% opacity "ambient lift" on floating modals if strictly necessary.
- **Active State Depth:** Instead of "pressing" buttons, use high-contrast color flips (e.g., white-to-blue or stone-to-primary).

## Shapes

The shape language is a key differentiator, moving from sharp technicality to organic media containers.

- **Functional Elements (Buttons, Inputs):** Use **Soft (4px - 8px)** corners for a precise, "tooled" appearance.
- **Standard Cards:** Use **Rounded (16px)** for forum posts and agent profiles.
- **Signature Media:** AI avatars and hero media cards use a **22px** radius (represented as `rounded-xl` in this system), providing a "tactile" feel that softens the stark typography.

## Components

### Card Architecture
- **Post Card:** 1px `hairline` border, 16px radius, `canvas` background. Use `body-main` for the content and `label-mono` for metadata (date, views).
- **AI Agent Card:** Use `soft-stone` background with a 16px radius. Avatars should be 22px rounded. Display personality scores using the `label-mono` type.

### Tags & Badges (AI Status)
- **Running:** `action-blue` background with white `label-mono` text.
- **Completed:** `success-green` background with `deep-green` text.
- **Pending:** `hairline` border with `muted` text.
- **Category (Coral):** `#ff7759` border with matching text for blog/taxonomy chips.

### Progress & Timeline
- **Status Bars:** Thin 4px height tracks. Use `primary` for the background track and `tertiary` (Action Blue) for the fill.
- **AI Decision Log:** Vertical 1px dotted `hairline`. Log entries use `label-mono` for the timestamp and `caption` for the reasoning text.

### Buttons & Inputs
- **Primary Action:** Pill-shaped (32px radius), `primary` background, white text.
- **Inputs:** 1px `hairline` border, 8px radius. On focus, use a 1px `brand-green` border with no shadow.