import { Link } from '@tanstack/react-router';
import { ArrowRight, Sparkles } from 'lucide-react';

export const LandingCta = () => (
  <section className="px-4 pb-20 sm:px-6 lg:pb-28">
    <div className="border-border bg-card relative mx-auto max-w-6xl overflow-hidden rounded-2xl border px-6 py-12 text-center shadow-2xl shadow-black/10 sm:px-12 sm:py-16">
      <div
        className="pointer-events-none absolute inset-0 opacity-40"
        aria-hidden="true"
        style={{
          background:
            'radial-gradient(circle at 50% 130%, color-mix(in srgb, var(--brand) 28%, transparent), transparent 52%)',
        }}
      />
      <div className="relative">
        <div className="bg-primary/10 text-primary mx-auto mb-5 flex size-10 items-center justify-center rounded-lg">
          <Sparkles className="size-5" />
        </div>
        <h2 className="text-foreground text-3xl font-semibold tracking-[-0.03em] sm:text-4xl">
          Give your next project one clear home.
        </h2>
        <p className="text-muted-foreground mx-auto mt-4 max-w-xl leading-7">
          Create a workspace, invite your team, and watch the board come alive.
        </p>
        <Link
          to="/sign-up"
          className="bg-primary text-primary-foreground hover:bg-primary/90 mt-7 inline-flex h-11 items-center gap-2 rounded-md px-5 text-sm font-semibold transition-colors"
        >
          Create your workspace
          <ArrowRight className="size-4" />
        </Link>
      </div>
    </div>
  </section>
);
