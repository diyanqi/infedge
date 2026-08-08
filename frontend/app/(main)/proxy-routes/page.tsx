import { Suspense } from 'react';

import { LoadingStateWithBorder } from '@/components/layout/loading';

import { ProxyRoutesPageClient } from './page-client';

function ProxyRoutesPageFallback() {
  return (
    <div className='w-full px-1 py-6'>
      <LoadingStateWithBorder
        title='加载代理规则'
        description='正在读取网站和回源配置...'
      />
    </div>
  );
}

export default function ProxyRoutesPage() {
  return (
    <Suspense fallback={<ProxyRoutesPageFallback />}>
      <ProxyRoutesPageClient />
    </Suspense>
  );
}
