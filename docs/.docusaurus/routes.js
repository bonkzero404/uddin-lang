import React from 'react';
import ComponentCreator from '@docusaurus/ComponentCreator';

export default [
  {
    path: '/uddin-lang/docs',
    component: ComponentCreator('/uddin-lang/docs', 'a53'),
    routes: [
      {
        path: '/uddin-lang/docs/contributing/syntax-highlighting',
        component: ComponentCreator('/uddin-lang/docs/contributing/syntax-highlighting', 'd6c'),
        exact: true
      },
      {
        path: '/uddin-lang/docs/examples/',
        component: ComponentCreator('/uddin-lang/docs/examples/', '0cf'),
        exact: true
      },
      {
        path: '/uddin-lang/docs/getting-started/installation',
        component: ComponentCreator('/uddin-lang/docs/getting-started/installation', 'efe'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/getting-started/quick-start',
        component: ComponentCreator('/uddin-lang/docs/getting-started/quick-start', '93f'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/getting-started/your-first-program',
        component: ComponentCreator('/uddin-lang/docs/getting-started/your-first-program', '537'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/intro',
        component: ComponentCreator('/uddin-lang/docs/intro', '37f'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/reference/',
        component: ComponentCreator('/uddin-lang/docs/reference/', 'ebf'),
        exact: true,
        sidebar: "referenceSidebar"
      },
      {
        path: '/uddin-lang/docs/reference/builtin-functions',
        component: ComponentCreator('/uddin-lang/docs/reference/builtin-functions', 'd20'),
        exact: true,
        sidebar: "referenceSidebar"
      },
      {
        path: '/uddin-lang/docs/reference/syntax',
        component: ComponentCreator('/uddin-lang/docs/reference/syntax', '35c'),
        exact: true,
        sidebar: "referenceSidebar"
      },
      {
        path: '/uddin-lang/docs/tutorial/advanced/best-practices',
        component: ComponentCreator('/uddin-lang/docs/tutorial/advanced/best-practices', 'b7d'),
        exact: true
      },
      {
        path: '/uddin-lang/docs/tutorial/advanced/built-in-functions',
        component: ComponentCreator('/uddin-lang/docs/tutorial/advanced/built-in-functions', '5ad'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/tutorial/advanced/development-tools',
        component: ComponentCreator('/uddin-lang/docs/tutorial/advanced/development-tools', '99b'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/tutorial/advanced/error-handling-and-debugging',
        component: ComponentCreator('/uddin-lang/docs/tutorial/advanced/error-handling-and-debugging', '09b'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/tutorial/advanced/math-functions',
        component: ComponentCreator('/uddin-lang/docs/tutorial/advanced/math-functions', 'bb4'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/tutorial/advanced/networking',
        component: ComponentCreator('/uddin-lang/docs/tutorial/advanced/networking', 'fb6'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/tutorial/advanced/regex-and-datetime',
        component: ComponentCreator('/uddin-lang/docs/tutorial/advanced/regex-and-datetime', '8c8'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/tutorial/advanced/xml-and-knowledge-systems',
        component: ComponentCreator('/uddin-lang/docs/tutorial/advanced/xml-and-knowledge-systems', '858'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/tutorial/basics/arrays-and-objects',
        component: ComponentCreator('/uddin-lang/docs/tutorial/basics/arrays-and-objects', 'd29'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/tutorial/basics/basic-syntax',
        component: ComponentCreator('/uddin-lang/docs/tutorial/basics/basic-syntax', '5ff'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/tutorial/basics/best-practices',
        component: ComponentCreator('/uddin-lang/docs/tutorial/basics/best-practices', '408'),
        exact: true
      },
      {
        path: '/uddin-lang/docs/tutorial/basics/control-flow',
        component: ComponentCreator('/uddin-lang/docs/tutorial/basics/control-flow', '8d4'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/tutorial/basics/error-handling',
        component: ComponentCreator('/uddin-lang/docs/tutorial/basics/error-handling', 'af6'),
        exact: true
      },
      {
        path: '/uddin-lang/docs/tutorial/basics/functions',
        component: ComponentCreator('/uddin-lang/docs/tutorial/basics/functions', '532'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/tutorial/basics/introduction',
        component: ComponentCreator('/uddin-lang/docs/tutorial/basics/introduction', 'd94'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/tutorial/basics/modules',
        component: ComponentCreator('/uddin-lang/docs/tutorial/basics/modules', 'ff1'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/tutorial/basics/operators',
        component: ComponentCreator('/uddin-lang/docs/tutorial/basics/operators', '9c4'),
        exact: true,
        sidebar: "tutorialSidebar"
      },
      {
        path: '/uddin-lang/docs/tutorial/basics/variables-and-data-types',
        component: ComponentCreator('/uddin-lang/docs/tutorial/basics/variables-and-data-types', '847'),
        exact: true,
        sidebar: "tutorialSidebar"
      }
    ]
  },
  {
    path: '/uddin-lang/',
    component: ComponentCreator('/uddin-lang/', '1a8'),
    exact: true
  },
  {
    path: '*',
    component: ComponentCreator('*'),
  },
];
