---
name: minimal-ui-design
description: Design system for building or redesigning minimal, clean UI for web and mobile apps, in any stack (plain HTML/CSS, Tailwind, React, Vue, Flutter). Use whenever the user asks to design, redesign, restyle, or theme a screen, component, dashboard, admin panel, or landing page in a minimal/clean/uncluttered style, or mentions light/dark mode, typography, spacing, buttons, cards, modals, shadows, or hover states. Covers color palette discipline, type scale, layout/spacing, texture, motion, and component specs, all in service of restraint and clarity rather than a specific color rule.
license: Complete terms in LICENSE.txt
---
 
# Minimal UI Design
 
A design system for minimal interfaces: a small, disciplined color palette, generous negative space, calm typography, and restrained motion. Minimalism here means every element earns its place, not that the interface is colorless — a minimal UI can be warm, cool, neutral, or brand-colored, as long as the palette is small and used consistently.
 
## Design philosophy
 
Minimal doesn't mean fewer decisions — it means every decision is deliberate and nothing is decorative for its own sake. Three things make a UI read as minimal:
 
1. **A small, closed palette.** A *counted* color budget — typically one neutral family plus one or two accent colors, each used consistently for the same meaning everywhere. The failure mode isn't "too much color," it's colors used inconsistently: a blue button here, a purple one there, with no system behind it.
2. **Negative space as a structural tool.** Whitespace isn't leftover space around content — it creates grouping, hierarchy, and breathing room. If a screen feels cluttered, the fix is usually removing an element or increasing spacing, not shrinking things to fit.
3. **One focal point per screen.** A minimal screen still needs hierarchy — a single clear primary action or piece of information the eye lands on first. Minimal without hierarchy just looks empty and directionless.
Ground the palette in the product's context before defaulting to a safe neutral. Ask what the product should feel like — warm and approachable, or precise and technical — and let that decide the neutral's undertone and the accent hue, rather than reaching for gray-plus-blue out of habit.
 
## Color palette
 
Structure: **one neutral ramp, one primary accent, an optional second accent, plus fixed semantic colors.** That's the whole budget — resist adding a color because one screen "needs" it; extend an existing token's use instead.
 
**Neutral ramp:** pick a base undertone (true gray, warm off-white/brown-tinted, or cool slate) and build an 8–10 step ramp from it. A warm neutral suits approachable, consumer-facing products; a cool or true neutral suits technical or data-heavy products.
 
Example neutral ramp (light theme):
| Token | Value | Use |
|---|---|---|
| `bg-canvas` | `#FAFAF8` | Page background |
| `bg-surface` | `#FFFFFF` | Cards, modals |
| `bg-subtle` | `#F1F0EC` | Inset panels, stripes |
| `border-default` | `#E5E3DC` | Dividers, input borders |
| `text-primary` | `#232220` | Headings, primary content |
| `text-secondary` | `#68655D` | Body copy |
| `text-tertiary` | `#A19D92` | Placeholders, metadata |
 
Dark theme: build a separate, intentional ramp rather than inverting the light one — surfaces should sit slightly lifted off true black, and borders need relatively more contrast than in light mode since dark surfaces have less natural edge definition. Carry the same undertone (warm/cool) through so dark mode still feels like the same product.
 
**Accent color(s):** one primary accent carries CTAs, links, and active states. A second accent is optional and should serve a genuinely distinct purpose (e.g., a data-visualization color separate from the CTA color) — never add a second accent purely for variety.
 
**Fixed semantic colors** (same role regardless of brand palette, not part of the creative budget):
- Success/positive → green family
- Warning → amber family
- Destructive/error → red family
Keep these consistent even if they clash slightly with the brand hue — users pattern-match red=stop, green=go across products, and reskinning them to match brand color undermines that.
 
**Discipline check:** if you can't state what each color in the palette means (this hue = this action or state), the palette has grown past minimal — audit and cut back before building.
 
## Typography
 
