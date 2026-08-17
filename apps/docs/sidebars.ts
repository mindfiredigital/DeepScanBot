import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    {
      type: 'doc',
      id: 'introduction',
      label: 'Getting Started',
    },
    {
      type: 'doc',
      id: 'installation',
      label: 'Installation',
    },
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
    {
      type: 'doc',
      id: 'cli-reference',
      label: 'CLI Reference',
    },
    {
      type: 'doc',
      id: 'output-json',
      label: 'Output & JSON',
    },
    {
      type: 'doc',
      id: 'automation-ci',
      label: 'Automation & CI/CD',
    },
    {
      type: 'doc',
      id: 'troubleshooting',
      label: 'Troubleshooting',
    },
    {
      type: 'doc',
      id: 'security',
      label: 'Security',
    },
    {
      type: 'doc',
      id: 'contributing',
      label: 'Contributing',
    },
  ],
};

export default sidebars;
