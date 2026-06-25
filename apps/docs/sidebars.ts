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
      label: 'Guides',
      items: ['usage', 'features'],
    },
    {
      type: 'category',
      label: 'Architecture',
      items: ['architecture'],
    },
    {
      type: 'category',
      label: 'Development',
      items: ['development-tools', 'contributing'],
    },
  ],
};

export default sidebars;