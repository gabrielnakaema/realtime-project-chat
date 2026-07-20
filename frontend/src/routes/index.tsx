import '@/styles/animations.css';
import { Link, createFileRoute } from '@tanstack/react-router';
import { ClipboardList, Kanban, MessageSquareMore, Users, Zap } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { cn } from '../lib/utils';

export const Route = createFileRoute('/')({
  component: App,
});

function App() {
  const [isVisible, setIsVisible] = useState(false);
  const activityRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true);
        }
      },
      { threshold: 0.3 },
    );

    if (activityRef.current) {
      observer.observe(activityRef.current);
    }

    return () => observer.disconnect();
  }, []);

  return (
    <div className="bg-muted relative min-h-screen">
      <header className="border-border flex items-center justify-between border-b px-6 py-6 lg:px-8">
        <div className="flex items-center space-x-3">
          <div className="bg-primary flex h-7 w-7 items-center justify-center rounded">
            <span className="text-primary-foreground text-sm font-bold">T</span>
          </div>
          <span className="text-foreground text-xl font-semibold">TaskFlow</span>
        </div>
        <div className="flex items-center space-x-6">
          <Link
            to="/login"
            className="text-muted-foreground hover:text-foreground text-sm font-medium transition-colors"
          >
            Log in
          </Link>
          <Link
            to="/sign-up"
            className="bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg px-4 py-2 text-sm font-medium transition-colors"
          >
            Get Started
          </Link>
        </div>
      </header>

      <main>
        <section className="container mx-auto px-6 py-24 lg:px-8 lg:py-32">
          <div className="mx-auto max-w-6xl">
            <div className="mb-16 text-center">
              <h1 className="text-foreground mb-8 text-5xl leading-tight font-bold text-balance lg:text-7xl">
                Real-time collaboration for small teams
              </h1>
              <p className="text-muted-foreground mx-auto mb-12 max-w-3xl text-xl leading-relaxed text-pretty lg:text-2xl">
                Manage projects and tasks with instant chat, real-time updates, and seamless team coordination.
                Everything happens instantly across your entire team.
              </p>
              <div className="flex flex-col justify-center gap-4 sm:flex-row">
                <Link
                  to="/sign-up"
                  className="bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg px-8 py-4 text-lg font-medium transition-colors"
                >
                  Start collaborating
                </Link>
                <Link
                  to="/login"
                  className="border-border text-foreground hover:bg-muted rounded-lg border px-8 py-4 text-lg font-medium transition-colors"
                >
                  View Features
                </Link>
              </div>
            </div>
          </div>
        </section>

        <section className="container mx-auto px-6 py-24 lg:px-8">
          <div className="mx-auto max-w-6xl">
            <div className="grid items-center gap-16 lg:grid-cols-2">
              <div>
                <div className="mb-8 flex items-center gap-3">
                  <Users />
                  <span className="text-muted-foreground text-sm font-medium tracking-wider">COLLABORATION</span>
                </div>
                <h2 className="text-foreground mb-8 text-5xl leading-tight font-bold text-balance lg:text-6xl">
                  Instant. Synchronized. Collaborative.
                </h2>
                <p className="text-muted-foreground text-xl leading-relaxed text-pretty">
                  Experience instant updates, seamless chat integration, and automatic synchronization that keeps your
                  team in perfect sync.
                </p>
              </div>

              <div className="border-border bg-card rounded-xl border p-8">
                <h3 className="text-foreground mb-2 text-2xl font-semibold">Stay synchronized in real-time.</h3>
                <p className="text-muted-foreground mb-8 text-lg">
                  Instant messaging, live task updates, and automatic notifications keep everyone on the same page.
                </p>
                <div ref={activityRef} className="space-y-4">
                  <div
                    className={cn(
                      'animate-slide-in animate-slide-in-1 bg-muted rounded-lg p-6',
                      isVisible && 'visible',
                    )}
                  >
                    <div className="mb-3 flex items-center gap-3">
                      <div className="animate-pulse-dot bg-primary h-2 w-2 rounded-full"></div>
                      <span className="text-foreground font-medium">Sarah moved 'Fix login bug' to In Progress</span>
                    </div>
                    <div className="text-muted-foreground text-sm">2 seconds ago</div>
                  </div>
                  <div
                    className={cn(
                      'animate-slide-in animate-slide-in-2 bg-muted rounded-lg p-6',
                      isVisible && 'visible',
                    )}
                  >
                    <div className="mb-3 flex items-center gap-3">
                      <div className="animate-pulse-dot-green bg-success h-2 w-2 rounded-full"></div>
                      <span className="text-foreground font-medium">New message in #design-team</span>
                    </div>
                    <div className="text-muted-foreground text-sm">12 seconds ago</div>
                  </div>
                  <div
                    className={cn(
                      'animate-slide-in animate-slide-in-3 bg-muted rounded-lg p-6',
                      isVisible && 'visible',
                    )}
                  >
                    <div className="text-muted-foreground mb-2 text-sm">Mike completed 'Database migration'</div>
                    <div className="text-muted-foreground text-sm">1 minute ago</div>
                  </div>
                  <div
                    className={cn(
                      'animate-slide-in animate-slide-in-4 bg-muted rounded-lg p-6',
                      isVisible && 'visible',
                    )}
                  >
                    <div className="text-muted-foreground mb-2 text-sm">3 new tasks added to Mobile App project</div>
                    <div className="text-muted-foreground text-sm">3 minutes ago</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="bg-card container mx-auto px-6 py-24 lg:px-8">
          <div className="mx-auto max-w-6xl">
            <div className="mb-16 text-center">
              <div className="mb-8 flex items-center justify-center gap-3">
                <ClipboardList />
                <span className="text-muted-foreground text-sm font-medium tracking-wider">FEATURES</span>
              </div>
              <h2 className="text-foreground mb-6 text-4xl leading-tight font-bold text-balance lg:text-5xl">
                Built for real-time team collaboration
              </h2>
              <p className="text-muted-foreground mx-auto max-w-3xl text-xl leading-relaxed text-pretty">
                Kanban boards, instant messaging, and live updates that work seamlessly together for effortless small
                team coordination.
              </p>
            </div>

            <div className="grid gap-8 md:grid-cols-2 lg:grid-cols-3">
              <div className="border-border bg-muted rounded-xl border p-8">
                <div className="bg-primary/10 mb-6 flex h-12 w-12 items-center justify-center rounded-lg">
                  <Kanban />
                </div>
                <h3 className="text-foreground mb-3 text-xl font-semibold">Live Kanban Boards</h3>
                <p className="text-muted-foreground leading-relaxed">
                  Real-time kanban boards with instant updates. See task movements and changes as they happen across
                  your team.
                </p>
              </div>

              <div className="border-border bg-muted rounded-xl border p-8">
                <div className="bg-success/10 mb-6 flex h-12 w-12 items-center justify-center rounded-lg">
                  <MessageSquareMore />
                </div>
                <h3 className="text-foreground mb-3 text-xl font-semibold">Instant Team Chat</h3>
                <p className="text-muted-foreground leading-relaxed">
                  Lightning-fast messaging that's always in sync. Chat seamlessly integrates with your projects and task
                  discussions.
                </p>
              </div>

              <div className="border-border bg-muted rounded-xl border p-8">
                <div className="bg-muted mb-6 flex h-12 w-12 items-center justify-center rounded-lg">
                  <Zap />
                </div>
                <h3 className="text-foreground mb-3 text-xl font-semibold">Always in Sync</h3>
                <p className="text-muted-foreground leading-relaxed">
                  Every action instantly appears for all team members. No refresh needed, no delays, just seamless
                  collaboration.
                </p>
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>
  );
}
