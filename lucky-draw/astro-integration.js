// src/pages/index.astro
---
import Layout from '../layouts/Layout.astro';
import LuckyDraw from '../components/LuckyDraw';
---

<Layout title="Lucky Draw System">
  <main>
    <LuckyDraw client:load />
  </main>
</Layout>

// src/layouts/Layout.astro
---
interface Props {
  title: string;
}

const { title } = Astro.props;
---

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width" />
    <title>{title}</title>
  </head>
  <body>
    <slot />
  </body>
</html>
