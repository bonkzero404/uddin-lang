// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
const config = {
    title: 'UDDIN-LANG',
    tagline: 'Unified Dynamic Decision Interpreter Notation - Flexible Rule Logic Platform with Programming Language Expressiveness',
    favicon: 'img/favicon.ico',

    // Set the production url of your site here
    url: 'https://bonkzero404.github.io',
    // Set the /<baseUrl>/ pathname under which your site is served
    // For GitHub pages deployment, it is often '/<repoName>/'
    baseUrl: '/uddin-lang/',

    // GitHub pages deployment config.
    // If you aren't using GitHub pages, you don't need these.
    organizationName: 'bonkzero404', // Usually your GitHub org/user name.
    projectName: 'uddin-lang', // Usually your repo name.

    onBrokenLinks: 'warn',
    onBrokenMarkdownLinks: 'warn',

    // Even if you don't use internalization, you can use this field to set useful
    // metadata like html lang. For example, if your site is Chinese, you may want
    // to replace "en" with "zh-Hans".
    i18n: {
        defaultLocale: 'en',
        locales: ['en'],
    },

    presets: [
        [
            'classic',
            /** @type {import('@docusaurus/preset-classic').Options} */
            ({
                docs: {
                    sidebarPath: require.resolve('./sidebars.js'),
                    // Please change this to your repo.
                    // Remove this to remove the "edit this page" links.
                    editUrl:
                        'https://github.com/bonkzero404/uddin-lang/tree/main/docs/',
                },
                theme: {
                    customCss: require.resolve('./src/css/custom.css'),
                },
            }),
        ],
    ],

    themeConfig:
        /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
        ({
            // Replace with your project's social card
            image: 'img/uddin-social-card.jpg',
            navbar: {
                title: 'Uddin Lang',
                logo: {
                    alt: 'Uddin Lang Logo',
                    src: 'img/logo.svg',
                },
                items: [
                    {
                        type: 'docSidebar',
                        sidebarId: 'tutorialSidebar',
                        position: 'left',
                        label: 'Tutorial',
                    },
                    {
                        type: 'docSidebar',
                        sidebarId: 'referenceSidebar',
                        position: 'left',
                        label: 'Reference',
                    },
                    {
                        to: '/docs/examples',
                        label: 'Examples',
                        position: 'left',
                    },
                    {
                        href: 'https://github.com/bonkzero404/uddin-lang',
                        label: 'GitHub',
                        position: 'right',
                    },
                ],
            },
            colorMode: {
                defaultMode: 'light',
                disableSwitch: false,
                respectPrefersColorScheme: true,
            },
            footer: {
                style: 'light',
                copyright: `© ${new Date().getFullYear()} UDDIN-LANG. Built with ❤️ • Licensed under <a href="https://github.com/bonkzero404/uddin-lang/blob/main/LICENSE" target="_blank" rel="noopener noreferrer">MIT</a>`,
            },
            prism: {
                theme: lightCodeTheme,
                darkTheme: darkCodeTheme,
                additionalLanguages: ['go', 'javascript', 'json', 'uddin'],
            },
        }),
};

module.exports = config;
