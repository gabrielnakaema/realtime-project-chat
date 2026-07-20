import { Link } from '@tanstack/react-router';
import { ArrowRight, ChevronRight, Radio } from 'lucide-react';

export const LandingHero = () => (
  <section className="relative px-4 pt-20 pb-12 sm:px-6 sm:pt-28 lg:pt-32">
    <div
      className="pointer-events-none absolute inset-x-0 top-0 mx-auto h-[560px] max-w-5xl opacity-70"
      aria-hidden="true"
      style={{
        background:
          'radial-gradient(circle at 50% 10%, color-mix(in srgb, var(--brand) 20%, transparent), transparent 58%)',
      }}
    />
    <div className="relative mx-auto max-w-4xl text-center">
      <div className="border-border bg-card text-muted-foreground mx-auto mb-6 flex w-fit items-center gap-2 rounded-full border px-3 py-1.5 text-xs font-medium shadow-sm">
        <Radio className="text-success size-3.5" />
        Every project update, live for the whole team
      </div>
      <h1 className="text-foreground text-4xl leading-[1.05] font-semibold tracking-[-0.045em] text-balance sm:text-6xl lg:text-7xl">
        Move work forward,
        <span className="text-muted-foreground block">without losing the conversation.</span>
      </h1>
      <p className="text-muted-foreground mx-auto mt-6 max-w-2xl text-base leading-7 text-pretty sm:text-lg">
        Plan on a clear kanban board, discuss the details in project chat, and keep every teammate in sync as work
        changes.
      </p>
      <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
        <Link
          to="/sign-up"
          className="bg-primary text-primary-foreground hover:bg-primary/90 inline-flex h-11 w-full items-center justify-center gap-2 rounded-md px-5 text-sm font-semibold transition-colors sm:w-auto"
        >
          Get started for free
          <ArrowRight className="size-4" />
        </Link>
        <a
          href="#product-preview"
          className="border-border bg-card text-foreground hover:bg-muted inline-flex h-11 w-full items-center justify-center gap-2 rounded-md border px-5 text-sm font-semibold transition-colors sm:w-auto"
        >
          Explore the board
          <ChevronRight className="size-4" />
        </a>
      </div>
    </div>
  </section>
);
