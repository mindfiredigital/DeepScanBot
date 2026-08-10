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
      items: ['guide/usage', 'guide/features'],
    },
    {
      type: 'category',
      label: 'CLI Reference',
      items: [
        {
          type: 'doc',
          id: 'guide/usage',
          label: 'Commands & Options',
        },
      ],
    },
    {
      type: 'category',
      label: 'Output & JSON',
      items: [
        {
          type: 'doc',
          id: 'guide/usage',
          label: 'Output Formats',
        },
      ],
    },
    {
      type: 'category',
      label: 'Automation & CI/CD',
      items: [
        {
          type: 'doc',
          id: 'guide/usage',
          label: 'Automation Guide',
        },
      ],
    },
    {
      type: 'category',
      label: 'Troubleshooting',
      items: [
        {
          type: 'doc',
          id: 'guide/usage',
          label: 'Common Issues',
        },
      ],
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
