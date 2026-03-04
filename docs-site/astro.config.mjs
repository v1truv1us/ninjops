import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  base: '/ninjops',
  output: 'static',
  integrations: [
    starlight({
      title: 'Ninjops Documentation',
      description: 'Production-ready Go CLI for Invoice Ninja quote/invoice management',
      sidebar: [
        {
          label: 'Getting Started',
          items: [
            { slug: 'getting-started/introduction' },
            { slug: 'getting-started/installation' },
            { slug: 'getting-started/quick-start' },
            { slug: 'getting-started/configuration' },
          ],
        },
        {
          label: 'Guides',
          items: [
            { slug: 'guides/creating-quotes' },
            { slug: 'guides/ai-assistance' },
            { slug: 'guides/invoice-ninja-sync' },
            { slug: 'guides/http-api' },
          ],
        },
        {
          label: 'Reference',
          items: [
            { slug: 'reference/commands' },
            { slug: 'reference/quotespec-schema' },
            { slug: 'reference/api-contract' },
            { slug: 'reference/providers' },
          ],
        },
        {
          label: 'Examples',
          items: [
            { slug: 'examples/church-website' },
            { slug: 'examples/business-site' },
            { slug: 'examples/web-application' },
          ],
        },
        {
          label: 'Architecture',
          items: [
            { slug: 'architecture/overview' },
            { slug: 'architecture/decisions' },
          ],
        },
      ],
      social: {
        github: 'https://github.com/v1truv1us/ninjops',
      },
      editLink: {
        baseUrl: 'https://github.com/v1truv1us/ninjops/edit/main/docs-site/',
      },
      lastUpdated: true,
      pagination: true,
      customCss: [
        './src/styles/custom.css',
      ],
    }),
  ],
});
