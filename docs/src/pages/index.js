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
                <div className="row" style={{ alignItems: 'stretch' }}>
                    <div className="col col--6">
                        <div style={{ height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
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
                    </div>
                    <div className="col col--6">
                        <div className="padding--md" style={{ backgroundColor: 'var(--ifm-color-emphasis-100)', borderRadius: '8px', height: '100%', display: 'flex', flexDirection: 'column' }}>
                            <h3>💡 Perfect For:</h3>
                            <div className="row" style={{ flex: 1 }}>
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
        <section className="padding-vert--xl versatility-section">
            <div className="container">
                <div className="row">
                    <div className="col col--12">
                        <h2 className="text--center margin-bottom--lg">🚀 Beyond Rules: UDDIN-LANG Versatility</h2>
                        <div className="text--center margin-bottom--lg">
                            <p className="subtitle" style={{ fontSize: '1.1rem', maxWidth: '800px', margin: '0 auto' }}>
                                UDDIN-LANG isn't just a rule engine - it's a complete foundation for building intelligent platforms and learning programming fundamentals.
                            </p>
                        </div>
                        <div className="row">
                            <div className="col col--6">
                                <div className="card shadow--md" style={{ height: '100%', padding: '1.5rem' }}>
                                    <h3>🏗️ Build Rule Platforms</h3>
                                    <p>
                                        Create sophisticated visual rule designers and decision management systems.
                                        UDDIN-LANG serves as the powerful interpreter foundation for platforms like React Flow,
                                        enabling business users to design complex logic visually while maintaining
                                        programming language precision under the hood.
                                    </p>
                                    <ul style={{ marginTop: '1rem' }}>
                                        <li><strong>Visual Rule Designers</strong> - Drag-and-drop rule creation interfaces</li>
                                        <li><strong>Decision Support Systems</strong> - Enterprise-grade rule management</li>
                                        <li><strong>Workflow Automation</strong> - Process rules with complex branching</li>
                                        <li><strong>Compliance Engines</strong> - Regulatory rules with audit trails</li>
                                    </ul>
                                </div>
                            </div>
                            <div className="col col--6">
                                <div className="card shadow--md" style={{ height: '100%', padding: '1.5rem' }}>
                                    <h3>🎓 Learn Logic-Focused Programming</h3>
                                    <p>
                                        Perfect for educational environments and developers who want to master
                                        logical thinking. UDDIN-LANG's clean syntax and focus on decision logic
                                        makes it an ideal language for learning programming fundamentals,
                                        especially conditional logic, functions, and algorithmic thinking.
                                    </p>
                                    <ul style={{ marginTop: '1rem' }}>
                                        <li><strong>Clean Syntax</strong> - Easy to read and understand for beginners</li>
                                        <li><strong>Logic First</strong> - Focus on decision-making and algorithms</li>
                                        <li><strong>Educational Tool</strong> - Perfect for teaching programming concepts</li>
                                        <li><strong>Debugging Friendly</strong> - Clear execution flow for learning</li>
                                    </ul>
                                </div>
                            </div>
                        </div>
                        <div className="row margin-top--lg">
                            <div className="col col--12">
                                <div className="card shadow--md" style={{ padding: '1.5rem', background: 'var(--ifm-color-emphasis-100)' }}>
                                    <h3 className="text--center">🌟 Real-World Applications</h3>
                                    <div className="row">
                                        <div className="col col--4">
                                            <h4>🛡️ Security & Fraud Detection</h4>
                                            <p>Multi-factor authentication rules, transaction monitoring, and behavioral analysis systems.</p>
                                        </div>
                                        <div className="col col--4">
                                            <h4>🤖 Business Automation</h4>
                                            <p>Approval workflows, document processing, and intelligent routing systems.</p>
                                        </div>
                                        <div className="col col--4">
                                            <h4>📊 Decision Support</h4>
                                            <p>Recommendation engines, risk assessment tools, and dynamic pricing systems.</p>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div className="text--center margin-top--lg">
                            <Link
                                className="button button--primary button--lg"
                                to="/docs/examples">
                                Explore Implementation Examples →
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
        <header className={clsx('hero', styles.heroBanner)}>
            <div className={styles.floatingDots}>
                <div className={styles.dot}></div>
                <div className={styles.dot}></div>
                <div className={styles.dot}></div>
                <div className={styles.dot}></div>
                <div className={styles.dot}></div>
            </div>
            <div className={styles.modernHero}>
                <div className={styles.heroContent}>
                    <h1 className="hero__title">
                        {siteConfig.title}
                    </h1>
                    <p className="hero__subtitle">
                        {siteConfig.tagline}
                    </p>
                    <p className="hero__description">
                        Build intelligent decision systems with the power of a programming language and the simplicity of business rules.
                        Perfect for fraud detection, automation, and complex decision-making systems.
                    </p>
                    <div className={styles.buttons}>
                        <Link
                            className={clsx(styles.heroButton, styles['heroButton--primary'])}
                            to="/docs/intro">
                            Learn 5 Minutes
                        </Link>
                        <Link
                            className={clsx(styles.heroButton, styles['heroButton--secondary'])}
                            to="/docs/examples">
                            View Examples
                        </Link>
                    </div>
                    <div className="hero__stats">
                        ⚡ Runtime Programmable • 🧩 Modular Design • 🔍 Debuggable
                    </div>
                </div>
                <div className={styles.heroVisual}>
                    <div className={styles.codeExample}>
                        <div className={styles.codeHeader}>
                            <div className={styles.codeDots}>
                                <span></span>
                                <span></span>
                                <span></span>
                            </div>
                            <span className={styles.codeTitle}>fraud-detection.din</span>
                        </div>
                        <div className={styles.codeContent}>
                            <div className={styles.lineNumbers}>
                                <div>1</div>
                                <div>2</div>
                                <div>3</div>
                                <div>4</div>
                                <div>5</div>
                                <div>6</div>
                                <div>7</div>
                                <div>8</div>
                                <div>9</div>
                                <div>10</div>
                                <div>11</div>
                                <div>12</div>
                                <div>13</div>
                                <div>14</div>
                            </div>
                            <pre>
                                <span className={styles.comment}>// Smart fraud detection rule</span>{'\n'}
                                <span className={styles.keyword}>fun</span> <span className={styles.function}>detectFraud</span>(<span className={styles.parameter}>transaction</span>):{'\n'}
                                {'    '}<span className={styles.keyword}>if</span> (transaction.<span className={styles.property}>amount</span> <span className={styles.operator}>></span> <span className={styles.number}>10000</span> <span className={styles.keyword}>and</span>{'\n'}
                                {'        '}transaction.<span className={styles.property}>time</span> <span className={styles.operator}>&lt;</span> <span className={styles.string}>"09:00"</span>) <span className={styles.keyword}>then</span>:{'\n'}
                                {'        '}<span className={styles.keyword}>return</span> <span className={styles.string}>"HIGH_RISK"</span>{'\n'}
                                {'    '}<span className={styles.keyword}>end</span>{'\n\n'}
                                {'    '}<span className={styles.keyword}>if</span> (transaction.<span className={styles.property}>location</span> <span className={styles.operator}>!=</span>{'\n'}
                                {'        '}user.<span className={styles.property}>usual_location</span>) <span className={styles.keyword}>then</span>:{'\n'}
                                {'        '}<span className={styles.keyword}>return</span> <span className={styles.string}>"MEDIUM_RISK"</span>{'\n'}
                                {'    '}<span className={styles.keyword}>end</span>{'\n\n'}
                                {'    '}<span className={styles.keyword}>return</span> <span className={styles.string}>"LOW_RISK"</span>{'\n'}
                                <span className={styles.keyword}>end</span>
                            </pre>
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
