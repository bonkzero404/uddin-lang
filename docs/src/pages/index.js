import React from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageFeatures from '@site/src/components/HomepageFeatures';

import styles from './index.module.css';

function WhyUddinLang() {
    return (
        <section className="padding-vert--xl">
            <div className="container">
                <div className="row">
                    <div className="col col--6">
                        <h2>🔥 Why Choose UDDIN-LANG?</h2>
                        <p>
                            Unlike traditional rule engines that limit you to simple if-then logic,
                            UDDIN-LANG provides the full power of a programming language while maintaining
                            the simplicity and flexibility needed for business rules.
                        </p>
                        <ul>
                            <li><strong>Beyond Simple Rules:</strong> Support for loops, functions, and complex logic</li>
                            <li><strong>Runtime Flexibility:</strong> Modify rules without system restarts</li>
                            <li><strong>Developer Friendly:</strong> Familiar syntax with powerful debugging</li>
                            <li><strong>Enterprise Ready:</strong> Built for scale and performance</li>
                        </ul>
                    </div>
                    <div className="col col--6">
                        <div className="padding--md" style={{ backgroundColor: 'var(--ifm-color-emphasis-100)', borderRadius: '8px' }}>
                            <h3>💡 Perfect For:</h3>
                            <div className="row">
                                <div className="col col--6">
                                    <ul style={{ listStyle: 'none', padding: 0 }}>
                                        <li>🛡️ Security & Anti-fraud</li>
                                        <li>🤖 Business Automation</li>
                                        <li>📊 Decision Support</li>
                                        <li>🔄 Workflow Management</li>
                                    </ul>
                                </div>
                                <div className="col col--6">
                                    <ul style={{ listStyle: 'none', padding: 0 }}>
                                        <li>⚖️ Compliance Rules</li>
                                        <li>💰 Financial Logic</li>
                                        <li>🎯 Recommendation Systems</li>
                                        <li>📋 Data Validation</li>
                                    </ul>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </section>
    );
}

function CodeExample() {
    return (
        <section className="padding-vert--xl" style={{ backgroundColor: 'var(--ifm-background-color)' }}>
            <div className="container">
                <div className="row">
                    <div className="col col--12">
                        <h2 className="text--center margin-bottom--lg">📝 See UDDIN-LANG in Action</h2>
                        <div className="row">
                            <div className="col col--6">
                                <h3>🚨 Fraud Detection Rule</h3>
                                <pre style={{ backgroundColor: 'var(--ifm-color-emphasis-100)', padding: '1rem', borderRadius: '8px', fontSize: '0.9rem' }}>
                                    {`// Check for suspicious transactions
fun checkFraud(transaction) {
    if (transaction.amount > 10000 &&
        transaction.time < "09:00" ||
        transaction.time > "23:00") {
        return "HIGH_RISK";
    }

    if (transaction.location != user.usual_location) {
        return "MEDIUM_RISK";
    }

    return "LOW_RISK";
}`}
                                </pre>
                            </div>
                            <div className="col col--6">
                                <h3>⚙️ Business Validation Rule</h3>
                                <pre style={{ backgroundColor: 'var(--ifm-color-emphasis-100)', padding: '1rem', borderRadius: '8px', fontSize: '0.9rem' }}>
                                    {`// Validate customer eligibility
fun validateCustomer(customer) {
    if (customer.age < 18) {
        return false;
    }

    for (account in customer.accounts) {
        if (account.status == "BLOCKED") {
            return false;
        }
    }

    return customer.credit_score > 650;
}`}
                                </pre>
                            </div>
                        </div>
                        <div className="text--center margin-top--lg">
                            <Link
                                className="button button--primary button--lg"
                                to="/docs/examples">
                                Explore More Examples →
                            </Link>
                        </div>
                    </div>
                </div>
            </div>
        </section>
    );
}

function HomepageHeader() {
    const { siteConfig } = useDocusaurusContext();
    return (
        <header className={clsx('hero hero--primary', styles.heroBanner)}>
            <div className="container">
                <div className="row">
                    <div className="col col--8 col--offset-2">
                        <h1 className="hero__title">{siteConfig.title}</h1>
                        <p className="hero__subtitle">{siteConfig.tagline}</p>
                        <div className="margin-top--lg margin-bottom--lg">
                            <p className="hero__description">
                                A powerful rule engine that combines programming language expressiveness with business rule flexibility.
                                Perfect for complex decision support systems, fraud detection, automation, and enterprise workflows.
                            </p>
                        </div>
                        <div className={styles.buttons}>
                            <Link
                                className="button button--secondary button--lg margin-right--md"
                                to="/docs/intro">
                                Get Started - 5min ⏱️
                            </Link>
                            <Link
                                className="button button--outline button--secondary button--lg"
                                to="/docs/examples">
                                View Examples 📋
                            </Link>
                        </div>
                        <div className="margin-top--lg">
                            <p className="hero__stats">
                                🚀 Runtime Programmable • 🧩 Modular Design • 🔍 Debuggable • ⚡ High Performance
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </header>
    );
}

export default function Home() {
    const { siteConfig } = useDocusaurusContext();
    return (
        <Layout
            title={`${siteConfig.title} - Unified Dynamic Decision Interpreter Notation`}
            description="UDDIN-LANG (Unified Dynamic Decision Interpreter Notation) is a specialized Flexible Rule Logic Platform that resembles a programming language, offering high expressiveness, full flow control, and runtime programmable capabilities for complex business decision support systems.">
            <HomepageHeader />
            <main>
                <HomepageFeatures />
                <WhyUddinLang />
                <CodeExample />
            </main>
        </Layout>
    );
}
