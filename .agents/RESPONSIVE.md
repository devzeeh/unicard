---
name: responsive-design
description: Rules and breakpoint system for making web and mobile UI responsive across screen sizes — for Unicard and other projects, in any stack (plain HTML/CSS, Tailwind, React, Vue, Flutter). Use whenever the user asks to make a screen responsive, fix mobile/tablet layout, adapt a design across breakpoints, handle safe areas or notches, or reflow a dashboard/table/form for smaller screens. Pairs with the monochrome-ui-design skill for visual tokens — use this one specifically for layout adaptation, breakpoints, touch targets, and web↔mobile parity.
license: Complete terms in LICENSE.txt
---

# Responsive Design

Rules for making layouts adapt cleanly from small phone screens up through large desktop monitors, and for keeping web (Vue SPA) and mobile (Flutter) versions of the same product feeling like one product at different sizes.

## Core principle

Design mobile-first, then add complexity as space grows — not the reverse. Starting from desktop and cramming it down produces cluttered phone screens; starting from the smallest useful layout and expanding gives every breakpoint a layout that was actually designed for it.

A responsive screen isn't just "things get smaller" — content, navigation pattern, and information density should each be reconsidered at every breakpoint band, not just scaled.

## Breakpoint scale

Use a consistent set of breakpoints across the whole product so components don't disagree about where "mobile" ends:

| Name | Min width | Typical device |
|---|---|---|
| `xs` | 0px | Small phones |
| `sm` | 480px | Large phones |
| `md` | 768px | Tablets (portrait) |
| `lg` | 1024px | Tablets (landscape), small laptops |
| `xl` | 1280px | Desktop |
| `2xl` | 1536px | Large desktop / wide monitors |