A neutral UI/body face (Inter, IBM Plex Sans, Manrope, or the system stack `-apple-system, "Segoe UI", Roboto, sans-serif`) covers most minimal interfaces. Enable tabular numerals (`font-variant-numeric: tabular-nums`) for any screen with financial or data figures so numbers align in columns without jitter. A more characterful display face can pair with the body face for marketing/landing headlines if the brand calls for it; product UI should keep display and body in the same neutral family.
 
**Type scale** (1.25 ratio for compact product UI, 1.333 for marketing pages that want more drama):
 
| Role | Size | Weight | Line-height |
|---|---|---|---|
| Display | 32–40px | 600–700 | 1.1–1.2 |
| H1 | 28px | 600 | 1.2 |
| H2 | 22px | 600 | 1.3 |
| H3 | 18px | 600 | 1.4 |
| Body | 15–16px | 400 | 1.5–1.6 |
| Small / caption | 13px | 400–500 | 1.4 |
| Micro (labels, badges) | 11–12px | 500–600, uppercase optional | 1.2 |
 
Rules: never go below 13px for readable text; use weight (500/600) rather than size jumps for in-paragraph emphasis; reserve underline for inline prose links, not UI chrome.
 
## Layout & spacing
 
Use a single spacing scale everywhere — no arbitrary values:
 
`4 · 8 · 12 · 16 · 24 · 32 · 48 · 64 · 96` (px, or dp on mobile)
 
- Card/section padding: 16–24px mobile, 24–32px desktop.
- Gap between related elements (label→input, icon→text): 8px.
- Gap between distinct components: 16–24px.
- Gap between page sections: 48–96px.
- Content max-width: 640–720px for text and forms, 1280–1440px for dashboards, centered with consistent gutters (16–24px mobile, 48–64px desktop).
Grid: 12-column desktop, 4-column mobile, 8-column tablet, with consistent gutters throughout.
 
## Texture
 
- **Borders over shadows for structure.** A 1px `border-default` line usually separates cards enough on a light background; reserve shadow for things that actually float (modals, dropdowns, toasts).
- **Tinted surfaces for subtle emphasis.** A light wash of the accent (4–6% opacity over `bg-surface`) marks a selected or active card without needing a hard accent border — a more restrained move than a solid-colored highlight.
- **One radius scale, applied consistently.** `4px` reads sharp/technical, `8px` is a balanced default, `12–16px` reads soft and consumer-friendly. Buttons, inputs, and cards should share the same radius family.
- **Dividers over boxes** in dense lists — a hairline row divider keeps visual weight lower than wrapping every row in its own bordered card.
- Optional subtle noise/grain (2–3% opacity) on large flat hero or empty-state surfaces only — skip it on dense data screens where legibility matters more.
## Shadows
 
Shadows should feel like soft directional light, not gray blur:
 
```
--shadow-sm: 0 1px 2px rgba(30,28,24,0.06);
--shadow-md: 0 4px 12px rgba(30,28,24,0.08), 0 1px 3px rgba(30,28,24,0.06);
--shadow-lg: 0 12px 32px rgba(30,28,24,0.12), 0 2px 6px rgba(30,28,24,0.08);
```
 
Tint the shadow color toward the neutral ramp's undertone (warm shadows for a warm-neutral palette, cooler/near-black shadows for a cool-neutral palette) — a small detail that keeps shadows from feeling like a bolted-on default.
 
Dark theme: shadows barely read on dark backgrounds, so elevation instead comes from a lighter surface step, with shadow reserved for true overlays over a dimmed scrim:
```
--elevation-1: bg-surface + 1px border-default;
--elevation-2: one step lighter than bg-surface;
--shadow-modal: 0 16px 48px rgba(0,0,0,0.5); /* modals/popovers only */
```
 
Elevation order: canvas < surface (cards) < modal/popover (lighter step or shadow-lg) < toast/tooltip (highest).
 
## Motion
 
Motion should confirm, not decorate — keep it short and consistent:
 
