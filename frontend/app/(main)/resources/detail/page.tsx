import { Suspense } from 'react';

import { LoadingStateWithBorder } from '@/components/layout/loading';

import ResourceDetailPage from './detail-page';

export default function ResourceDetailRoutePage() {
  return (
    <Suspense
      fallback={
        <div className='w-full px-1 py-6'>
          <LoadingStateWithBorder
            title='加载站点控制台'
            description='正在读取站点信息...'
          />
        </div>
      }
    >
      <ResourceDetailPage />
    </Suspense>
  );
}
