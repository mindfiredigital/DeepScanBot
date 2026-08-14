import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    {
      type: 'category',
      label: 'Getting Started',
      items: ['introduction', 'installation'],
    },
    {
      type: 'category',
      label: 'Usage Guide',
      items: [
        {
          type: 'doc',
          id: 'guide/usage',
          label: 'Usage Guide',
        },
        {
          type: 'doc',
          id: 'guide/features',
          label: 'Features',
        },
      ],
    },
    {
      type: 'category',
      label: 'CLI Reference',
      items: ['cli-reference'],
    },
    {
      type: 'category',
      label: 'Output & JSON',
      items: ['output-json'],
    },
    {
      type: 'category',
      label: 'Automation & CI/CD',
      items: ['automation-ci'],
    },
    {
      type: 'category',
      label: 'Troubleshooting',
      items: ['troubleshooting'],
    },
    {
      type: 'category',
      label: 'Contributing',
      items: [
        'contribution-guide/how-to-contribute',
        'contribution-guide/code-of-conduct',
      ],
    },
  ],
};

export default sidebars;
