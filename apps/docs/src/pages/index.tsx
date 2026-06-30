import type {ReactNode} from 'react';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';
import styles from './index.module.css';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={styles.heroBanner}>
      <div className={styles.heroInner}>
        <Heading as="h1" className={styles.heroTitle}>
          <span className={styles.heroTitleGradient}>{siteConfig.title}</span>
        </Heading>
        <p className={styles.heroSubtitle}>A powerful, feature-rich web crawler for modern applications</p>
        <div className={styles.buttons}>
          <Link className={styles.buttonPrimary} to="/docs/introduction">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="9 18 15 12 9 6" /></svg>
            Get Started
          </Link>
          <a className={styles.buttonSecondary} href="https://github.com/mindfiredigital/DeepScanBot" target="_blank" rel="noopener noreferrer">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/></svg>
            GitHub
          </a>
        </div>
        <div className={styles.heroTerminal}>
          <div className={styles.terminalHeader}>
            <span className={styles.terminalDot} />
            <span className={styles.terminalDot} />
            <span className={styles.terminalDot} />
          </div>
          <div className={styles.terminalBody}>
            <span className={styles.terminalPrompt}>$</span>{' '}
            <span className={styles.terminalCommand}>deepscanbot -url https://example.com -depth 2 -json -output scan</span>
            <br />
            <span className={styles.terminalOutput}>[INFO] Starting crawl: url=https://example.com max-depth=2 concurrency=4 retries=0 delay=0s</span>
            <br />
            <span className={styles.terminalOutput}>[INFO] Crawling https://example.com (depth=0)</span>
            <br />
            <span className={styles.terminalOutput}>[INFO] Crawled https://example.com [status=200] [result=passed]</span>
            <br />
            <span className={styles.terminalOutput}>[INFO] Crawl finished: total=12 passed=12 failed=0 skipped=0 duration=1.2s</span>
          </div>
        </div>
      </div>
    </header>
  );
}

const featuresData = [
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="16 18 22 12 16 6" />
        <polyline points="8 6 2 12 8 18" />
      </svg>
    ),
    title: 'Multi-Threaded Crawling',
    description: 'Concurrent architecture with configurable worker pools, per-host rate limiting, and CPU-aware auto-scaling for optimal performance.',
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
    ),
    title: 'Robust & Configurable',
    description: 'Robots.txt compliance, retry logic with exponential backoff, proxy support, TLS options, and depth control out of the box.',
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <ellipse cx="12" cy="5" rx="9" ry="3" />
        <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3" />
        <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5" />
      </svg>
    ),
    title: 'Rich Output Formats',
    description: 'JSON or text reports with detailed summaries, status code distribution, skip reason breakdowns, and retry distribution analytics.',
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
        <line x1="3" y1="9" x2="21" y2="9" />
        <line x1="9" y1="21" x2="9" y2="9" />
      </svg>
    ),
    title: 'Sitemap & Resume Mode',
    description: 'Auto-discover sitemaps, resume interrupted crawls without recrawling visited URLs, and handle 1000+ page sites gracefully.',
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
        <line x1="12" y1="9" x2="12" y2="13" />
        <line x1="12" y1="17" x2="12.01" y2="17" />
      </svg>
    ),
    title: 'Rate-Limit Handling',
    description: 'Smart Retry-After header parsing, automatic backoff for 429 responses, and configurable politeness delays between requests.',
  },
  {
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
        <polyline points="14 2 14 8 20 8" />
        <line x1="16" y1="13" x2="8" y2="13" />
        <line x1="16" y1="17" x2="8" y2="17" />
      </svg>
    ),
    title: 'Content-Type Filtering',
    description: 'Filter downloads by MIME type, enforce page size limits, and focus on specific content types like HTML or PDFs.',
  },
];

function FeatureCard({icon, title, description}: {icon: ReactNode; title: string; description: string}) {
  return (
    <div className="col col--4" style={{marginBottom: '1.5rem'}}>
      <div className={styles.featureCard}>
        <div className={styles.featureIcon}>
          {icon}
        </div>
        <h3 className={styles.featureTitle}>{title}</h3>
        <p className={styles.featureDesc}>{description}</p>
      </div>
    </div>
  );
}

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title="Home | Deep Scan Bot"
      description="DeepScanBot - A powerful, feature-rich web crawler built with Go">
      <HomepageHeader />
      <main>
        <section className={styles.features}>
          <div className="container">
            <div className={styles.sectionHeader}>
              <Heading as="h2" className={styles.sectionTitle}>
                Everything you need for web crawling
              </Heading>
              <p className={styles.sectionSubtitle}>
                A complete toolkit for crawling, scraping, and analyzing websites at scale.
              </p>
            </div>
            <div className="row">
              {featuresData.map((feature, idx) => (
                <FeatureCard key={idx} {...feature} />
              ))}
            </div>
          </div>
        </section>
      </main>
      <section className={styles.cta}>
        <div className={styles.ctaInner}>
          <Heading as="h2" className={styles.ctaTitle}>
            Ready to start crawling?
          </Heading>
          <p className={styles.ctaSubtitle}>
            Get started in minutes. No configuration needed — just point and crawl.
          </p>
          <Link className={styles.ctaButton} to="/docs/introduction">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="5" y1="12" x2="19" y2="12" /><polyline points="12 5 19 12 12 19" /></svg>
            Get Started
          </Link>
        </div>
      </section>
    </Layout>
  );
}