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
      label: 'Guide',
      items: ['guide/usage', 'guide/features'],
    },
    {
      type: 'category',
      label: 'Contribution Guide',
      items: ['contribution-guide/how-to-contribute', 'contribution-guide/code-of-conduct'],
    },
  ],
};

export default sidebars;