Tailwind's default scale (`sm:640 md:768 lg:1024 xl:1280 2xl:1536`) is close enough to this and fine to use as-is — don't invent a custom scale unless the product has unusual device targets. For plain CSS, define these as `min-width` media queries; always design content-out (query when *this component's* content breaks, not just at the named breakpoints) in addition to the standard set.

## Layout adaptation by breakpoint band

**xs–sm (phone):**
- Single column. Stack everything vertically.
- Navigation: bottom tab bar (mobile app) or hamburger/drawer (mobile web) — never a full horizontal nav bar.
- Tables: don't shrink columns to fit — convert to stacked cards (each row becomes a card with label:value pairs) or allow horizontal scroll with a sticky first column (e.g., transaction date/amount) if row-to-card conversion loses too much scanability.
- Modals: full-screen or bottom sheet, not a small centered box.
- Forms: one field per row, full-width inputs, submit button full-width and sticky to the bottom of the viewport for long forms.

**md (tablet):**
- 2-column layouts become viable for cards/grids; forms can go 2-column for short paired fields (city/zip) but stay 1-column for anything sequential.
- Navigation can move to a persistent left rail or stay as a drawer, depending on how much nav depth the product has — Unicard's admin panel likely wants a rail here, not a drawer.
- Tables can show more columns but still consider hiding low-priority columns (see Priority below) rather than shrinking font size to fit.

**lg–xl (desktop):**
- Multi-column layouts, persistent sidebar navigation, tables shown in full.
- Introduce secondary panels (detail view alongside list view) where a phone would need two full screens — this is the main structural difference desktop affords, not just "more space," but the ability to show list + detail simultaneously.
- Cap content width for readability (640–720px for text, up to 1280–1440px for dashboards) — full-bleed content on an ultrawide monitor hurts scanability as much as a cramped phone screen does.

**2xl (wide desktop):**
- Don't just stretch the lg layout wider — add a third column/panel, increase gutters, or cap the max-width and center the layout with generous margins. Stretched form fields and stretched tables at 1800px+ are a common miss.

## Reflow strategy for common patterns

**Navigation:** horizontal bar (desktop) → hamburger/drawer or bottom tabs (mobile). Pick bottom tabs for an app with ≤5 primary destinations (fits Flutter conventions well); pick a drawer for deeper/nested nav (admin panels, settings-heavy apps).

**Data tables (transactions, users, cards):**
1. Priority-rank columns: which 2–3 columns does the user need at a glance (date, amount, status)? Those survive to the smallest screen.
2. At `md` and below, either (a) collapse to stacked cards with label:value rows, or (b) keep a table with only priority columns and put the rest behind a row-expand/detail tap.
3. Never shrink font below the 13px accessibility floor to force columns to fit — reflow instead.

**Forms:** single column below `md`; group related fields into 2-column rows only at `md`+ and only for genuinely paired short fields. Keep field order identical across breakpoints — reordering fields between mobile and desktop confuses returning users.

**Cards/grids:** define column count per breakpoint explicitly rather than relying purely on `auto-fit`/`minmax` guessing — e.g., 1 column `xs`, 2 columns `md`, 3 columns `lg`, 4 columns `xl`, so card proportions stay intentional rather than accidentally-whatever-fits.

**Modals → sheets:** centered modal at `md`+, full-screen or bottom sheet below `md`. This is one of the highest-value adaptations for perceived native quality on mobile.

## Touch vs. pointer

- Touch targets ≥ 44×44px (iOS HIG) / 48×48dp (Material) regardless of the visual size of the icon or label inside — pad the tappable area.
- Hover-dependent affordances (tooltips-on-hover, hover-to-reveal actions) need a touch fallback: tap-to-reveal, always-visible on touch, or long-press — never leave an action reachable only via hover.
- Gate hover *styling* behind `@media (hover: hover) and (pointer: fine)` so touch devices don't get stuck ":hover" states after tapping.
- Increase spacing between adjacent tappable elements on touch layouts (8px minimum gap) to avoid mis-taps — pointer layouts can pack tighter.

## Typography & spacing scaling

- Don't scale type linearly with viewport by default — define explicit sizes per breakpoint band for headings (e.g., H1 28px mobile → 36px desktop) rather than `vw`-based fluid type, which can overshoot on very large or very small screens. Fluid type (`clamp()`) is fine for hero/marketing headlines specifically, with explicit min/max caps.
- Spacing scale stays the same token values across breakpoints (per monochrome-ui-design's 4/8px scale), but *which* tokens you use changes — e.g., section gap might be `32px` (token) on mobile and `64–96px` (token) on desktop, rather than inventing new arbitrary values per breakpoint.

## Safe areas & device quirks

- Respect safe-area insets on mobile web and native (notches, home indicator, punch-hole cameras): `env(safe-area-inset-top/bottom/left/right)` in CSS; `SafeArea` widget in Flutter. Apply to sticky headers/footers and full-bleed bottom sheets especially.
- Sticky bottom action bars (checkout, submit) need bottom safe-area padding or they sit under the home indicator.
- Account for the on-screen keyboard covering inputs on mobile — ensure the focused field scrolls into view above the keyboard (usually automatic on Flutter/native; verify manually on web forms in an iframe or complex scroll container).

## Web ↔ Flutter parity (Unicard-specific)

Since Unicard spans a Vue web app and a Flutter mobile client:
- Keep the token layer (colors, spacing, radius, type scale from monochrome-ui-design) identical in both, sourced from one written spec so they don't drift.
- Navigation pattern *can* legitimately differ (bottom tabs in Flutter, top nav/drawer in Vue) — that's platform convention, not inconsistency.
- Component *behavior* should match even when the visual chrome differs slightly per platform: same validation rules, same error copy, same states (loading/empty/error) for the same feature.
- Test the same screen's information hierarchy on both — if the Flutter app shows amount → status → date, the Vue table should prioritize the same order when it reflows to a card at small width, so the product feels like one system across web and mobile.

## Process

1. Design the `xs` layout first: what's the single most important content and action here, with everything else stacked below it.
2. Identify the 1–2 breakpoints where the layout structurally changes (not just shrinks) — usually `md` (nav pattern shift) and `xl` (list+detail split becomes viable). Not every breakpoint needs a unique layout.
3. For any table/list, priority-rank columns before deciding the reflow strategy.
4. Verify touch targets, hover fallbacks, and safe-area handling on the smallest breakpoint.
5. Check typography against the explicit per-breakpoint sizes rather than scaling everything fluidly.
6. If building for both Unicard's Vue web and Flutter mobile, confirm token and information-hierarchy parity per the section above.