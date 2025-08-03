import styles from './styles.module.css';

export default function RuleEngineSection() {
  return (
    <section className={styles.ruleEngineSection}>
      <div className="container">
        <div className="row">
          <div className="col col--12">
            <div className="text--center margin-bottom--lg">
              <h2>Why UDDIN-LANG as a Rule Engine?</h2>
              <p className="hero__subtitle">
                UDDIN-LANG is not designed to be a primary programming or scripting language,
                but rather as a supporting tool for projects and applications to help make
                informed decisions in your business rules.
              </p>
            </div>
          </div>
        </div>

        <div className="row">
          <div className="col col--6">
            <div className={styles.featureBox}>
              <h3>🎯 High Expressiveness</h3>
              <p>
                Create complex rules with nested logic and conditional layers.
                Easily model edge-case logic with loops, short-circuiting, and nested evaluation.
              </p>
            </div>

            <div className={styles.featureBox}>
              <h3>🔧 Reusable & Modular</h3>
              <p>
                Functions (fun) enable rules to be broken into reusable blocks.
                Developers can build ruleset libraries for better maintainability.
              </p>
            </div>

            <div className={styles.featureBox}>
              <h3>🔄 Full Flow Control</h3>
              <p>
                Rules aren't just "if this then that" - they can include loops
                (e.g., fraud checking in last 10 transactions), break/continue for early exit,
                and nested rule evaluation.
              </p>
            </div>
          </div>

          <div className="col col--6">
            <div className={styles.featureBox}>
              <h3>🐛 Debuggable</h3>
              <p>
                Since it runs like a program, you can add logging/debug prints
                within rules for better observability and troubleshooting.
              </p>
            </div>

            <div className={styles.featureBox}>
              <h3>⚡ Declarative + Imperative</h3>
              <p>
                Choose between declarative (if-this-then-that) or imperative
                (run this logic if...) styles based on your needs. Perfect for security,
                anti-fraud, and automation domains.
              </p>
            </div>

            <div className={styles.featureBox}>
              <h3>🚀 Runtime Programmable</h3>
              <p>
                Modify rules without recompiling your main system.
                Deploy rule changes dynamically for rapid business adaptation.
              </p>
            </div>
          </div>
        </div>

        <div className="row margin-top--lg">
          <div className="col col--12">
            <div className="text--center">
              <p className={styles.conclusion}>
                <strong>UDDIN-LANG goes beyond traditional rule engines</strong> - it's a flexible
                rule logic platform similar to Lua or Starlark, but customizable to your domain.
                It combines full programming language features with rule engine capabilities,
                making it ideal for complex business decision support systems.
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