| Interaction | Duration | Easing |
|---|---|---|
| Hover | 120–150ms | `ease-out` |
| Press/active | 80–100ms | `ease-out` |
| Modal/sheet enter | 200–250ms | `cubic-bezier(0.16, 1, 0.3, 1)` |
| Modal/sheet exit | 150ms | `ease-in` |
| Page/route transition | 200ms | `ease-in-out` |
| Toast/notification | 250ms in, 200ms out | `ease-out` / `ease-in` |
 
Gate hover styling behind `@media (hover: hover)` so touch devices don't get stuck hover states. Buttons shift background one token step on hover, with scale (0.97–0.98) reserved for press/active. Cards and list rows shift `bg-canvas`→`bg-subtle` (or a light accent wash for a selected state) plus a border step-up; avoid gratuitous `translateY` lift on dense lists. Always respect `prefers-reduced-motion` with an instant or opacity-only fallback.
 
## Components
 
- **Primary button:** accent background, white/near-white label, radius and padding shared with secondary so they swap cleanly.
- **Secondary button:** `bg-surface` + `border-default`, `text-primary` label.
- **Ghost button:** no background or border, `text-secondary` → `text-primary` on hover — for low-emphasis actions.
- **Destructive button:** fixed red, never the brand accent, regardless of how warm or friendly the rest of the palette is.
- Button states: default → hover (bg step) → active (scale 0.97 + further bg step) → focus (visible ring) → disabled (opacity 40–50%, no hover response).
- **Cards:** `bg-surface`, `border-default` 1px, radius per scale, `shadow-sm` or border-only on dense screens. Use a light accent-wash background for a "selected" state rather than a heavy border.
- **Modals/sheets:** scrim `rgba(0,0,0,0.4)` light / `rgba(0,0,0,0.6)` dark, `bg-surface` (or a lighter dark-mode step), radius 12–16px, `shadow-lg`. Prefer a bottom sheet on mobile over a centered modal.
- **Inputs:** `bg-surface` or `bg-subtle`, `border-default` at rest, accent-tinted focus ring, `text-tertiary` placeholder. Errors get a red border plus red helper text plus an icon — never color alone.
- **Tables/lists:** `text-tertiary` uppercase-optional headers, hairline row dividers, tabular numerals right-aligned for amounts, zebra striping via `bg-subtle` only past roughly 10 rows.
- **Badges/status pills:** neutral (`bg-subtle`), active (light accent wash + accent text), success/warning/error using the fixed semantic colors — always paired with a label or icon, never a bare colored dot for anything critical.
## Accessibility floor (non-optional)
 
- Text contrast ≥ 4.5:1 for body text, ≥ 3:1 for large text (24px+, or 18px bold), against the actual background token in use.
- Every interactive element gets a visible focus state — never remove `outline` without a replacement ring.
- Never rely on hue alone for status, error, or amount sign — pair with an icon, symbol, or label. This matters more in a colored palette than in strict grayscale, since it's easier to lean on color-only signaling by habit.
- Touch targets ≥ 44×44px on mobile regardless of visual icon size — pad the tap area.
## Applying tokens across stacks
 
Keep the token layer separate from the component layer so redesigns and theme switches stay centralized:
 
- **Plain CSS / Tailwind:** define tokens as CSS custom properties on `:root` and `[data-theme="dark"]`, then point Tailwind's `theme.extend.colors` at those variables so `bg-surface` and its `dark:` variant stay in sync from one source.
- **React / Vue:** keep tokens in a single theme file or composable consumed by a ThemeProvider — never hardcode hex values inside component files.
- **Flutter:** mirror the ramp as light/dark `ThemeData`/`ColorScheme`, and mirror the spacing scale as named constants (`Spacing.xs/sm/md/lg`).
## Process
 
1. Decide the product's character (warm/approachable vs. cool/technical) before picking hues.
2. Set the color budget explicitly: one neutral ramp, one primary accent, optional second accent, plus the three fixed semantic colors — write down what each means.
3. Build light and dark ramps as separate, intentional sets rather than a computed inversion.
4. Map spacing, type, radius, and shadow to the shared scales above; every value should trace back to a token, never an arbitrary literal.
5. Run the palette discipline check — can you state what every color means? If not, cut back before building.
6. Self-check against the Accessibility floor before calling it done.
 