import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: '🔥 High Expressiveness',
    icon: '⚡',
    description: (
      <>
        Create complex rules with nested logic and layered conditions.
        Easily model edge-case logic like loops, short-circuiting,
        or nested evaluation for complex business requirements.
      </>
    ),
  },
  {
    title: '🧩 Reusable & Modular',
    icon: '🔧',
    description: (
      <>
        Functions (fun) enable rules to be broken down into reusable blocks.
        Developers can build ruleset libraries for various application domains.
      </>
    ),
  },
  {
    title: '🚀 Full Flow Control',
    icon: '⚙️',
    description: (
      <>
        Rules go beyond "if this then that" to include loops
        (check fraud in last 10 transactions), break/continue for early exit,
        and complex nested rule evaluation.
      </>
    ),
  },
  {
    title: '🔍 Debuggable & Observable',
    icon: '🐛',
    description: (
      <>
        Since it executes like a program, you can add logging/debug prints
        inside rules for easy observability and troubleshooting.
      </>
    ),
  },
  {
    title: '🎯 Declarative + Imperative',
    icon: '📝',
    description: (
      <>
        Flexibility to write rules declaratively (if this then that) or
        imperatively (run this logic if...), perfect for security,
        anti-fraud, and automation domains.
      </>
    ),
  },
  {
    title: '🔄 Runtime Programmable',
    icon: '💾',
    description: (
      <>
        Store rules in database, load and execute dynamically.
        Modify rules without recompiling the main system for
        maximum adaptability.
      </>
    ),
  },
];

function Feature({ icon, title, description }) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <div className={styles.featureIcon}>{icon}</div>
      </div>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
