import { Suspense } from 'react';

import { LoadingStateWithBorder } from '@/components/layout/loading';

import ConfigurePage from '../configure-page';

export default function ConfigureRoutePage() {
  return (
    <Suspense
      fallback={
        <div className='w-full px-1 py-6'>
          <LoadingStateWithBorder
            title='加载网站配置'
            description='正在读取网站、域名和源站配置...'
          />
        </div>
      }
    >
      <ConfigurePage />
    </Suspense>
  );
}
