/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a "Next" and "Previous" button
 - automatically add a "Edit this page" link

 {tutorialSidebar: ["intro", "hello"]} -> shorthand for [{type: "doc", id: "intro"}, {type: "doc", id: "hello"}]
 */

// @ts-check
const sidebars = {
    // By default, Docusaurus generates a sidebar from the docs folder structure
    tutorialSidebar: [
        'intro',
        {
            type: 'category',
            label: 'Getting Started',
            collapsed: false,
            items: [
                'getting-started/installation',
                'getting-started/quick-start',
                'getting-started/your-first-program',
            ],
        },
        {
            type: 'category',
            label: 'Tutorial - Basics',
            collapsed: false,
            items: [
                'tutorial/basics/introduction',
                'tutorial/basics/basic-syntax',
                'tutorial/basics/variables-and-data-types',
                'tutorial/basics/operators',
                'tutorial/basics/control-flow',
                'tutorial/basics/functions',
                'tutorial/basics/arrays-and-objects',
                'tutorial/basics/modules',
            ],
        },
        {
            type: 'category',
            label: 'Advanced Features',
            collapsed: false,
            items: [
                'tutorial/advanced/built-in-functions',
                'tutorial/advanced/math-functions',
                'tutorial/advanced/memoization',
                'tutorial/advanced/networking',
                'tutorial/advanced/regex-and-datetime',
                'tutorial/advanced/xml-and-knowledge-systems',
                'tutorial/advanced/error-handling-and-debugging',
                'tutorial/advanced/development-tools',
            ],
        }
    ],

    // Reference sidebar
    referenceSidebar: [
        'reference/index',
        'reference/syntax',
        'reference/builtin-functions',
        'reference/memory-optimization',
        'reference/library',
    ],

    // But you can create a sidebar manually
    /*
    tutorialSidebar: [
      'intro',
      'hello',
      {
        type: 'category',
        label: 'Tutorial',
        items: ['tutorial-basics', 'tutorial-advanced'],
      },
    ],
     */
};

module.exports = sidebars;
