import { Kanban, MessageSquare, Search, Users } from 'lucide-react';
import type { ReactNode } from 'react';

interface FeatureCardProps {
  icon: ReactNode;
  title: string;
  description: string;
}

const FeatureCard = ({ icon, title, description }: FeatureCardProps) => (
  <article className="border-border bg-card rounded-xl border p-5 sm:p-6">
    <div className="bg-primary/10 text-primary mb-5 flex size-9 items-center justify-center rounded-md">{icon}</div>
    <h3 className="text-foreground text-base font-semibold">{title}</h3>
    <p className="text-muted-foreground mt-2 text-sm leading-6">{description}</p>
  </article>
);

export const LandingFeatures = () => (
  <section className="border-border border-t px-4 py-20 sm:px-6 lg:py-28">
    <div className="mx-auto max-w-6xl">
      <div className="grid gap-12 lg:grid-cols-[0.8fr_1.2fr] lg:items-start">
        <div className="lg:sticky lg:top-28">
          <p className="text-primary mb-4 text-xs font-semibold tracking-[0.16em] uppercase">One shared context</p>
          <h2 className="text-foreground text-3xl leading-tight font-semibold tracking-[-0.03em] text-balance sm:text-4xl">
            The board tells you what. The conversation tells you why.
          </h2>
          <p className="text-muted-foreground mt-5 max-w-md leading-7">
            Tasks, decisions, and activity live together, so the latest context is always within reach.
          </p>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <FeatureCard
            icon={<Kanban className="size-4" />}
            title="A board that stays current"
            description="Priorities, owners, tags, and due dates are visible at a glance—and update for everyone in real time."
          />
          <FeatureCard
            icon={<MessageSquare className="size-4" />}
            title="Conversation beside the work"
            description="Project chat and task comments keep decisions attached to the people and work they affect."
          />
          <FeatureCard
            icon={<Search className="size-4" />}
            title="Find the thread again"
            description="Search across projects and tasks instead of reconstructing context from scattered tools."
          />
          <FeatureCard
            icon={<Users className="size-4" />}
            title="Built for small teams"
            description="A focused workspace with the structure you need and none of the process theater you don't."
          />
        </div>
      </div>
    </div>
  </section>
);
