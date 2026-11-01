# Sezzle-inspired Design System

Extracted on 2026-11-01 from the computed styles of https://sezzle.com and the Sezzle mobile app screenshot.
Target stack: **React + Tailwind CSS v4** (tokens in `tokens.css` via `@theme`).

> Inspired by Sezzle's look for the calculator challenge. Don't ship Sezzle's logo or brand marks.

## Look and feel
- **Dark-first.** A deep purple background (`#201733`) with slightly lighter purple surfaces. Nothing is pure black.
- **Lavender text.** Body text is `#F2E6FF`, not white. Pure white is reserved for headlines and big numbers.
- **Mint is the single action color.** The primary CTA is always a mint pill (`#9FF4D9`) with dark purple text.
- **Purple glows** (radial gradients `#BB6AD3 → #7A3FB8`) behind hero content add depth.
- **Pills and circles.** Buttons are fully rounded, cards use 16–24px radius, and avatars and brand tiles are circles.
- **Satoshi** geometric sans. Headlines use weight 500–600 with tight negative tracking.

## Colors
| Token | Value | Use |
|---|---|---|
| `bg` | `#201733` | App background |
| `bg-deep` | `#1F1530` | Bottom nav, deepest layer |
| `surface` | `#241B3B` | Cards, inputs, search bar |
| `surface-2` | `#2C2247` | Raised / pressed |
| `overlay` | `#F2E6FF` @ 10% | Translucent fills on dark |
| `fg` | `#F2E6FF` | Primary text |
| `fg-strong` | `#FFFFFF` | Headlines, results |
| `fg-muted` | `#F2E6FF` @ 70% | Secondary text |
| `fg-subtle` | `#F2E6FF` @ 60% | Placeholders, captions |
| `line` | `#F2E6FF` @ 12% | Borders, dividers |
| `brand-500` | `#8333D4` | Accent purple (avatar, active tab, links) |
| `brand-400` | `#8E3DF2` | Focus rings, highlights |
| `mint` | `#9FF4D9` | Primary CTA |
| `danger` | `#FF5667` | Errors, badges |
| `warning` | `#FF8C34` | Warnings |

## Typography (Satoshi)
| Token | Size / line-height | Weight | Tracking |
|---|---|---|---|
| `text-display` | 50.7 / 54.7px | 500 | -0.025em |
| `text-h1` | 42.2 / 50.7px | 600 | -0.02em |
| `text-h2` | 28 / 33.6px | 500 | – |
| `text-lead` | 21.3 / 27.7px | 600 | – |
| `text-body-lg` | 18 / 23.4px | 400 | – |
| `text-body` | 16px | 400 | – |
| `text-sm` | 14 / 18.2px | 400 | – |
| `text-caption` | 13 / 16.9px | 400 | – |
| `text-label` | 13 / 20px | 600 | 0.02em |

On mobile, scale `display` and `h1` down (for example `text-h2 md:text-display`).

## Spacing, radius and elevation
- **Spacing:** 4px grid. The most common values are 8, 12, 16, 20, 24, 32, 48 and 56. Mobile page padding is 16–20px.
- **Radius:** `xs` 5px (tags), `sm` 8px (nav links), `lg` 12px (inputs), `2xl` 16–24px (cards), `pill` 100px (buttons), `full` (avatars).
- **Shadows:** `shadow-card` 0 10px 28px / 35% black. `shadow-modal` 0 24px 60px / 50% black.

## Components
| Component | Spec |
|---|---|
| Primary button | `h-11 px-5 rounded-pill bg-mint text-mint-fg text-sm font-medium` |
| Outline button | `border-2 border-mint text-mint`, transparent fill |
| Secondary button | `bg-brand-100 text-fg-inverse` (the "Sign Up" style) |
| Ghost button | `border border-fg text-fg` (the "Log In" style) |
| Input / search | `h-12 rounded-lg bg-surface border-line text-fg`, placeholder `fg-subtle`, focus ring `brand-400` |
| Card | `rounded-2xl bg-surface border border-line shadow-card` |
| Badge | `rounded-full bg-danger text-white text-[11px] min-w-5 h-5` |
| Bottom nav | `bg-bg-deep`, icon plus `text-caption`; active item `text-brand-300`, inactive `text-fg-muted` |

## Calculator mapping (suggestion)
- **Display:** a `card` showing the result in `text-display text-fg-strong tabular-nums` and the expression in `text-body text-fg-muted`.
- **Digit keys:** `bg-surface` circles or rounded-2xl buttons.
- **Operator keys:** `bg-brand-500 text-white`.
- **`=` key:** `bg-mint text-mint-fg`.
- **Clear key:** `bg-overlay`.
- **Errors (division by zero):** `text-danger` with a `border-danger` card outline.
