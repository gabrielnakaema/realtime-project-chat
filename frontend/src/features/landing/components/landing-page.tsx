import { BoardPreview } from './board-preview';
import { LandingCta } from './landing-cta';
import { LandingFeatures } from './landing-features';
import { LandingFooter } from './landing-footer';
import { LandingHeader } from './landing-header';
import { LandingHero } from './landing-hero';

export const LandingPage = () => (
  <div className="bg-background min-h-screen overflow-hidden">
    <LandingHeader />

    <main>
      <LandingHero />
      <BoardPreview />
      <LandingFeatures />
      <LandingCta />
    </main>

    <LandingFooter />
  </div>
);
